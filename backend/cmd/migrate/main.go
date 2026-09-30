package main

import (
	"log"

	"github.com/maksimovyuriy/tutorina/backend/internal/lib/migrator"
)

func main() {
	if err := migrator.Run(); err != nil {
		log.Fatal(err)
	}
}
