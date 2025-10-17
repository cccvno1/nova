package middleware

import (
	"log/slog"

	"github.com/cccvno1/nova/pkg/errors"
	"github.com/cccvno1/nova/pkg/logger"
	"github.com/cccvno1/nova/pkg/response"
	customValidator "github.com/cccvno1/nova/pkg/validator"
	"github.com/labstack/echo/v4"
)

func ErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr)
			return
		}

		if he, ok := err.(*echo.HTTPError); ok {
			code := errors.ErrInternalServer
			switch he.Code {
			case 400:
				code = errors.ErrBadRequest
			case 401:
				code = errors.ErrUnauthorized
			case 403:
				code = errors.ErrForbidden
			case 404:
				code = errors.ErrNotFound
			case 405:
				code = errors.ErrMethodNotAllowed
			case 409:
				code = errors.ErrConflict
			case 429:
				code = errors.ErrTooManyRequests
			}

			message := code.String()
			if msg, ok := he.Message.(string); ok {
				message = msg
			}

			response.ErrorWithMessage(c, code, message)
			return
		}

		if validationErr := customValidator.FormatValidationError(err); len(validationErr) > 0 {
			appErr := errors.NewWithDetails(errors.ErrInvalidParams, "validation failed", validationErr)
			response.Error(c, appErr)
			return
		}

		logger.Error("unhandled error",
			slog.Any("error", err),
			slog.String("path", c.Request().URL.Path),
		)

		response.ErrorWithCode(c, errors.ErrInternalServer)
	}
}
