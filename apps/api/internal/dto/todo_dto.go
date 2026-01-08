package dto

type TodoResponse struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Assigned    string `json:"assigned"`
}

type CreateTodoRequest struct {
	Description string `json:"description"`
	Assigned    string `json:"assigned"`
}

type UpdateTodoRequest struct {
	Description string `json:"description"`
	Assigned    string `json:"assigned"`
}