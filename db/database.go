package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

// Init initializes the database connection pool.
func Init(databaseURL string) error {
	if databaseURL == "" {
		return fmt.Errorf("database URL is empty")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("unable to parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	Pool = pool
	log.Println("Database connection pool established")
	return nil
}

// GetRefreshToken retrieves the refresh token for a given user ID from the database.
func GetRefreshToken(userID string) (string, error) {
	if Pool == nil {
		return "", fmt.Errorf("database connection pool is not initialized")
	}

	var refreshToken string
	query := `SELECT refresh_token FROM spotify_connect WHERE owner_id = $1`
	err := Pool.QueryRow(context.Background(), query, userID).Scan(&refreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to get refresh token for user %s: %w", userID, err)
	}

	return refreshToken, nil
}
