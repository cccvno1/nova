package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cccvno1/nova/pkg/config"
	"github.com/cccvno1/nova/pkg/logger"
	"github.com/cccvno1/nova/pkg/middleware"
	customValidator "github.com/cccvno1/nova/pkg/validator"
	"github.com/labstack/echo/v4"
)

type Server struct {
	echo   *echo.Echo
	config *config.Config
}

func New(cfg *config.Config) *Server {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true
	e.Validator = customValidator.New()
	e.HTTPErrorHandler = middleware.ErrorHandler()

	e.Use(middleware.Recovery())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	return &Server{
		echo:   e,
		config: cfg,
	}
}

func (s *Server) Echo() *echo.Echo {
	return s.echo
}

func (s *Server) Start() error {
	addr := s.config.GetServerAddr()

	logger.Info("starting http server",
		slog.String("addr", addr),
		slog.String("mode", s.config.Server.Mode))

	go func() {
		if err := s.echo.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start server", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.echo.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	logger.Info("server exited")
	return nil
}
