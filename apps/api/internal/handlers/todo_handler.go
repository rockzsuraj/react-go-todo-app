package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"react-todos/apps/api/internal/dto"
	"react-todos/apps/api/internal/middleware"
	"react-todos/apps/api/internal/services"
	"react-todos/apps/api/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

/*
|--------------------------------------------------------------------------
| Init
|--------------------------------------------------------------------------
*/

var todoService services.TodoServicer

func InitHandlers(service services.TodoServicer) {
	todoService = service
}

/*
|--------------------------------------------------------------------------
| Helpers
|--------------------------------------------------------------------------
*/

func getUserUUID(r *http.Request) (uuid.UUID, error) {
	userIDStr := middleware.UserIDFromContext(r.Context())
	return uuid.Parse(userIDStr)
}

/*
|--------------------------------------------------------------------------
| Handlers
|--------------------------------------------------------------------------
*/

/*
GET /api/todos
*/
func GetTodos(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserUUID(r)
	if err != nil {
		middleware.SendJSONErrorWithCode(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Unauthorized")
		return
	}

	// Parse pagination query params with validation
	page := 1
	limit := 10 // Default limit
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 && v <= 1000 {
			page = v
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	// Parse sorting parameters with validation
	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := strings.ToUpper(r.URL.Query().Get("sort_order"))
	allowedSortFields := map[string]bool{
		"created_at": true, "updated_at": true, "description": true, "completed": true,
	}
	if !allowedSortFields[sortBy] {
		sortBy = "created_at"
	}
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "ASC"
	}

	// Parse filtering parameters with validation
	var filterCompleted *bool
	if fc := r.URL.Query().Get("completed"); fc != "" {
		if v, err := strconv.ParseBool(fc); err == nil {
			filterCompleted = &v
		}
	}

	filterAssigned := r.URL.Query().Get("assigned")
	if len(filterAssigned) > 100 {
		middleware.SendJSONErrorWithCode(w, http.StatusBadRequest, "ERR_INVALID_FILTER", "Filter value too long")
		return
	}

	offset := (page - 1) * limit

	todos, total, err := todoService.GetAll(r.Context(), userID, limit, offset, sortBy, sortOrder, filterCompleted, filterAssigned)
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	// Map to response DTOs and ensure we return an empty array (not null)
	resp := make([]dto.TodoResponse, 0, len(todos))
	for _, t := range todos {
		resp = append(resp, dto.TodoResponse{
			ID:             t.ID,
			Description:    t.Description,
			AssignedToName: t.AssignedToName,
			Completed:      t.Completed,
			CreatedAt:      t.CreatedAt,
			UpdatedAt:      t.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	meta := &dto.Meta{Total: total, Page: page, Limit: limit, Offset: offset}
	_ = json.NewEncoder(w).Encode(dto.SuccessResponseWithMeta(resp, meta))
}

/*
POST /api/todos
Body: { "description": "Buy milk" }
*/
func CreateTodoHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserUUID(r)
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	var body dto.CreateTodoRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.SendError(w, err)
		return
	}

	if err := utils.ValidateStruct(body); err != nil {
		middleware.SendValidationError(w, err)
		return
	}

	if err := todoService.Create(
		r.Context(),
		userID,
		body.AssignedToName,
		body.Description,
	); err != nil {
		middleware.SendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(dto.SuccessResponse(nil))
}

/*
PUT /api/todos/{id}
Body: { "description": "...", "completed": true }
*/
func UpdateTodoHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserUUID(r)
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	var body dto.UpdateTodoRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.SendError(w, err)
		return
	}

	if err := utils.ValidateStruct(body); err != nil {
		middleware.SendValidationError(w, err)
		return
	}

	// Provide sensible defaults if fields are nil when calling service
	desc := ""
	if body.Description != nil {
		desc = *body.Description
	}
	assignedToName := ""
	if body.AssignedToName != nil {
		assignedToName = *body.AssignedToName
	}
	completed := false
	if body.Completed != nil {
		completed = *body.Completed
	}

	if err := todoService.Update(
		r.Context(),
		userID,
		id,
		assignedToName,
		desc,
		completed,
	); err != nil {
		middleware.SendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto.SuccessResponse(nil))
}

/*
DELETE /api/todos/{id}
*/
func DeleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserUUID(r)
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		middleware.SendError(w, err)
		return
	}

	if err := todoService.Delete(
		r.Context(),
		userID,
		id,
	); err != nil {
		middleware.SendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto.SuccessResponse(nil))
}
