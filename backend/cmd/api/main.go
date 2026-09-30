package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/maksimovyuriy/tutorina/backend/internal/app"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8081/health", nil)
		if err != nil {
			log.Fatal(err)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			log.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			log.Fatal(fmt.Errorf("healthcheck returned %s", response.Status))
		}
		return
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
