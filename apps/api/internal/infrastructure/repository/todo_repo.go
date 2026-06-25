package repository

import (
	"context"
	"fmt"

	"react-todos/apps/api/internal/domain/models"

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
| Filter builder
|--------------------------------------------------------------------------
*/

// todoFilterClause builds the WHERE conditions that are shared between the
// main paginated query and the fallback count query.
//
// It returns the additional SQL fragment (starting with " AND ...") and the
// extra arguments to append after userID ($1). argStart is the index of the
// first placeholder to use (usually 2, since $1 is userID).
func todoFilterClause(argStart int, filterCompleted *bool, filterAssigned string) (clause string, args []any, nextArgIndex int) {
	idx := argStart
	if filterCompleted != nil {
		clause += fmt.Sprintf(" AND completed = $%d", idx)
		args = append(args, *filterCompleted)
		idx++
	}
	if filterAssigned != "" {
		clause += fmt.Sprintf(" AND assigned_to_name ILIKE $%d", idx)
		args = append(args, "%"+filterAssigned+"%")
		idx++
	}
	return clause, args, idx
}

/*
|--------------------------------------------------------------------------
| Queries
|--------------------------------------------------------------------------
*/

// GetAllByUser returns a page of todos for a user together with the total
// matching row count. The total is obtained from the window function on
// non-empty pages; a single fallback COUNT(*) is issued only when the page
// is empty (e.g. the caller requested a page beyond the last one).
// Both paths share todoFilterClause so filter logic is never duplicated.
func (r *TodoRepository) GetAllByUser(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
	sortBy, sortOrder string,
	filterCompleted *bool,
	filterAssigned string,
) ([]models.Todo, int, error) {
	// Build shared filter clause (args start at $2; $1 is userID)
	filterSQL, filterArgs, nextIdx := todoFilterClause(2, filterCompleted, filterAssigned)

	// ── Main query ────────────────────────────────────────────────────────
	// count(*) OVER() gives us the total matching rows without a second trip.
	query := `SELECT id, user_id, assigned_to_name, description, completed, created_at, updated_at,
	                 count(*) OVER() AS total_count
	          FROM todos
	          WHERE user_id = $1` + filterSQL

	// Sorting — sortBy and sortOrder are validated by the handler before reaching here.
	switch sortBy {
	case "description":
		query += fmt.Sprintf(" ORDER BY description %s", sortOrder)
	case "assigned_to_name":
		query += fmt.Sprintf(" ORDER BY assigned_to_name %s", sortOrder)
	case "updated_at":
		query += fmt.Sprintf(" ORDER BY updated_at %s", sortOrder)
	default: // "created_at" and anything else
		query += fmt.Sprintf(" ORDER BY created_at %s", sortOrder)
	}

	// Pagination placeholders follow immediately after the filter args.
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", nextIdx, nextIdx+1)

	// Assemble full args: userID, filter args, limit, offset
	args := append([]any{userID}, filterArgs...)
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
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// ── Fallback count ────────────────────────────────────────────────────
	// Reached only when the requested page is empty (offset past last row).
	// Uses the same todoFilterClause so filter logic stays in one place.
	if total == 0 {
		countSQL, countFilterArgs, _ := todoFilterClause(2, filterCompleted, filterAssigned)
		countQuery := `SELECT count(*) FROM todos WHERE user_id = $1` + countSQL
		countArgs := append([]any{userID}, countFilterArgs...)

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
