package handlers

import (
	"encoding/json"
	"net/http"
	"react-todos/internal/models"
)

var todos = []models.Todo{}

func getTodos(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(todos)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var todo models.Todo
	json.NewDecoder(r.Body).Decode(&todo)
	todos = append(todos, todo)
	json.NewEncoder(w).Encode(todo)
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var updatedTodo models.Todo
	json.NewDecoder(r.Body).Decode(&updatedTodo)

	for i, todo := range todos {
		if todo.ID == id {
			todos[i] = updatedTodo
			json.NewEncoder(w).Encode(updatedTodo)
			return
		}
	}

	http.NotFound(w, r)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	for i, todo := range todos {
		if todo.ID == id {
			todos = append(todos[:i], todos[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.NotFound(w, r)
}

func GetTodosHandler(w http.ResponseWriter, r *http.Request) {
	getTodos(w, r)
}

func CreateTodoHandler(w http.ResponseWriter, r *http.Request) {
	createTodo(w, r)
}

func UpdateTodoHandler(w http.ResponseWriter, r *http.Request) {
	updateTodo(w, r)
}

func DeleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	deleteTodo(w, r)
}
