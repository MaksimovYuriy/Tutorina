package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/maksimovyuriy/tutorina/backend/internal/config"
	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	userrepo "github.com/maksimovyuriy/tutorina/backend/internal/repo/user"
	"github.com/maksimovyuriy/tutorina/backend/internal/storage/postgres"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
	userusecase "github.com/maksimovyuriy/tutorina/backend/internal/usecase/user"
)

func main() {
	username := flag.String("username", "", "user username")
	roleList := flag.String("roles", "admin", "comma-separated roles: admin")
	flag.Parse()

	password := os.Getenv("BOOTSTRAP_PASSWORD")
	if password == "" {
		log.Fatal("BOOTSTRAP_PASSWORD is required")
	}
	roles, err := parseRoles(*roleList)
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	database, err := postgres.New(cfg.DB)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	service := userusecase.New(userrepo.New(database))
	created, err := service.Create(context.Background(), *username, password, roles)
	if errors.Is(err, usecase.ErrConflict) {
		log.Fatal("a user with this username already exists")
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created user %s with id %d and roles %v\n", created.Username, created.ID, created.Roles)
}

func parseRoles(value string) ([]entity.Role, error) {
	parts := strings.Split(value, ",")
	roles := make([]entity.Role, 0, len(parts))
	for _, part := range parts {
		role := entity.Role(strings.TrimSpace(part))
		if !role.Valid() {
			return nil, fmt.Errorf("unknown role %q", role)
		}
		roles = append(roles, role)
	}
	return roles, nil
}
