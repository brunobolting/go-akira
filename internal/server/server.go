package server

import (
	"akira/internal/entity"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

const (
	GRACEFUL_TIMEOUT = 10 * time.Second
)

type Server struct {
	ctx     context.Context
	logger  entity.Logger
	srv     *http.Server
	cleanup []func() error
}

func NewServer(ctx context.Context, addr, port string, handler http.Handler, logger entity.Logger) *Server {
	return &Server{
		ctx:    ctx,
		logger: logger,
		srv: &http.Server{
			Addr:    net.JoinHostPort(addr, port),
			Handler: handler,
		},
	}
}

func (s *Server) RegisterCleanup(f func() error) {
	s.cleanup = append(s.cleanup, f)
}

func (s *Server) Run() error {
	go func() {
		s.logger.Info(s.ctx, fmt.Sprintf("server running on %s", s.srv.Addr), nil)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error(s.ctx, "error listening and serving", err, nil)
		}
	}()
	<-s.ctx.Done()
	s.logger.Info(s.ctx, "shutdown signal received, initiating graceful shutdown", nil)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), GRACEFUL_TIMEOUT)
	defer cancel()

	for i := len(s.cleanup) - 1; i >= 0; i-- {
		if err := s.cleanup[i](); err != nil {
			s.logger.Error(shutdownCtx, "error during cleanup", err, map[string]any{
				"cleanup_index": i,
			})
		}
	}
	if err := s.srv.Shutdown(shutdownCtx); err != nil {
		s.logger.Error(shutdownCtx, "error while gracefully shutting down http server", err, nil)
		return err
	}
	s.logger.Info(shutdownCtx, "server shutdown completed", nil)
	return nil
}
