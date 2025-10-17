package errors

import "fmt"

type AppError struct {
	Code    Code        `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func New(code Code, message string) *AppError {
	if message == "" {
		message = code.String()
	}
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func NewWithDetails(code Code, message string, details interface{}) *AppError {
	if message == "" {
		message = code.String()
	}
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

func Wrap(code Code, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: code.String(),
		Details: err.Error(),
	}
}

func WrapWithMessage(code Code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: err.Error(),
	}
}
