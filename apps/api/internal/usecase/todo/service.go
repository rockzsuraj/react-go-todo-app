package todo

import (
	"context"
	"errors"
	"time"

	"react-todos/apps/api/internal/domain/models"
	"react-todos/apps/api/internal/domain/repository"
	"react-todos/apps/api/internal/domain/services"

	"github.com/google/uuid"
)

// TodoService implements TodoServicer. It depends on the TodoRepository
// interface — not the concrete pgx type — so it can be tested with any mock.
type TodoService struct {
	repo      repository.TodoRepository
	publisher services.TodoEventPublisher
}

// NewTodoService accepts any TodoRepository implementation (real or mock) and TodoEventPublisher.
func NewTodoService(repo repository.TodoRepository, publisher services.TodoEventPublisher) *TodoService {
	return &TodoService{repo: repo, publisher: publisher}
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

	err := s.repo.Create(ctx, models.Todo{
		UserID:         userID,
		AssignedToName: assignedToName,
		Description:    description,
	})
	if err != nil {
		return err
	}

	s.publisher.PublishTodoChange(ctx, userID)
	return nil
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

	err := s.repo.Update(
		ctx,
		id,
		userID,
		models.Todo{
			AssignedToName: assignedToName,
			Description:    description,
			Completed:      completed,
		},
	)
	if err != nil {
		return err
	}

	s.publisher.PublishTodoChange(ctx, userID)
	return nil
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

	err := s.repo.Delete(ctx, id, userID)
	if err != nil {
		return err
	}

	s.publisher.PublishTodoChange(ctx, userID)
	return nil
}

