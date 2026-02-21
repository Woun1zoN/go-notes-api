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
	// Variables

	r := chi.NewRouter()
	ctx := context.Background()
	validate := validator.New()
	godotenv.Load("../.env")

	// Middleware

	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(middleware.Context)

	// DB Connection

	dbServer, err := db.InitDB(ctx)
	if err != nil {
		log.Fatal("Нет подключения к БД:", err)
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

	// Starting

	log.Println("Сервер запущен на http://localhost:8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("Сервер словил грустного:", err)
	}
}