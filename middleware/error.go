package middleware

import (
	"fmt"
	"net/http"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

type HttpError struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (e *HttpError) Error() string {
	return e.Message
}

func NewValidationError(verr validator.ValidationErrors, trans ut.Translator) *HttpError {
	fields := make(map[string]string)

	for _, field := range verr {
		fields[field.Field()] = field.Translate(trans)
	}

	return &HttpError{
		Code:    http.StatusBadRequest,
		Message: "validation error",
		Fields:  fields,
	}
}

func NewNotFoundError() *HttpError {
	return &HttpError{
		Code:    http.StatusNotFound,
		Message: "Not found",
	}
}

func NewNamedNotFoundError(name string) *HttpError {
	return &HttpError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%s not found", name),
	}
}

func NewBadRequestError(msg string) *HttpError {
	return &HttpError{
		Code:    http.StatusBadRequest,
		Message: msg,
	}
}
