package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"project/internal/db"
	"project/internal/middleware"
	"project/internal/models"
	"project/internal/httperrors"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
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
	if httperrors.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
	}
	defer rows.Close()

	notes := []models.Note{}

	for rows.Next() {
		note := models.Note{}

		err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt)
		if httperrors.HTTPErrors(w, err, middleware.GetRequestID(r)) {
            return
        }

		notes = append(notes, note)
	}

	if err := rows.Err(); httperrors.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
    }

	json.NewEncoder(w).Encode(notes)
}

// Create Note

func (Server *NotesHandlerDB) CreateNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var input models.CreateDTO

	err := json.NewDecoder(r.Body).Decode(&input)
	if httperrors.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
    }
	defer r.Body.Close()

	err = Server.Validate.Struct(input)
	if httperrors.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
    }

	note := models.Note{
		Title:     input.Title,
		Content:   input.Content,
	}

	err = Server.DB.QueryRow(r.Context(), "INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING id, created_at", note.Title, note.Content).Scan(&note.ID, &note.CreatedAt)
	if httperrors.HTTPErrors(w, err, middleware.GetRequestID(r)) {
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
	if httperrors.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
    }

	row := Server.DB.QueryRow(r.Context(), "SELECT id, title, content, created_at FROM notes WHERE id = $1", id)

	var note models.Note

	err = row.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt)
	if httperrors.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
    }

	json.NewEncoder(w).Encode(note)
}

// Update Note

func (Server *NotesHandlerDB) UpdateNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := chi.URLParam(r, "ID")
	id, err := strconv.Atoi(idStr)
	if httperrors.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
    }

	input := models.UpdateDTO{}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&input); httperrors.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
    }

	if err := Server.Validate.Struct(input); httperrors.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
    }

	note := models.Note{}

	query := "UPDATE notes SET title = COALESCE($1, title), content = COALESCE($2, content) WHERE id = $3 RETURNING id, title, content, created_at"
	err = Server.DB.QueryRow(r.Context(), query, input.Title, input.Content, id).Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt)
	if httperrors.HTTPErrors(w, err, middleware.GetRequestID(r)) {
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
	if httperrors.HTTPErrors(w, err, middleware.GetRequestID(r)) {
        return
    }

	if cmd.RowsAffected() == 0 {
        httperrors.HTTPErrors(w, httperrors.ErrNoRowsAffected, middleware.GetRequestID(r))
        return
}

	w.WriteHeader(http.StatusNoContent)
}