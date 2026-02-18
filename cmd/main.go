package main

import (
	"context"
	"log"
	"net/http"
	"project/internal/db"
	"project/internal/handlers"
	"project/internal/middleware"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

func main() {
	r := chi.NewRouter()
	ctx := context.Background()
	_ = godotenv.Load()
	validate := validator.New()

	// Middleware

	r.Use(middleware.MiddlewareRecovery)
	r.Use(middleware.MiddlewareLogger)
	r.Use(middleware.MiddlewareContext)

	// DB Connection

	dbServer, err := db.InitDB(ctx)
	if err != nil {
		log.Fatal("Нет подключения к БД")
	}
	defer dbServer.DB.Close()

	log.Println("Подключен к БД")

	notesHandler := handlers.NotesHandler(dbServer, validate)

	// Handlers

	r.Get("/notes", notesHandler.ReadNotes)

	r.Get("/notes/{ID}", notesHandler.ReadNote)
	r.Post("/notes", notesHandler.CreateNote)
	r.Patch("/notes/{ID}", notesHandler.UpdateNote)
	r.Delete("/notes/{ID}", notesHandler.DeleteNote)

	log.Println("Сервер запущен на http://localhost:8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("Сервер словил грустного:", err)
	}
}