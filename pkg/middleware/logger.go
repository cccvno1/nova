package middleware

import (
	"time"

	"log/slog"

	"github.com/cccvno1/nova/pkg/logger"
	"github.com/labstack/echo/v4"
)

func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			req := c.Request()
			res := c.Response()

			err := next(c)

			latency := time.Since(start)

			logger.Info("request completed",
				slog.String("method", req.Method),
				slog.String("uri", req.RequestURI),
				slog.Int("status", res.Status),
				slog.Duration("latency", latency),
				slog.String("ip", c.RealIP()),
				slog.String("user_agent", req.UserAgent()),
			)

			return err
		}
	}
}
