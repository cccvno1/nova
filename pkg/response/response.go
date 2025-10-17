package response

import (
	"github.com/cccvno1/nova/pkg/database"
	"github.com/cccvno1/nova/pkg/errors"
	"github.com/labstack/echo/v4"
)

type Response struct {
	Code    errors.Code `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageData struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

func Success(c echo.Context, data interface{}) error {
	return c.JSON(200, Response{
		Code:    errors.Success,
		Message: "success",
		Data:    data,
	})
}

func SuccessWithMessage(c echo.Context, message string, data interface{}) error {
	return c.JSON(200, Response{
		Code:    errors.Success,
		Message: message,
		Data:    data,
	})
}

func Page(c echo.Context, list interface{}, total int64, page, size int) error {
	return c.JSON(200, Response{
		Code:    errors.Success,
		Message: "success",
		Data: PageData{
			List:  list,
			Total: total,
			Page:  page,
			Size:  size,
		},
	})
}

func SuccessWithPagination(c echo.Context, list interface{}, pagination *database.Pagination) error {
	return c.JSON(200, Response{
		Code:    errors.Success,
		Message: "success",
		Data: PageData{
			List:  list,
			Total: pagination.Total,
			Page:  pagination.Page,
			Size:  pagination.PageSize,
		},
	})
}

func Error(c echo.Context, err *errors.AppError) error {
	return c.JSON(err.Code.HTTPStatus(), Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    err.Details,
	})
}

func ErrorWithCode(c echo.Context, code errors.Code) error {
	return c.JSON(code.HTTPStatus(), Response{
		Code:    code,
		Message: code.String(),
	})
}

func ErrorWithMessage(c echo.Context, code errors.Code, message string) error {
	return c.JSON(code.HTTPStatus(), Response{
		Code:    code,
		Message: message,
	})
}
