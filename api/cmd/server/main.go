package main

import (
	"log"
	"net/http"
	"os"

	"github.com/pedersenderekserif/san-diego-hackathon/api/internal/router"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router.New(),
	}

	log.Printf("api server listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
