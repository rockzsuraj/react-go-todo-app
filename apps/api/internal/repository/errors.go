package repository

import "errors"

// Domain-level errors (used across services & middleware)
var (
	ErrNotFoundOrForbidden = errors.New("resource not found or access forbidden")
)
