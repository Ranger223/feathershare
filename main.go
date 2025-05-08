package main

import (
	"log"
	"main/handlers"
	"main/models"
	"main/sessionmiddleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	//Init DB
	err := models.InitDB("app.db")
	if err != nil {
		log.Fatal("Failed to init DB:", err)
	}

	//Set up api router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)    // Log API calls
	r.Use(middleware.Recoverer) // Recover from panics

	// // Basic test route
	// r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Write([]byte(`{"message": "pong"}`))
	// })
	r.Route("/api", func(r chi.Router) {
		r.Post("/signup", handlers.Signup)
		r.Post("/login", handlers.Login)

		//Secure endpoints
		r.Group(func(r chi.Router) {
			r.Use(sessionmiddleware.SessionAuth)
			r.Post("/logout", handlers.Logout)
			r.Post("/upload", handlers.UploadFile)
			r.Get("/files", handlers.ListFiles)
			r.Get("/files/download", handlers.DownloadFile)
			r.Delete("/files/delete", handlers.DeleteFiles)
			r.Get("/admin/logs", handlers.ListAllLogs)
		})
	})

	log.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
