package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"getitdone/internal/apperror"
	"getitdone/internal/middleware"
	"getitdone/internal/service"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Phone    string `json:"phone" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Timezone string `json:"timezone" validate:"required"`
}

type LoginRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthHandler struct {
	authService *service.AuthService
	logger      *slog.Logger
	validator   *RequestValidator
}

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Phone    string    `json:"phone"`
	Email    string    `json:"email"`
	Timezone string    `json:"timezone"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func NewAuthHandler(
	authService *service.AuthService,
	logger *slog.Logger,
	requestValidator *RequestValidator,
) *AuthHandler {

	return &AuthHandler{
		authService: authService,
		logger:      logger,
		validator:   requestValidator,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			nil,
		)
		return
	}

	if !h.validator.ValidateRequest(w, req) {
		return
	}

	input := service.RegisterInput{
		Name:     req.Name,
		Phone:    req.Phone,
		Email:    req.Email,
		Password: req.Password,
		Timezone: req.Timezone,
	}

	user, err := h.authService.Register(r.Context(), input)
	if err != nil {
		var conflictErr *apperror.RegistrationConflictError

		if errors.As(err, &conflictErr) {
			writeError(
				w,
				http.StatusConflict,
				"CONFLICT",
				"User already exists",
				conflictErr.Fields,
			)

			return
		}

		switch {
		case errors.Is(err, apperror.ErrPhoneAlreadyExists):
			writeError(
				w,
				http.StatusConflict,
				"CONFLICT",
				"User already exists",
				map[string][]string{
					"phone": {"already_exists"},
				},
			)

		case errors.Is(err, apperror.ErrEmailAlreadyExists):
			writeError(
				w,
				http.StatusConflict,
				"CONFLICT",
				"User already exists",
				map[string][]string{
					"email": {"already_exists"},
				},
			)

		case errors.Is(err, apperror.ErrInvalidPhone):
			writeError(
				w,
				http.StatusUnprocessableEntity,
				"VALIDATION_ERROR",
				"Request validation failed",
				map[string][]string{
					"phone": {"invalid"},
				},
			)

		default:
			h.logger.Error(
				"failed to register user",
				"error",
				err,
			)

			writeError(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"Internal server error",
				nil,
			)
		}

		return
	}

	res := UserResponse{
		ID:       user.ID,
		Name:     user.Name,
		Phone:    user.Phone,
		Email:    user.Email,
		Timezone: user.Timezone,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.logger.Error(
			"failed to encode register response",
			"error",
			err,
		)
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			nil,
		)

		return
	}

	if !h.validateRequest(w, req) {
		return
	}

	input := service.LoginInput{
		Phone:    req.Phone,
		Password: req.Password,
	}

	tokenPair, err := h.authService.Login(
		r.Context(),
		input,
	)

	if err != nil {
		if errors.Is(err, apperror.ErrInvalidCredentials) {
			writeError(
				w,
				http.StatusUnauthorized,
				"INVALID_CREDENTIALS",
				"Invalid credentials",
				nil,
			)

			return
		}

		h.logger.Error(
			"failed to login",
			"error",
			err,
		)

		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Internal server error",
			nil,
		)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(tokenPair); err != nil {
		h.logger.Error(
			"failed to encode login response",
			"error",
			err,
		)
	}
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request",
			nil,
		)
		return
	}

	if !h.validateRequest(w, req) {
		return
	}

	hash := sha256.Sum256([]byte(req.RefreshToken))
	tokenHash := hex.EncodeToString(hash[:])

	tokenPair, err := h.authService.Refresh(r.Context(), tokenHash)

	if err != nil {
		if errors.Is(err, apperror.ErrInvalidRefreshToken) {
			writeError(
				w,
				http.StatusUnauthorized,
				"INVALID_REFRESH_TOKEN",
				"Invalid refresh token",
				nil,
			)
			return
		}

		h.logger.Error(
			"failed to refresh tokens",
			"error",
			err,
		)

		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Internal server error",
			nil,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(tokenPair); err != nil {
		h.logger.Error(
			"failed to encode refresh response",
			"error",
			err,
		)
	}
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	jti, ok := middleware.JTI(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	expiresAt, ok := middleware.ExpiresAt(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req RefreshRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request body",
			nil,
		)
		return
	}

	if !h.validateRequest(w, req) {
		return
	}

	hash := sha256.Sum256([]byte(req.RefreshToken))
	refreshTokenHash := hex.EncodeToString(hash[:])

	err := h.authService.Logout(
		r.Context(),
		userID,
		jti,
		expiresAt,
		refreshTokenHash,
	)
	if err != nil {
		h.logger.Error(
			"failed to logout",
			"error",
			err,
			"user_id",
			userID,
		)

		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Internal server error",
			nil,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) validateRequest(
	w http.ResponseWriter,
	req any,
) bool {
	err := h.validator.validate.Struct(req)
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
