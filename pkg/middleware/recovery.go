package middleware

import (
	"runtime/debug"

	"log/slog"

	"github.com/cccvno1/nova/pkg/errors"
	"github.com/cccvno1/nova/pkg/logger"
	"github.com/cccvno1/nova/pkg/response"
	"github.com/labstack/echo/v4"
)

func Recovery() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("panic recovered",
						slog.Any("error", r),
						slog.String("stack", string(debug.Stack())),
					)

					err := errors.New(errors.ErrInternalServer, "internal server error")
					response.Error(c, err)
				}
			}()

			return next(c)
		}
	}
}
