package handler

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Fields  map[string][]string `json:"fields,omitempty"`
}

func writeError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
	fields map[string][]string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
			Fields:  fields,
		},
	})
}
