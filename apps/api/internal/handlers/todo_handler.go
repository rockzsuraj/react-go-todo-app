package handlers

import (
	"encoding/json"
	"net/http"
	"react-todos/apps/api/internal/dto"
	"react-todos/apps/api/internal/models"
	"react-todos/apps/api/internal/services"
	"strconv"

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
