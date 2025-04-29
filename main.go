package feathershare

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)    // Log API calls
	r.Use(middleware.Recoverer) // Recover from panics

	// Basic test route
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "pong"}`))
	})

	// TODO: Add more routes: signup, login, upload, files list, download

	log.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
