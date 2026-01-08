package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type Todo struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Assigned    string `json:"assigned"`
}

type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

var todos = []Todo{
	{ID: 1, Description: "Deploy to Vercel", Assigned: "You"},
	{ID: 2, Description: "Setup Database", Assigned: "Dev"},
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		response := APIResponse{
			Success:   true,
			Data:      todos,
			Timestamp: time.Now(),
		}
		json.NewEncoder(w).Encode(response)
	case "POST":
		var newTodo Todo
		json.NewDecoder(r.Body).Decode(&newTodo)
		newTodo.ID = len(todos) + 1
		todos = append(todos, newTodo)
		
		w.WriteHeader(http.StatusCreated)
		response := APIResponse{
			Success:   true,
			Data:      map[string]string{"message": "Todo created"},
			Timestamp: time.Now(),
		}
		json.NewEncoder(w).Encode(response)
	}
}