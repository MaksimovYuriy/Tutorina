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
		" Teacher@Example.com ",
		"correct horse battery staple",
		[]entity.Role{entity.RoleTeacher, entity.RoleAdmin, entity.RoleTeacher},
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repository.email != "teacher@example.com" {
		t.Fatalf("email = %q", repository.email)
	}
	if len(repository.roles) != 2 {
		t.Fatalf("roles = %#v", repository.roles)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repository.passwordHash), []byte("correct horse battery staple")); err != nil {
		t.Fatalf("password was not hashed correctly: %v", err)
	}
}

func TestCreateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		roles    []entity.Role
	}{
		{name: "email", email: "not-an-email", password: "a sufficiently long password", roles: []entity.Role{entity.RoleTeacher}},
		{name: "password", email: "user@example.com", password: "short", roles: []entity.Role{entity.RoleTeacher}},
		{name: "roles", email: "user@example.com", password: "a sufficiently long password"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(&repositoryStub{}).Create(context.Background(), test.email, test.password, test.roles)
			if !errors.Is(err, usecase.ErrInvalidInput) {
				t.Fatalf("Create() error = %v", err)
			}
		})
	}
}

type repositoryStub struct {
	email        string
	passwordHash string
	roles        []entity.Role
}

func (stub *repositoryStub) Create(_ context.Context, email, passwordHash string, roles []entity.Role) (entity.User, error) {
	stub.email = email
	stub.passwordHash = passwordHash
	stub.roles = append([]entity.Role(nil), roles...)
	return entity.User{ID: 1, Email: email, Roles: roles}, nil
}
