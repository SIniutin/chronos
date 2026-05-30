package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/SIniutin/history-app-backend/internal/platform/migration"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/history-db?sslmode=disable"

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = defaultDatabaseURL
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(fmt.Errorf("open db: %w", err))
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		panic(fmt.Errorf("ping db: %w", err))
	}
	if err := migration.Up(context.Background(), db); err != nil {
		panic(err)
	}
}
