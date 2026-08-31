package postgres

import (
	"fmt"
	"time"

	"github.com/meteoradev/BlogApi/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewPostgresDB(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", cfg.PostgresDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	return db, nil
}
