package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"strconv"
	"os"
	"fmt"
	"errors"
	"runtime/debug"

	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

// Variables

type DBServer struct {
	DB *pgxpool.Pool
}

var Server *DBServer
var validate = validator.New()

type Note struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// DTO Models

type CreateDTO struct {
	Title      string    `json:"title" validate:"required"`
	Content    string    `json:"content" validate:"min=1"`
}

type PatchDTO struct {
    Title   *string `json:"title" validate:"required,min=1"`
    Content *string `json:"content" validate:"omitempty,min=1"`
}

// Middleware Recovery

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			stack := debug.Stack()

			if err := recover(); err != nil {
				log.Printf("PANIC: %v\n[%s] | {%s} | [IP:PORT - %s]\nStack:\n%s\n", err, r.Method, r.URL.String(), r.RemoteAddr, stack)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Middleware Logger

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		log.Printf(
			"[%s] | {%s} | [Status: %d] %v | [IP:PORT - %s]\nUser Agent: %s", r.Method, r.URL.Path, rw.status, duration, r.RemoteAddr, r.UserAgent(),
		)
	})
}

// Middleware Context

func ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// General handler

func GetNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := Server.DB.Query(r.Context(), "SELECT id, title, content, created_at FROM notes;")
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
            http.Error(w, "Время ожидания запроса истекло", http.StatusRequestTimeout)
            log.Println("Время ожидания запроса истекло:", err)
            return
        }
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		log.Println("Ошибка БД:", err)
		return
	}
	defer rows.Close()

	notes := []Note{}

	for rows.Next() {
		note := Note{}

		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
                http.Error(w, "Время ожидания запроса истекло", http.StatusRequestTimeout)
                log.Println("Время ожидания запроса истекло:", err)
                return
            }
			http.Error(w, "Ошибка БД", http.StatusInternalServerError)
			log.Println("Ошибка сканирования:", err)
			return
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
            http.Error(w, "Время ожидания запроса истекло", http.StatusRequestTimeout)
            log.Println("Время ожидания запроса истекло:", err)
            return
        }
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		log.Println("Ошибка итерации строк:", err)
		return
	}

	json.NewEncoder(w).Encode(notes)
}

// Create

func CreateNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var input CreateDTO

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Некорретный JSON", http.StatusBadRequest)
		log.Println("Ошибка парсинга JSON:", err)
		return
	}

	err = validate.Struct(input)
	if err != nil {
		http.Error(w, "Ошибка валидации", http.StatusBadRequest)
		log.Printf("Ошибка валидации: %v\n%+v", err, input)
		return
	}

	note := Note{
		Title:     input.Title,
		Content:   input.Content,
	}

	err = Server.DB.QueryRow(r.Context(), "INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING id, created_at", note.Title, note.Content).Scan(&note.ID, &note.CreatedAt)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
            http.Error(w, "Время ожидания запроса истекло", http.StatusRequestTimeout)
            log.Println("Время ожидания запроса истекло:", err)
            return
        }
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		log.Println("Ошибка БД:", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

// Get

func GetNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := chi.URLParam(r, "ID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID не найден", http.StatusBadRequest)
		return
	}

	row := Server.DB.QueryRow(r.Context(), "SELECT id, title, content, created_at FROM notes WHERE id = $1", id)

	var note Note

	err = row.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
            http.Error(w, "Время ожидания запроса истекло", http.StatusRequestTimeout)
            log.Println("Время ожидания запроса истекло:", err)
            return
        }
		if err == pgx.ErrNoRows {
			http.Error(w, "Запись не найдена", http.StatusNotFound)
			log.Println("Запись не найдена:", err)
			return
		}
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		log.Println("Ошибка сканирования данных:", err)
		return
	}

	json.NewEncoder(w).Encode(note)
}

// Update

func UpdateNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := chi.URLParam(r, "ID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	input := PatchDTO{}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		log.Printf("Запрос с некорректным JSON: %v\n%+v", err, input)
		return
	}

	if err := validate.Struct(input); err != nil {
		http.Error(w, "Ошибка валидации", http.StatusBadRequest)
		log.Printf("Ошибка валидации: %v\n%+v", err, input)
		return
	}

	note := Note{}

	query := "UPDATE notes SET title = COALESCE($1, title), content = COALESCE($2, content) WHERE id = $3 RETURNING id, title, content, created_at"
	err = Server.DB.QueryRow(r.Context(), query, input.Title, input.Content, id).Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
            http.Error(w, "Время ожидания запроса истекло", http.StatusRequestTimeout)
            log.Println("Время ожидания запроса истекло:", err)
            return
        }
		if err == pgx.ErrNoRows {
			http.Error(w, "Не найдено", http.StatusNotFound)
			return
		}
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		log.Println("Ошибка БД:", err)
		return
	}

    w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(note)
}

// Delete

func DeleteNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := chi.URLParam(r, "ID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	cmd, err := Server.DB.Exec(r.Context(), "DELETE FROM notes WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
            http.Error(w, "Время ожидания запроса истекло", http.StatusRequestTimeout)
            log.Println("Время ожидания запроса истекло:", err)
            return
        }
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		log.Println("Ошибка БД:", err)
		return
	}

	if cmd.RowsAffected() == 0 {
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	// Connection DB & Configs

	r := chi.NewRouter()

	r.Use(RecoveryMiddleware)
	r.Use(Logger)
	r.Use(ContextMiddleware)

	ctx := context.Background()

	_ = godotenv.Load()
	conn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
)

	Server = &DBServer{}
	var err error

	Server.DB, err = pgxpool.New(ctx, conn)
	if err != nil {
		log.Fatal("Нет подключения к БД:", err)
	}
	defer Server.DB.Close()

	log.Println("Подключен к БД")

	// Handlers

	r.Get("/notes", GetNotes)

	r.Get("/notes/{ID}", GetNote)
	r.Post("/notes", CreateNote)
	r.Patch("/notes/{ID}", UpdateNote)
	r.Delete("/notes/{ID}", DeleteNote)

	// Starting

	log.Println("Сервер запущен на http://localhost:8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("Сервер словил грустного:", err)
	}
}