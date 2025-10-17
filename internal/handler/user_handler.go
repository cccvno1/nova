package handler

import (
	"strconv"

	"github.com/cccvno1/nova/internal/service"
	"github.com/cccvno1/nova/pkg/database"
	"github.com/cccvno1/nova/pkg/response"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Create(c echo.Context) error {
	req := new(service.CreateUserRequest)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	user, err := h.userService.Create(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return response.Success(c, user)
}

func (h *UserHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return err
	}

	user, err := h.userService.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return err
	}

	return response.Success(c, user)
}

func (h *UserHandler) List(c echo.Context) error {
	pagination := &database.Pagination{}
	if err := c.Bind(pagination); err != nil {
		return err
	}

	if err := c.Validate(pagination); err != nil {
		return err
	}

	users, err := h.userService.List(c.Request().Context(), pagination)
	if err != nil {
		return err
	}

	return response.Page(c, users, pagination.Total, pagination.Page, pagination.PageSize)
}

func (h *UserHandler) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return err
	}

	req := new(service.UpdateUserRequest)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.userService.Update(c.Request().Context(), uint(id), req); err != nil {
		return err
	}

	return response.Success(c, nil)
}

func (h *UserHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return err
	}

	if err := h.userService.Delete(c.Request().Context(), uint(id)); err != nil {
		return err
	}

	return response.Success(c, nil)
}
