package main

import (
	"akira/internal/config/env"
	"akira/internal/server"
	"akira/internal/usecase/logger"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
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
	// app := chi.NewRouter()

	s := server.NewServer(ctx, "", env.PORT, nil, logger)

	return s.Run()
}
