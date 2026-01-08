package models

type Todo struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Assigned    string `json:"assigned"`
}
