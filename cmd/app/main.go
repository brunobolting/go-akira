package main

import (
	"akira/internal/config/env"
	"akira/internal/db"
	"akira/internal/server"
	"akira/internal/usecase/logger"
	"akira/internal/usecase/session"
	"akira/internal/usecase/user"
	"akira/internal/web"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()
	err := env.Load()
	if err != nil {
		log.Fatalf("failed to load env: %v", err)
	}
	logger := logger.NewLogger()
	defer logger.Close()
	sqlite, err := db.NewSqliteConnection(db.SqliteConfig{
		Path: env.DATABASE_DSN,
		MaxOpenConns: 25,
		MaxIdleConns: 25,
		ConnMaxLifetime: 5 * time.Minute,
	})
	if err != nil {
		logger.Error(ctx, "failed to connect to db", err, nil)
		return err
	}
	defer sqlite.Close()
	userService, _ := user.Make(ctx, sqlite, logger)
	sessionService, _ := session.Make(ctx, sqlite, logger)
	app := chi.NewRouter()
	web := web.NewHandler(app, userService, sessionService, logger, web.Options{
		AllowedOrigins: []string{"same-origin"},
	})
	s := server.NewServer(ctx, "", env.PORT, web, logger)
	s.RegisterCleanup(func() error {
		return sqlite.Close()
	})
	return s.Run()
}
