package user

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

const minimumPasswordLength = 12

type Repository interface {
	Create(context.Context, string, string, []entity.Role) (entity.User, error)
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (service *Service) Create(ctx context.Context, email, password string, roles []entity.Role) (entity.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return entity.User{}, fmt.Errorf("%w: invalid email", usecase.ErrInvalidInput)
	}
	if len(password) < minimumPasswordLength {
		return entity.User{}, fmt.Errorf("%w: password must contain at least %d characters", usecase.ErrInvalidInput, minimumPasswordLength)
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
	created, err := service.repository.Create(ctx, email, string(passwordHash), roles)
	if errors.Is(err, repo.ErrConflict) {
		return entity.User{}, usecase.ErrConflict
	}
	return created, err
}
