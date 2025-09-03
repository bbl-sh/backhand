package main

import (
	"log"
	"net/http"

	"github.com/john221wick/golang-backend/internal/handlers"
	"github.com/john221wick/golang-backend/internal/middleware"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Login endpoint
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")

	// Protected test endpoint
	r.Handle("/testConnection",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.TestConnectionHandler)),
	).Methods("GET")

	r.Handle("/challenge01",
		middleware.AuthMiddleware(http.HandlerFunc(handlers.TestHandler)),
	).Methods("POST")

	log.Println("Server starting on http://localhost:8080")
	log.Println("Endpoints:")
	log.Println("  POST /login - Login with PocketBase")
	log.Println("  GET  /test  - Protected endpoint with Docker")
	log.Println("  GET  /health - Health check")

	log.Fatal(http.ListenAndServe(":8080", r))
}
