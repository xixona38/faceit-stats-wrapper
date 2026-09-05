package main

import (
	"context"
	"faceit_stats_wrapper/internal/repository"
	"faceit_stats_wrapper/internal/repository/postgres"
	"faceit_stats_wrapper/internal/service"
	apphttp "faceit_stats_wrapper/internal/transport/http"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	envPath, err := loadDotEnv()
	if err == nil {
		log.Printf("Loaded environment from %s", envPath)
	} else if !os.IsNotExist(err) {
		log.Fatalf("failed to load .env: %v", err)
	}

	apiKey := os.Getenv("FACEIT_API_KEY")
	if apiKey == "" {
		log.Fatal("FACEIT_API_KEY environment variable is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	pgURL := os.Getenv("PG_URL")
	if pgURL == "" {
		log.Fatal("PG_URL environment variable is not set")
	}

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, pgURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Successfully connected to database")

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	faceitRepo := repository.NewFaceitAPI(httpClient, apiKey)
	dbRepoSet := postgres.NewPostgresRepoSet(pool)
	dbRepoGet := postgres.NewPostgresRepoGet(pool)
	svc := service.NewStatsService(faceitRepo, dbRepoSet, dbRepoGet)
	handler := apphttp.NewHandler(svc)

	mux := http.NewServeMux()
	handler.InitRoutes(mux)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Starting server...")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func loadDotEnv() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	for {
		envPath := filepath.Join(dir, ".env")
		if err := godotenv.Load(envPath); err == nil {
			return envPath, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("load %s: %w", envPath, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
