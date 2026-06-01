package handlers

import (
	"encoding/json"
	"net/http"

	"react-todos/apps/api/internal/dto"
	"react-todos/apps/api/internal/middleware"
	"react-todos/apps/api/internal/services"
)

var adminAuthService services.AuthServicer

func InitAdminHandlers(service services.AuthServicer) {
	adminAuthService = service
}

func RevokeUserTokens(w http.ResponseWriter, r *http.Request) {
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

	if err := adminAuthService.BlacklistAllForUser(r.Context(), req.UserID); err != nil {
		middleware.SendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := dto.SuccessResponse(map[string]string{"status": "revoked"})
	_ = json.NewEncoder(w).Encode(response)
}

func UnblockUser(w http.ResponseWriter, r *http.Request) {
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

	if err := adminAuthService.UnblockUser(r.Context(), req.UserID); err != nil {
		middleware.SendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := dto.SuccessResponse(map[string]string{"status": "unblocked"})
	_ = json.NewEncoder(w).Encode(response)
}
