package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPostgresFromEnv(ctx context.Context) (*sql.DB, error) {
	host := os.Getenv("PG_HOST")
	user := os.Getenv("PG_USER")
	password := os.Getenv("PG_PASSWORD")

	if host == "" || user == "" || password == "" {
		return nil, errors.New("PG_HOST, PG_USER, and PG_PASSWORD are required")
	}

	port := os.Getenv("PG_PORT")
	if port == "" {
		port = "5432"
	}

	database := os.Getenv("PG_DATABASE")
	if database == "" {
		database = "postgres"
	}

	sslMode := os.Getenv("PG_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(user),
		url.QueryEscape(password),
		host,
		port,
		database,
		sslMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
