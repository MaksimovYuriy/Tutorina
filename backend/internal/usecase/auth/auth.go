package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"time"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

type KeyRepository interface {
	Hash(context.Context) ([]byte, error)
}
type SessionRepository interface {
	Create(context.Context, []byte, []byte, time.Time) error
	FindActive(context.Context, []byte, time.Time) error
	Revoke(context.Context, []byte, time.Time) error
}
type Service struct {
	keys     KeyRepository
	sessions SessionRepository
	ttl      time.Duration
	now      func() time.Time
}

func New(keys KeyRepository, sessions SessionRepository, ttl time.Duration) *Service {
	return &Service{keys, sessions, ttl, time.Now}
}
func (s *Service) Login(ctx context.Context, key string) (entity.Session, error) {
	// A key is exactly 32 random bytes encoded as lowercase hexadecimal.
	decoded, err := hex.DecodeString(key)
	if err != nil || len(decoded) != 32 || hex.EncodeToString(decoded) != key {
		return entity.Session{}, usecase.ErrInvalidCredentials
	}
	stored, err := s.keys.Hash(ctx)
	if errors.Is(err, repo.ErrNotFound) {
		return entity.Session{}, usecase.ErrInvalidCredentials
	}
	if err != nil {
		return entity.Session{}, err
	}
	hash := sha256.Sum256([]byte(key))
	if subtle.ConstantTimeCompare(hash[:], stored) != 1 {
		return entity.Session{}, usecase.ErrInvalidCredentials
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return entity.Session{}, err
	}
	token := hex.EncodeToString(random)
	tokenHash := sha256.Sum256([]byte(token))
	expires := s.now().Add(s.ttl)
	if err := s.sessions.Create(ctx, hash[:], tokenHash[:], expires); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return entity.Session{}, usecase.ErrInvalidCredentials
		}
		return entity.Session{}, err
	}
	return entity.Session{Token: token, ExpiresAt: expires}, nil
}
func (s *Service) Authenticate(ctx context.Context, token string) error {
	if len(token) != 64 {
		return usecase.ErrUnauthorized
	}
	hash := sha256.Sum256([]byte(token))
	err := s.sessions.FindActive(ctx, hash[:], s.now())
	if errors.Is(err, repo.ErrNotFound) {
		return usecase.ErrUnauthorized
	}
	return err
}
func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	hash := sha256.Sum256([]byte(token))
	return s.sessions.Revoke(ctx, hash[:], s.now())
}
