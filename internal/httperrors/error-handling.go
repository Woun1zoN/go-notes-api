package httperrors

import (
	"errors"
	"net/http"
	"context"
	"encoding/json"
	"io"
	"strconv"
	"log"

	"github.com/jackc/pgx/v5"
)

type HTTPError struct {
	Code 	int
	Message string
}

var Errors = map[string]HTTPError {
	"Timeout":         {Code: http.StatusRequestTimeout, Message: http.StatusText(http.StatusRequestTimeout)},
	"Internal":        {Code: http.StatusInternalServerError, Message: http.StatusText(http.StatusInternalServerError)},
	"BadJSON":         {Code: http.StatusBadRequest, Message: http.StatusText(http.StatusBadRequest)},
	"ErrorValidation": {Code: http.StatusBadRequest, Message: http.StatusText(http.StatusBadRequest)},
	"NotFound":        {Code: http.StatusNotFound, Message: http.StatusText(http.StatusNotFound)},
}

func HTTPErrors(w http.ResponseWriter, err error, id string) bool {
	if err == nil {
        return false
    }

	var httpErr HTTPError
	var logger string

	var syntaxErr *json.SyntaxError
    var typeErr *json.UnmarshalTypeError

	switch {
	// Timeout
	case errors.Is(err, context.DeadlineExceeded):
		httpErr = Errors["Timeout"]
		logger = "Время ожидания запроса истекло"

	// NotFound
	case errors.Is(err, pgx.ErrNoRows), errors.Is(err, ErrNoRowsAffected):
        httpErr = Errors["NotFound"]
		logger = "Не найдено"

	// BadJSON
    case errors.Is(err, io.EOF):
        httpErr = Errors["BadJSON"]
		logger = "Пустое тело запроса"

	// ErrorValidation
	case errors.Is(err, strconv.ErrSyntax), errors.Is(err, strconv.ErrRange):
        httpErr = Errors["ValidationError"]
		logger = "Недопустимое значение ввода"

	// Internal
	default:
        httpErr = Errors["Internal"]
		logger = "Внутренняя ошибка сервера"
    }

	if errors.As(err, &syntaxErr) {
        httpErr = Errors["BadJSON"]
		logger = "Недопустимый синтаксис JSON"
    } else if errors.As(err, &typeErr) {
        httpErr = Errors["BadJSON"]
		logger = "Несоответствие типов JSON"
    }

	log.Printf("Error [%s] | %s: %v", id, logger, err)

	http.Error(w, httpErr.Message, httpErr.Code)
	return true
}