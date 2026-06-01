package repository

import (
	"context"
	"fmt"

	"react-todos/apps/api/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
|--------------------------------------------------------------------------
| Repository
|--------------------------------------------------------------------------
*/

type TodoRepository struct {
	DB *pgxpool.Pool
}

func NewTodoRepository(db *pgxpool.Pool) *TodoRepository {
	return &TodoRepository{DB: db}
}

/*
|--------------------------------------------------------------------------
| Queries
|--------------------------------------------------------------------------
*/

// Get all todos for a user
func (r *TodoRepository) GetAllByUser(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
	sortBy, sortOrder string,
	filterCompleted *bool,
	filterAssigned string,
) ([]models.Todo, int, error) {
	// Build the base query
	query := `
		SELECT id, user_id, assigned_to_name, description, completed, created_at, updated_at, count(*) OVER() as total_count
		FROM todos
		WHERE user_id = $1
	`
	
	// Build arguments slice
	args := []any{userID}
	argIndex := 2
	
	// Add filters
	if filterCompleted != nil {
		query += fmt.Sprintf(" AND completed = $%d", argIndex)
		args = append(args, *filterCompleted)
		argIndex++
	}
	
	if filterAssigned != "" {
		query += fmt.Sprintf(" AND assigned_to_name ILIKE $%d", argIndex)
		args = append(args, "%"+filterAssigned+"%")
		argIndex++
	}
	
	// Add sorting
	switch sortBy {
	case "created_at":
		query += fmt.Sprintf(" ORDER BY created_at %s", sortOrder)
	case "description":
		query += fmt.Sprintf(" ORDER BY description %s", sortOrder)
	case "assigned_to_name":
		query += fmt.Sprintf(" ORDER BY assigned_to_name %s", sortOrder)
	case "updated_at":
		query += fmt.Sprintf(" ORDER BY updated_at %s", sortOrder)
	default:
		query += " ORDER BY created_at DESC"
	}
	
	// Add pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)
	
	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var todos []models.Todo
	total := 0
	for rows.Next() {
		var t models.Todo
		var tc int
		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.AssignedToName,
			&t.Description,
			&t.Completed,
			&t.CreatedAt,
			&t.UpdatedAt,
			&tc,
		); err != nil {
			return nil, 0, err
		}
		total = tc
		todos = append(todos, t)
	}

	// If no rows were returned, still compute total count
	if total == 0 {
		countQuery := `SELECT count(*) FROM todos WHERE user_id = $1`
		countArgs := []any{userID}
		countArgIndex := 2
		
		if filterCompleted != nil {
			countQuery += fmt.Sprintf(" AND completed = $%d", countArgIndex)
			countArgs = append(countArgs, *filterCompleted)
			countArgIndex++
		}
		
		if filterAssigned != "" {
			countQuery += fmt.Sprintf(" AND assigned_to_name ILIKE $%d", countArgIndex)
			countArgs = append(countArgs, "%"+filterAssigned+"%")
		}
		
		var cnt int
		if err := r.DB.QueryRow(ctx, countQuery, countArgs...).Scan(&cnt); err == nil {
			total = cnt
		}
	}

	return todos, total, nil
}

// Create a todo
func (r *TodoRepository) Create(
	ctx context.Context,
	todo models.Todo,
) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO todos (user_id, assigned_to_name, description)
		VALUES ($1, $2, $3)
	`, todo.UserID, todo.AssignedToName, todo.Description)

	return err
}

// Update a todo (user-scoped)
func (r *TodoRepository) Update(
	ctx context.Context,
	id int,
	userID uuid.UUID,
	todo models.Todo,
) error {
	cmd, err := r.DB.Exec(ctx, `
		UPDATE todos
		SET description = $1,
		assigned_to_name = $2,
		    completed = $3,
		    updated_at = now()
		WHERE id = $4 AND user_id = $5
	`, todo.Description, todo.AssignedToName, todo.Completed, id, userID)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFoundOrForbidden
	}

	return nil
}

// Delete a todo (user-scoped)
func (r *TodoRepository) Delete(
	ctx context.Context,
	id int,
	userID uuid.UUID,
) error {
	cmd, err := r.DB.Exec(ctx, `
		DELETE FROM todos
		WHERE id = $1 AND user_id = $2
	`, id, userID)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFoundOrForbidden
	}

	return nil
}
