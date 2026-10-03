package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/maksimovyuriy/tutorina/backend/internal/config"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo/accesskey"
	"github.com/maksimovyuriy/tutorina/backend/internal/storage/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := postgres.New(cfg.DB)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		log.Fatal(err)
	}
	key := hex.EncodeToString(random)
	hash := sha256.Sum256([]byte(key))
	if err := accesskey.New(db).Replace(context.Background(), hash[:]); err != nil {
		log.Fatal(err)
	}
	// The generated key is shown once to the operator; only its SHA-256 hash is stored.
	fmt.Println(key)
}
