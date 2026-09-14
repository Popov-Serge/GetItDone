package handler

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type RequestValidator struct {
	validate *validator.Validate
}

func NewRequestValidator(v *validator.Validate) *RequestValidator {
	return &RequestValidator{validate: v}
}

func (rv *RequestValidator) ValidateRequest(
	w http.ResponseWriter,
	req any,
) bool {
	err := rv.validate.Struct(req)
	if err == nil {
		return true
	}

	var validationErrors validator.ValidationErrors

	if errors.As(err, &validationErrors) {
		fieldErrors := make(map[string][]string)

		for _, err := range validationErrors {
			fieldErrors[err.Field()] = append(
				fieldErrors[err.Field()],
				err.Tag(),
			)
		}

		writeError(
			w,
			http.StatusUnprocessableEntity,
			"VALIDATION_ERROR",
			"Request validation failed",
			fieldErrors,
		)

		return false
	}

	writeError(
		w,
		http.StatusUnprocessableEntity,
		"VALIDATION_ERROR",
		"Request validation failed",
		nil,
	)

	return false
}
