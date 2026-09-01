package main

import (
	"faceit_stats_wrapper/internal/repository"
	"faceit_stats_wrapper/internal/service"
	apphttp "faceit_stats_wrapper/internal/transport/http"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	apiKey := os.Getenv("FACEIT_API_KEY")
	if apiKey == "" {
		log.Fatal("FACEIT_API_KEY environment variable is not set")
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	repo := repository.NewFaceitAPI(httpClient, apiKey)
	svc := service.NewStatsService(repo)
	handler := apphttp.NewHandler(svc)

	mux := http.NewServeMux()
	handler.InitRoutes(mux)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Starting server...")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
