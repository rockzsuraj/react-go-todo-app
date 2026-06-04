package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"react-todos/apps/api/internal/middleware"
	"react-todos/apps/api/internal/models"

	"github.com/google/uuid"
)

// --- Mock Service ---

type MockTodoService struct {
	GetAllFunc func(ctx context.Context, userID uuid.UUID, limit, offset int, sortBy, sortOrder string, filterCompleted *bool, filterAssigned string) ([]models.Todo, int, error)
	CreateFunc func(ctx context.Context, userID uuid.UUID, assignedToName, description string) error
	UpdateFunc func(ctx context.Context, userID uuid.UUID, id int, assignedToName, description string, completed bool) error
	DeleteFunc func(ctx context.Context, userID uuid.UUID, id int) error
}

func (m *MockTodoService) GetAll(ctx context.Context, userID uuid.UUID, limit, offset int, sortBy, sortOrder string, filterCompleted *bool, filterAssigned string) ([]models.Todo, int, error) {
	return m.GetAllFunc(ctx, userID, limit, offset, sortBy, sortOrder, filterCompleted, filterAssigned)
}

func (m *MockTodoService) Create(ctx context.Context, userID uuid.UUID, assignedToName, description string) error {
	return m.CreateFunc(ctx, userID, assignedToName, description)
}

func (m *MockTodoService) Update(ctx context.Context, userID uuid.UUID, id int, assignedToName, description string, completed bool) error {
	return m.UpdateFunc(ctx, userID, id, assignedToName, description, completed)
}

func (m *MockTodoService) Delete(ctx context.Context, userID uuid.UUID, id int) error {
	return m.DeleteFunc(ctx, userID, id)
}

// --- Tests ---

func TestGetTodos(t *testing.T) {
	// Setup user ID
	userID := uuid.New()

	// Create request
	req, _ := http.NewRequest("GET", "/api/todos", nil)

	// Mock middleware context
	ctx := middleware.WithUserID(req.Context(), userID.String())
	req = req.WithContext(ctx)

	// Mock Service
	mockService := &MockTodoService{
		GetAllFunc: func(ctx context.Context, uid uuid.UUID, limit, offset int, sortBy, sortOrder string, filterCompleted *bool, filterAssigned string) ([]models.Todo, int, error) {
			if uid != userID {
				t.Errorf("Expected userID %s, got %s", userID, uid)
			}
			return []models.Todo{
				{ID: 1, Description: "Test Todo", AssignedToName: "Me", Completed: false},
			}, 1, nil
		},
	}
	InitHandlers(mockService)

	// Recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetTodos)
	handler.ServeHTTP(rr, req)

	// Assertions
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Read response directly as map
	var respMap map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&respMap); err != nil {
		t.Fatal(err)
	}

	data, ok := respMap["data"].([]interface{})
	if !ok {
		t.Fatal("Response data is not an array")
	}
	if len(data) != 1 {
		t.Errorf("Expected 1 todo, got %d", len(data))
	}
}

func TestCreateTodo_Validation(t *testing.T) {
	// Setup user ID
	userID := uuid.New()

	// Invalid body (missing description)
	body := []byte(`{"assigned_to_name": "Me"}`)

	req, _ := http.NewRequest("POST", "/api/todos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := middleware.WithUserID(req.Context(), userID.String())
	req = req.WithContext(ctx)

	// Mock Service (should not be called)
	mockService := &MockTodoService{
		CreateFunc: func(ctx context.Context, uid uuid.UUID, assignedToName, description string) error {
			t.Fatal("Service should not be called on validation error")
			return nil
		},
	}
	InitHandlers(mockService)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CreateTodoHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}
