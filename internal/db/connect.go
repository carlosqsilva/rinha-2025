package db

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"github.com/carlosqsilva/rinha-2025/internal/config"
	_ "github.com/mattn/go-sqlite3"

	"github.com/pressly/goose/v3"
)

const (
	DBDir  = "./data"
	DBPAth = "./data/rinha.db"
)

func setupDatabase() {
	if err := os.MkdirAll(DBDir, 0o700); err != nil {
		slog.Error("failed to create data directory", "error", err)
		os.Exit(1)
	}

	if err := os.Remove(DBPAth); err != nil && os.IsNotExist(err) {
		slog.Info("sqlite db does not exist", "info", err)
	} else {
		slog.Info("creating new sqlite db")
	}
}

func Connect(cfg *config.Config) *Queries {
	if cfg.InitDB {
		setupDatabase()
	}

	db, err := sql.Open("sqlite3", DBPAth)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	if err = db.PingContext(ctx); err != nil {
		db.Close()
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	pragmas := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA page_size = 4096;",
		"PRAGMA cache_size = -8000;",
		"PRAGMA synchronous = NORMAL;",
	}

	// await conn.execute("PRAGMA journal_mode = WAL")
	//  await conn.execute("PRAGMA synchronous = NORMAL")
	//  await conn.execute("PRAGMA cache_size = 10000")
	//  await conn.execute("PRAGMA temp_store = MEMORY")
	//  await conn.execute("PRAGMA foreign_keys = ON")
	//  await conn.execute("PRAGMA mmap_size = 268435456")

	for _, pragma := range pragmas {
		if _, err = db.ExecContext(ctx, pragma); err != nil {
			slog.Error("Failed to set pragma", pragma, err)
		} else {
			slog.Debug("Set pragma", "pragma", pragma)
		}
	}

	goose.SetBaseFS(FS)

	if err := goose.SetDialect("sqlite3"); err != nil {
		slog.Error("Failed to set dialect", "error", err)
		os.Exit(1)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		slog.Error("Failed to apply migrations", "error", err)
		os.Exit(1)
	}

	return New(db)
}
