package handlers

import (
    "net/http"
	"errors"
	"log"
	"context"
	"encoding/json"
	"strconv"
	"project/internal/models"
	"project/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"github.com/go-playground/validator/v10"
	"github.com/go-chi/chi"
)

type NotesHandlerDB struct {
	DB *pgxpool.Pool
	Validate *validator.Validate
}

func NotesHandler(dbServer *db.DBServer, validate *validator.Validate) *NotesHandlerDB {
	return &NotesHandlerDB{
		DB: dbServer.DB,
		Validate: validate,
	}
}

// Read All Notes

func (Server *NotesHandlerDB) ReadNotes(w http.ResponseWriter, r *http.Request) {
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

	notes := []models.Note{}

	for rows.Next() {
		note := models.Note{}

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

// Create Note

func (Server *NotesHandlerDB) CreateNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var input models.CreateDTO

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Некорретный JSON", http.StatusBadRequest)
		log.Println("Ошибка парсинга JSON:", err)
		return
	}
	defer r.Body.Close()

	err = Server.Validate.Struct(input)
	if err != nil {
		http.Error(w, "Ошибка валидации", http.StatusBadRequest)
		log.Printf("Ошибка валидации: %v\n%+v", err, input)
		return
	}

	note := models.Note{
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

// Read Note

func (Server *NotesHandlerDB) ReadNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := chi.URLParam(r, "ID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID не найден", http.StatusBadRequest)
		return
	}

	row := Server.DB.QueryRow(r.Context(), "SELECT id, title, content, created_at FROM notes WHERE id = $1", id)

	var note models.Note

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

// Update Note

func (Server *NotesHandlerDB) UpdateNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := chi.URLParam(r, "ID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	input := models.UpdateDTO{}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		log.Printf("Запрос с некорректным JSON: %v\n%+v", err, input)
		return
	}

	if err := Server.Validate.Struct(input); err != nil {
		http.Error(w, "Ошибка валидации", http.StatusBadRequest)
		log.Printf("Ошибка валидации: %v\n%+v", err, input)
		return
	}

	note := models.Note{}

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

// Delete Note

func (Server *NotesHandlerDB) DeleteNote(w http.ResponseWriter, r *http.Request) {
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