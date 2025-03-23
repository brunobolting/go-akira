package main

import (
	"akira/internal/config/env"
	"akira/internal/db"
	"akira/internal/locale"
	"akira/internal/server"
	"akira/internal/usecase/i18n"
	"akira/internal/usecase/logger"
	"akira/internal/usecase/session"
	"akira/internal/usecase/theme"
	"akira/internal/usecase/user"
	"akira/internal/web"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/invopop/ctxi18n"
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
	if err := ctxi18n.LoadWithDefault(locale.Content, "en"); err != nil {
		logger.Error(ctx, "failed to load i18n translations", err, nil)
		return err
	}
	userService, _ := user.Make(ctx, sqlite, logger)
	sessionService, _ := session.Make(ctx, sqlite, logger)
	i18n := i18n.Make(ctx, logger)
	theme := theme.Make(ctx, logger)
	app := chi.NewRouter()
	web := web.NewHandler(app, userService, sessionService, logger, i18n, theme, web.Options{
		AllowedOrigins: []string{"same-origin"},
	})
	s := server.NewServer(ctx, "", env.PORT, web, logger)
	s.RegisterCleanup(func() error {
		return sqlite.Close()
	})
	return s.Run()
}
