package main

import (
	"log"
	"net/http"
	"react-todos/internal/routes"
)

func main() {
	r := routes.SetupRoutes()

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
