package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

func TestLoginCreatesHashedSession(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct horse battery staple"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	users := &userRepositoryStub{credentials: entity.Credentials{User: entity.User{ID: 42, IsActive: true}, PasswordHash: string(passwordHash)}}
	sessions := &sessionRepositoryStub{}
	service := New(users, sessions, 2*time.Hour)
	fixedNow := time.Date(2026, time.September, 30, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return fixedNow }

	session, err := service.Login(context.Background(), " Teacher@Example.com ", "correct horse battery staple")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if users.requestedEmail != "teacher@example.com" {
		t.Fatalf("normalized email = %q", users.requestedEmail)
	}
	if session.Token == "" || !session.ExpiresAt.Equal(fixedNow.Add(2*time.Hour)) {
		t.Fatalf("session = %#v", session)
	}
	wantHash := sha256.Sum256([]byte(session.Token))
	if sessions.createdUserID != 42 || string(sessions.createdTokenHash) != string(wantHash[:]) {
		t.Fatal("session repository did not receive the user id and SHA-256 token hash")
	}
}

func TestLoginDoesNotRevealCredentialFailure(t *testing.T) {
	validHash, err := bcrypt.GenerateFromPassword([]byte("correct password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		users *userRepositoryStub
	}{
		{name: "unknown email", users: &userRepositoryStub{credentialsError: repo.ErrNotFound}},
		{name: "wrong password", users: &userRepositoryStub{credentials: entity.Credentials{User: entity.User{IsActive: true}, PasswordHash: string(validHash)}}},
		{name: "inactive user", users: &userRepositoryStub{credentials: entity.Credentials{User: entity.User{IsActive: false}, PasswordHash: string(validHash)}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.users, &sessionRepositoryStub{}, time.Hour).Login(context.Background(), "user@example.com", "wrong password")
			if !errors.Is(err, usecase.ErrInvalidCredentials) {
				t.Fatalf("Login() error = %v", err)
			}
		})
	}
}

type userRepositoryStub struct {
	credentials      entity.Credentials
	credentialsError error
	user             entity.User
	userError        error
	requestedEmail   string
}

func (stub *userRepositoryStub) FindCredentialsByEmail(_ context.Context, email string) (entity.Credentials, error) {
	stub.requestedEmail = email
	return stub.credentials, stub.credentialsError
}

func (stub *userRepositoryStub) FindByID(context.Context, int64) (entity.User, error) {
	return stub.user, stub.userError
}

type sessionRepositoryStub struct {
	createdUserID    int64
	createdTokenHash []byte
	createdExpiresAt time.Time
}

func (stub *sessionRepositoryStub) Create(_ context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error {
	stub.createdUserID = userID
	stub.createdTokenHash = append([]byte(nil), tokenHash...)
	stub.createdExpiresAt = expiresAt
	return nil
}

func (stub *sessionRepositoryStub) FindActiveUserID(context.Context, []byte, time.Time) (int64, error) {
	return 0, repo.ErrNotFound
}

func (stub *sessionRepositoryStub) Revoke(context.Context, []byte, time.Time) error { return nil }
