package services

import (
	"context"
	"errors"
	"time"

	"react-todos/apps/api/internal/models"
	"react-todos/apps/api/internal/repository"

	"github.com/google/uuid"
)

type TodoService struct {
	repo *repository.TodoRepository
}

func NewTodoService(repo *repository.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

/*
GET
*/
func (s *TodoService) GetAll(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
	sortBy, sortOrder string,
	filterCompleted *bool,
	filterAssigned string,
) ([]models.Todo, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return s.repo.GetAllByUser(ctx, userID, limit, offset, sortBy, sortOrder, filterCompleted, filterAssigned)
}

/*
CREATE
*/
func (s *TodoService) Create(
	ctx context.Context,
	userID uuid.UUID,
	assignedToName string,
	description string,
) error {
	if description == "" {
		return errors.New("description is required")
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return s.repo.Create(ctx, models.Todo{
		UserID:         userID,
		AssignedToName: assignedToName,
		Description:    description,
	})
}

/*
UPDATE
*/
func (s *TodoService) Update(
	ctx context.Context,
	userID uuid.UUID,
	id int,
	assignedToName string,
	description string,
	completed bool,
) error {
	if assignedToName == "" {
		return errors.New("assigned_to_name is required")
	}
	if description == "" {
		return errors.New("description is required")
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return s.repo.Update(
		ctx,
		id,
		userID,
		models.Todo{
			AssignedToName: assignedToName,
			Description:    description,
			Completed:      completed,
		},
	)
}

/*
DELETE
*/
func (s *TodoService) Delete(
	ctx context.Context,
	userID uuid.UUID,
	id int,
) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return s.repo.Delete(ctx, id, userID)
}
