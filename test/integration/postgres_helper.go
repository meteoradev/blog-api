package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/meteoradev/BlogApi/test/fixtures"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDatabase struct {
	Container *postgres.PostgresContainer
	ConnStr   string
	DB        *sqlx.DB
}

func StartPostgresContainer(ctx context.Context) (*TestDatabase, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp"),
			wait.ForLog("database system is ready to accept connections").WithStartupTimeout(180*time.Second)))
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	host, err := pgContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}
	port, err := pgContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get port: %w", err)
	}

	// forcing IPv4
	if host == "localhost" {
		host = "127.0.0.1"
	}
	connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &TestDatabase{
		Container: pgContainer,
		ConnStr:   connStr,
		DB:        db,
	}, nil
}

func StartTestPostgres(t *testing.T) *TestDatabase {
	testDB, err := StartPostgresContainer(context.Background())
	if err != nil {
		t.Fatal("failed to start postgres container:", err)
	}

	testcontainers.CleanupContainer(t, testDB.Container)

	if err := fixtures.RunMigrations(testDB.ConnStr); err != nil {
		t.Fatal("failed to run migrations:", err)
	}

	return testDB
}
