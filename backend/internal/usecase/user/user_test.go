package user

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
)

func TestCreateNormalizesUserAndHashesPassword(t *testing.T) {
	repository := &repositoryStub{}
	service := New(repository)
	_, err := service.Create(
		context.Background(),
		" Admin ",
		"correct horse battery staple",
		[]entity.Role{entity.RoleAdmin, entity.RoleAdmin},
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repository.username != "admin" {
		t.Fatalf("username = %q", repository.username)
	}
	if len(repository.roles) != 1 {
		t.Fatalf("roles = %#v", repository.roles)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repository.passwordHash), []byte("correct horse battery staple")); err != nil {
		t.Fatalf("password was not hashed correctly: %v", err)
	}
}

func TestCreateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		roles    []entity.Role
	}{
		{name: "username", username: "invalid username", password: "a sufficiently long password", roles: []entity.Role{entity.RoleAdmin}},
		{name: "password", username: "admin", password: "short", roles: []entity.Role{entity.RoleAdmin}},
		{name: "roles", username: "admin", password: "a sufficiently long password"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(&repositoryStub{}).Create(context.Background(), test.username, test.password, test.roles)
			if !errors.Is(err, usecase.ErrInvalidInput) {
				t.Fatalf("Create() error = %v", err)
			}
		})
	}
}

type repositoryStub struct {
	username     string
	passwordHash string
	roles        []entity.Role
}

func (stub *repositoryStub) Create(_ context.Context, username, passwordHash string, roles []entity.Role) (entity.User, error) {
	stub.username = username
	stub.passwordHash = passwordHash
	stub.roles = append([]entity.Role(nil), roles...)
	return entity.User{ID: 1, Username: username, Roles: roles}, nil
}
