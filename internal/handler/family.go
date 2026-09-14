package handler

import (
	"encoding/json"
	"errors"
	"getitdone/internal/apperror"
	"getitdone/internal/domain"
	"getitdone/internal/middleware"
	"getitdone/internal/service"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type FamilyHandler struct {
	familyService *service.FamilyService
	logger        *slog.Logger
	validator     *RequestValidator
}

type FamilyCreateRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type FamilyUpdateRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

func NewFamilyHandler(
	familyService *service.FamilyService,
	logger *slog.Logger,
	validator *RequestValidator,
) *FamilyHandler {

	return &FamilyHandler{
		familyService: familyService,
		logger:        logger,
		validator:     validator,
	}
}

func (h *FamilyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req FamilyCreateRequest

	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		writeError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"User not authenticated",
			nil,
		)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid Request Body",
			nil,
		)
		return
	}

	if !h.validator.ValidateRequest(w, &req) {
		return
	}

	family := domain.Family{
		CreatedBy: userID,
		Name:      req.Name,
	}

	if err := h.familyService.Create(r.Context(), &family); err != nil {
		h.logger.Error("failed to create family", "error", err)
		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to create family",
			nil,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(family); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *FamilyHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r)
	if !ok {
		writeError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"User not authenticated",
			nil,
		)
		return
	}

	familyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid family ID",
			nil,
		)
		return
	}

	family, err := h.familyService.GetByID(
		r.Context(),
		familyID,
		userID,
	)
	if err != nil {
		if errors.Is(err, apperror.ErrFamilyNotFound) {
			writeError(
				w,
				http.StatusNotFound,
				"FAMILY_NOT_FOUND",
				"Family not found",
				nil,
			)
			return
		}

		h.logger.Error(
			"failed to get family",
			"error", err,
		)

		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to get family",
			nil,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(family); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *FamilyHandler) GetFamilies(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		writeError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"User not authenticated",
			nil,
		)
		return
	}

	families, err := h.familyService.GetFamilies(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get families", "error", err)

		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"failed to get families",
			nil,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(families); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *FamilyHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req FamilyUpdateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID REQUEST",
			"Couldn't read json",
			nil,
		)

		return
	}

	if !h.validator.ValidateRequest(w, &req) {
		return
	}

	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		writeError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"User not authenticated",
			nil,
		)
		return
	}

	familyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid family ID",
			nil,
		)
		return
	}

	family := &domain.Family{
		ID:        familyID,
		Name:      req.Name,
		CreatedBy: userID,
	}

	err = h.familyService.Update(r.Context(), family)
	if err != nil {
		if errors.Is(err, apperror.ErrFamilyNotFound) {
			writeError(
				w,
				http.StatusNotFound,
				"FAMILY_NOT_FOUND",
				"Family not found",
				nil,
			)
			return
		}

		h.logger.Error(
			"failed to update family",
			"error", err,
		)

		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to update family",
			nil,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(family); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *FamilyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	familyID, err := uuid.Parse(chi.URLParam(r, "id"))

	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid family ID",
			nil,
		)
		return
	}

	err = h.familyService.Delete(r.Context(), familyID, userID)
	if err != nil {
		if errors.Is(err, apperror.ErrFamilyNotFound) {
			writeError(
				w,
				http.StatusNotFound,
				"FAMILY_NOT_FOUND",
				"Family not found",
				nil,
			)
			return
		}

		writeError(w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to delete family",
			nil,
		)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
