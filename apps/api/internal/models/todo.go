package models

type Todo struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Assigned    string `json:"assigned"`
}
