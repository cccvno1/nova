package handler

import (
	"github.com/cccvno1/nova/pkg/response"
	"github.com/labstack/echo/v4"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(c echo.Context) error {
	return response.Success(c, echo.Map{
		"status":  "ok",
		"service": "nova",
	})
}

type PingRequest struct {
	Message string `json:"message" validate:"required,min=1,max=100"`
}

type PingResponse struct {
	Echo string `json:"echo"`
}

func (h *HealthHandler) Ping(c echo.Context) error {
	req := new(PingRequest)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	return response.Success(c, PingResponse{
		Echo: req.Message,
	})
}
