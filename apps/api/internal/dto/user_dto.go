package dto

import (
	"time"
	"react-todos/apps/api/internal/models"
)

// UserResponse represents the user data returned to clients
type UserResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Picture     string     `json:"picture"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// NewUserResponse creates a user response DTO from a user model
func NewUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Picture:     user.Picture,
		Role:        user.Role,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLoginAt: user.LastLoginAt,
	}
}
