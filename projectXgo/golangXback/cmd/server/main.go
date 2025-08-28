package main

import (
	"log"
	"net/http"
	"time"

	"your-app/internal/handlers"
	pbstart "your-app/internal/pb"

	"github.com/go-chi/chi/v5"
)

func main() {
	// Start embedded PocketBase and get SDK client.
	pbClient := pbstart.StartEmbedded()
	log.Println("PocketBase embedded started (http://127.0.0.1:8090)")

	// Create app handlers with PB SDK client
	app := handlers.NewApp(pbClient)

	r := chi.NewRouter()
	r.Post("/login", app.LoginHandler)
	r.Get("/test-connection", app.TestConnectionHandler)
	r.Post("/challenge01", app.TestHandler)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("Server running at http://127.0.0.1:8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
