package handlers

import (
	"encoding/json"
	"net/http"

	"react-todos/apps/api/internal/delivery/http/dto"
	"react-todos/apps/api/internal/delivery/http/middleware"
	"react-todos/apps/api/internal/domain/services"
)

// AdminHandler holds the dependencies for admin-only endpoints.
// Uses the same AuthServicer as AuthHandler — no duplicate injection needed.
type AdminHandler struct {
	service services.AuthServicer
}

// NewAdminHandler constructs an AdminHandler.
func NewAdminHandler(service services.AuthServicer) *AdminHandler {
	return &AdminHandler{service: service}
}

func (h *AdminHandler) RevokeUserTokens(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Unauthorized")
		return
	}

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendJSONErrorWithCode(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.UserID == "" {
		middleware.SendJSONErrorWithCode(w, http.StatusBadRequest, "ERR_MISSING_USER_ID", "user_id field is required")
		return
	}

	if err := h.service.BlacklistAllForUser(r.Context(), req.UserID); err != nil {
		middleware.SendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto.SuccessResponse(map[string]string{"status": "revoked"}))
}

func (h *AdminHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Unauthorized")
		return
	}

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendJSONErrorWithCode(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.UserID == "" {
		middleware.SendJSONErrorWithCode(w, http.StatusBadRequest, "ERR_MISSING_USER_ID", "user_id field is required")
		return
	}

	if err := h.service.UnblockUser(r.Context(), req.UserID); err != nil {
		middleware.SendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto.SuccessResponse(map[string]string{"status": "unblocked"}))
}
