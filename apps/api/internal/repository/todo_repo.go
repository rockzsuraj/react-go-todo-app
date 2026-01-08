package repository

import (
	"context"
	"react-todos/apps/api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoRepository struct {
	DB *pgxpool.Pool
}

func NewTodoRepository(db *pgxpool.Pool) *TodoRepository {
	return &TodoRepository{DB: db}
}

func (r *TodoRepository) GetAll(ctx context.Context) ([]models.Todo, error) {
	rows, err := r.DB.Query(ctx,
		"SELECT id, description, assigned FROM todos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var todos []models.Todo
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.Description, &t.Assigned); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, nil
}

func (r *TodoRepository) Create(ctx context.Context, todo models.Todo) error {
	_, err := r.DB.Exec(
		ctx,
		"INSERT INTO todos (description, assigned) VALUES ($1, $2)",
		todo.Description,
		todo.Assigned,
	)
	return err
}

func (r *TodoRepository) Update(ctx context.Context, id int, todo models.Todo) error {
	_, err := r.DB.Exec(
		ctx,
		"UPDATE todos SET description = $1, assigned = $2 WHERE id = $3",
		todo.Description,
		todo.Assigned,
		id,
	)
	return err
}

func (r *TodoRepository) Delete(ctx context.Context, id int) error {
	_, err := r.DB.Exec(
		ctx,
		"DELETE FROM todos WHERE id = $1",
		id,
	)
	return err
}
