// Command api boots the Detective Files backend: config, Postgres (with
// auto-migrations), Redis, embedded case content, and the Gin router.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"detectivefiles/internal/cache"
	"detectivefiles/internal/cases"
	"detectivefiles/internal/config"
	"detectivefiles/internal/handlers"
	"detectivefiles/internal/repository"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	cfg, err := config.Load()
	if err != nil {
		log.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	reg, err := cases.Load()
	if err != nil {
		log.Error("case content failed validation", "err", err)
		os.Exit(1)
	}
	log.Info("cases loaded", "count", len(reg.IDs()), "ids", reg.IDs())

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("postgres connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Error("postgres unreachable", "err", err)
		os.Exit(1)
	}

	if err := runMigrations(cfg.DatabaseURL); err != nil {
		log.Error("migrations failed", "err", err)
		os.Exit(1)
	}

	ch := cache.New(cfg.RedisAddr, log)
	rctx, rcancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer rcancel()
	if err := ch.Ping(rctx); err != nil {
		log.Error("redis unreachable", "err", err)
		os.Exit(1)
	}

	gin.SetMode(gin.ReleaseMode)
	app := &handlers.App{
		Cfg: cfg, Pool: pool, Repo: repository.New(pool),
		Cache: ch, Cases: reg, Log: log,
	}
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: app.NewRouter()}
	log.Info("listening", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server failed", "err", err)
		os.Exit(1)
	}
}

// runMigrations applies backend/migrations via golang-migrate. It looks for
// the migrations dir next to the working directory, then next to the binary.
func runMigrations(databaseURL string) error {
	dir := "migrations"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if exe, err := os.Executable(); err == nil {
			alt := filepath.Join(filepath.Dir(exe), "migrations")
			if _, serr := os.Stat(alt); serr == nil {
				dir = alt
			}
		}
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open sql db: %w", err)
	}
	defer db.Close()
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+abs, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
