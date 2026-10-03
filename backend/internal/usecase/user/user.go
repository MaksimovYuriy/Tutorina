package user

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

const minimumPasswordLength = 12

var usernamePattern = regexp.MustCompile(`^[a-z0-9_-]{3,64}$`)

type Repository interface {
	Create(context.Context, string, string, []entity.Role) (entity.User, error)
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (service *Service) Create(ctx context.Context, username, password string, roles []entity.Role) (entity.User, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	if !usernamePattern.MatchString(username) {
		return entity.User{}, fmt.Errorf("%w: invalid username", usecase.ErrInvalidInput)
	}
	if len(password) < minimumPasswordLength || len(password) > 72 {
		return entity.User{}, fmt.Errorf("%w: password must contain %d to 72 bytes", usecase.ErrInvalidInput, minimumPasswordLength)
	}
	if len(roles) == 0 {
		return entity.User{}, fmt.Errorf("%w: at least one role is required", usecase.ErrInvalidInput)
	}
	uniqueRoles := make([]entity.Role, 0, len(roles))
	seenRoles := make(map[entity.Role]struct{}, len(roles))
	for _, role := range roles {
		if !role.Valid() {
			return entity.User{}, fmt.Errorf("%w: unknown role %q", usecase.ErrInvalidInput, role)
		}
		if _, exists := seenRoles[role]; !exists {
			seenRoles[role] = struct{}{}
			uniqueRoles = append(uniqueRoles, role)
		}
	}
	roles = uniqueRoles

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return entity.User{}, err
	}
	created, err := service.repository.Create(ctx, username, string(passwordHash), roles)
	if errors.Is(err, repo.ErrConflict) {
		return entity.User{}, usecase.ErrConflict
	}
	return created, err
}
