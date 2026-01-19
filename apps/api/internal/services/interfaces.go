package services

import (
	"context"
	"react-todos/apps/api/internal/models"

	"github.com/google/uuid"
)

type TodoServicer interface {
	GetAll(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Todo, int, error)
	Create(ctx context.Context, userID uuid.UUID, assignedToName, description string) error
	Update(ctx context.Context, userID uuid.UUID, id int, assignedToName, description string, completed bool) error
	Delete(ctx context.Context, userID uuid.UUID, id int) error
}
