package services

import (
	"context"
	"errors"
	"react-todos/apps/api/internal/models"
	"react-todos/apps/api/internal/repository"
	"time"
)

type TodoService struct {
	repo *repository.TodoRepository
}

func NewTodoService(repo *repository.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

func (s *TodoService) GetAll(ctx context.Context) ([]models.Todo, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	
	return s.repo.GetAll(ctx)
}

func (s *TodoService) Create(ctx context.Context, todo models.Todo) error {
	if todo.Description == "" {
		return errors.New("description is required")
	}
	if todo.Assigned == "" {
		return errors.New("assigned person is required")
	}
	
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	
	return s.repo.Create(ctx, todo)
}

func (s *TodoService) Update(ctx context.Context, id int, todo models.Todo) error {
	if todo.Description == "" {
		return errors.New("description is required")
	}
	if todo.Assigned == "" {
		return errors.New("assigned person is required")
	}
	
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	
	return s.repo.Update(ctx, id, todo)
}

func (s *TodoService) Delete(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	
	return s.repo.Delete(ctx, id)
}