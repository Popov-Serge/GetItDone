package apperror

import "errors"

var (
	ErrPhoneAlreadyExists  = errors.New("phone already exists")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidPhone        = errors.New("invalid phone")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrFamilyNotFound      = errors.New("family not found")
)

type RegistrationConflictError struct {
	Fields map[string][]string
}

func (e *RegistrationConflictError) Error() string {
	return "registration conflict"
}
