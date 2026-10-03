package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

type UserRepository interface {
	FindCredentialsByUsername(context.Context, string) (entity.Credentials, error)
	FindCredentialsByID(context.Context, int64) (entity.Credentials, error)
	UpdatePassword(context.Context, int64, string) error
	FindByID(context.Context, int64) (entity.User, error)
}

type SessionRepository interface {
	Create(context.Context, int64, []byte, time.Time) error
	FindActiveUserID(context.Context, []byte, time.Time) (int64, error)
	Revoke(context.Context, []byte, time.Time) error
	RevokeAll(context.Context, int64, time.Time) error
}

type Service struct {
	users    UserRepository
	sessions SessionRepository
	ttl      time.Duration
	now      func() time.Time
}

func New(users UserRepository, sessions SessionRepository, ttl time.Duration) *Service {
	return &Service{users: users, sessions: sessions, ttl: ttl, now: time.Now}
}

func (service *Service) Login(ctx context.Context, username, password string) (entity.Session, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	credentials, err := service.users.FindCredentialsByUsername(ctx, username)
	if errors.Is(err, repo.ErrNotFound) {
		return entity.Session{}, usecase.ErrInvalidCredentials
	}
	if err != nil {
		return entity.Session{}, err
	}
	if !credentials.IsActive || bcrypt.CompareHashAndPassword([]byte(credentials.PasswordHash), []byte(password)) != nil {
		return entity.Session{}, usecase.ErrInvalidCredentials
	}

	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return entity.Session{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(randomBytes)
	tokenHash := sha256.Sum256([]byte(token))
	expiresAt := service.now().Add(service.ttl)
	if err := service.sessions.Create(ctx, credentials.ID, tokenHash[:], expiresAt); err != nil {
		return entity.Session{}, err
	}
	return entity.Session{Token: token, ExpiresAt: expiresAt}, nil
}

func (service *Service) Authenticate(ctx context.Context, token string) (entity.User, error) {
	if token == "" {
		return entity.User{}, usecase.ErrUnauthorized
	}
	tokenHash := sha256.Sum256([]byte(token))
	userID, err := service.sessions.FindActiveUserID(ctx, tokenHash[:], service.now())
	if errors.Is(err, repo.ErrNotFound) {
		return entity.User{}, usecase.ErrUnauthorized
	}
	if err != nil {
		return entity.User{}, err
	}
	user, err := service.users.FindByID(ctx, userID)
	if errors.Is(err, repo.ErrNotFound) || (err == nil && !user.IsActive) {
		return entity.User{}, usecase.ErrUnauthorized
	}
	return user, err
}

func (service *Service) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	if userID <= 0 || len(newPassword) < 12 || len(newPassword) > 72 {
		return usecase.ErrInvalidInput
	}
	credentials, err := service.users.FindCredentialsByID(ctx, userID)
	if errors.Is(err, repo.ErrNotFound) {
		return usecase.ErrUnauthorized
	}
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(credentials.PasswordHash), []byte(currentPassword)) != nil {
		return usecase.ErrInvalidCredentials
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := service.users.UpdatePassword(ctx, userID, string(passwordHash)); err != nil {
		return err
	}
	return service.sessions.RevokeAll(ctx, userID, service.now())
}
func (service *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	tokenHash := sha256.Sum256([]byte(token))
	return service.sessions.Revoke(ctx, tokenHash[:], service.now())
}
