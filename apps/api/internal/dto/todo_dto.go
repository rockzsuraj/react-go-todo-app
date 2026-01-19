package dto

import "time"

/*
========================
 RESPONSE DTO
========================
*/

type TodoResponse struct {
	ID             int       `json:"id"`
	Description    string    `json:"description"`
	AssignedToName string    `json:"assigned_to_name"`
	Completed      bool      `json:"completed"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

/*
========================
 CREATE REQUEST DTO
========================
*/

type CreateTodoRequest struct {
	Description    string `json:"description" validate:"required"`
	AssignedToName string `json:"assigned_to_name" validate:"required"`
}

/*
========================
 UPDATE REQUEST DTO
========================
*/

type UpdateTodoRequest struct {
	Description    *string `json:"description,omitempty"`
	AssignedToName *string `json:"assigned_to_name,omitempty"`
	Completed      *bool   `json:"completed,omitempty"`
}
