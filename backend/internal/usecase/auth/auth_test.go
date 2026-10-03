package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

const testKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

type keyStub struct {
	hash []byte
	err  error
}

func (k keyStub) Hash(context.Context) ([]byte, error) { return k.hash, k.err }

type sessionStub struct {
	keyHash, tokenHash, revokedHash []byte
	expiresAt                       time.Time
	createErr, findErr              error
	created                         bool
}

func (s *sessionStub) Create(_ context.Context, key, token []byte, expires time.Time) error {
	s.keyHash = key
	s.tokenHash = token
	s.expiresAt = expires
	s.created = true
	return s.createErr
}
func (s *sessionStub) FindActive(context.Context, []byte, time.Time) error { return s.findErr }
func (s *sessionStub) Revoke(_ context.Context, hash []byte, _ time.Time) error {
	s.revokedHash = hash
	return nil
}
func TestKeyLoginCreatesIndependentHashedSession(t *testing.T) {
	hash := sha256.Sum256([]byte(testKey))
	sessions := &sessionStub{}
	service := New(keyStub{hash: hash[:]}, sessions, time.Hour)
	now := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	session, err := service.Login(context.Background(), testKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(session.Token) != 64 || session.Token == testKey || !session.ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Fatalf("unexpected session")
	}
	tokenHash := sha256.Sum256([]byte(session.Token))
	if string(sessions.keyHash) != string(hash[:]) || string(sessions.tokenHash) != string(tokenHash[:]) {
		t.Fatal("repository did not receive hashed credentials")
	}
}
func TestInvalidKeyNeverCreatesSession(t *testing.T) {
	hash := sha256.Sum256([]byte(testKey))
	for _, key := range []string{"", strings.Repeat("a", 63), strings.Repeat("a", 65), strings.Repeat("z", 64), strings.ToUpper(testKey), strings.Repeat("b", 64)} {
		sessions := &sessionStub{}
		_, err := New(keyStub{hash: hash[:]}, sessions, time.Hour).Login(context.Background(), key)
		if !errors.Is(err, usecase.ErrInvalidCredentials) || sessions.created {
			t.Fatalf("invalid key accepted")
		}
	}
}
func TestUnconfiguredOrRotatedKeyRejectsLogin(t *testing.T) {
	hash := sha256.Sum256([]byte(testKey))
	for _, tc := range []struct{ keyErr, sessionErr error }{{repo.ErrNotFound, nil}, {nil, repo.ErrNotFound}} {
		_, err := New(keyStub{hash: hash[:], err: tc.keyErr}, &sessionStub{createErr: tc.sessionErr}, time.Hour).Login(context.Background(), testKey)
		if !errors.Is(err, usecase.ErrInvalidCredentials) {
			t.Fatal(err)
		}
	}
}
func TestSessionExpiryOrRevocationRejectsAuthentication(t *testing.T) {
	service := New(keyStub{}, &sessionStub{findErr: repo.ErrNotFound}, time.Hour)
	for _, token := range []string{"", testKey} {
		if err := service.Authenticate(context.Background(), token); !errors.Is(err, usecase.ErrUnauthorized) {
			t.Fatal(err)
		}
	}
}
func TestLogoutUsesHashedToken(t *testing.T) {
	sessions := &sessionStub{}
	service := New(keyStub{}, sessions, time.Hour)
	if err := service.Logout(context.Background(), testKey); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256([]byte(testKey))
	if string(hash[:]) != string(sessions.revokedHash) {
		t.Fatal("logout used raw token")
	}
}
