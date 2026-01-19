package models

import (
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	ID             int       `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Description    string    `json:"description"`
	AssignedToName string    `json:"assigned_to_name"`
	Completed      bool      `json:"completed"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
