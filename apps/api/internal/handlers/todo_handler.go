package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"react-todos/apps/api/internal/dto"
	"react-todos/apps/api/internal/models"
	"react-todos/apps/api/internal/services"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

var todoService *services.TodoService

func InitHandlers(service *services.TodoService) {
	todoService = service
}

func GetTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	todos, err := todoService.GetAll(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(dto.ErrorResponse(
			"FETCH_TODOS_ERROR",
			"Failed to fetch todos",
			err.Error(),
		))
		return
	}
	
	response := make([]dto.TodoResponse, len(todos))
	for i, todo := range todos {
		response[i] = dto.TodoResponse{
			ID:          todo.ID,
			Description: todo.Description,
			Assigned:    todo.Assigned,
		}
	}
	
	meta := &dto.Meta{
		Total: len(response),
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto.SuccessResponseWithMeta(response, meta))
}

func CreateTodoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var req dto.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.ErrorResponse(
			"INVALID_REQUEST_BODY",
			"Invalid request body format",
			err.Error(),
		))
		return
	}

	// 🔒 FIELD VALIDATION
	if err := validateTodoFields(req.Description, req.Assigned); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.ErrorResponse(
			"VALIDATION_ERROR",
			"Invalid field values",
			err.Error(),
		))
		return
	}

	todo := models.Todo{
		Description: req.Description,
		Assigned:    req.Assigned,
	}

	if err := todoService.Create(r.Context(), todo); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.ErrorResponse(
			"CREATE_TODO_ERROR",
			"Failed to create todo",
			err.Error(),
		))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.SuccessResponse(map[string]string{
		"message": "Todo created successfully",
	}))
}

func UpdateTodoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.ErrorResponse(
			"INVALID_TODO_ID",
			"Invalid todo ID format",
			"ID must be a valid integer",
		))
		return
	}

	var req dto.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.ErrorResponse(
			"INVALID_REQUEST_BODY",
			"Invalid request body format",
			err.Error(),
		))
		return
	}

	todo := models.Todo{
		Description: req.Description,
		Assigned:    req.Assigned,
	}

	if err := todoService.Update(r.Context(), id, todo); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.ErrorResponse(
			"UPDATE_TODO_ERROR",
			"Failed to update todo",
			err.Error(),
		))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto.SuccessResponse(map[string]string{
		"message": "Todo updated successfully",
	}))
}

func DeleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.ErrorResponse(
			"INVALID_TODO_ID",
			"Invalid todo ID format",
			"ID must be a valid integer",
		))
		return
	}

	if err := todoService.Delete(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(dto.ErrorResponse(
			"DELETE_TODO_ERROR",
			"Failed to delete todo",
			err.Error(),
		))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto.SuccessResponse(map[string]string{
		"message": "Todo deleted successfully",
	}))
}

// 🔒 FIELD VALIDATION FUNCTIONS
func validateTodoFields(description, assigned string) error {
	// Check if fields are empty
	if description == "" {
		return fmt.Errorf("description is required")
	}
	if assigned == "" {
		return fmt.Errorf("assigned field is required")
	}
	
	// Check field lengths
	if len(description) < 3 {
		return fmt.Errorf("description must be at least 3 characters")
	}
	if len(description) > 500 {
		return fmt.Errorf("description must be less than 500 characters")
	}
	if len(assigned) < 2 {
		return fmt.Errorf("assigned must be at least 2 characters")
	}
	if len(assigned) > 100 {
		return fmt.Errorf("assigned must be less than 100 characters")
	}
	
	// Check for malicious content
	if containsMaliciousContent(description) {
		return fmt.Errorf("description contains invalid characters")
	}
	if containsMaliciousContent(assigned) {
		return fmt.Errorf("assigned contains invalid characters")
	}
	
	return nil
}

func containsMaliciousContent(input string) bool {
	// Check for script tags and SQL injection
	dangerous := []string{
		"<script", "</script>", "javascript:", "<iframe",
		"SELECT", "INSERT", "DELETE", "UPDATE", "DROP",
		"UNION", "--", "/*", "*/", ";",
	}
	
	lower := strings.ToLower(input)
	for _, pattern := range dangerous {
		if strings.Contains(lower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}
