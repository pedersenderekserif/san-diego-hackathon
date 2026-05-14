package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/pedersenderekserif/san-diego-hackathon/api/internal/db"
	"github.com/pedersenderekserif/san-diego-hackathon/api/internal/handlers"
	"github.com/pedersenderekserif/san-diego-hackathon/api/internal/router"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	conn, err := db.NewPostgresFromEnv(context.Background())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close()

	aetnaSource := handlers.LoadAetnaSource(context.Background())
	bcbsilSource := handlers.LoadBCBSILSource(context.Background())
	bcbstxSource := handlers.LoadBCBSTXSource(context.Background())

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router.New(conn, aetnaSource, bcbsilSource, bcbstxSource),
	}

	log.Printf("api server listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
