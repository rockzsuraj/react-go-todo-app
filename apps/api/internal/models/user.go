package models

import "time"

type User struct {
	ID             string
	Provider       string
	ProviderUserID string
	Email          string
	Name           string
	Picture        string
	Role           string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastLoginAt    *time.Time
}
