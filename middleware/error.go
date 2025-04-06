package middleware

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// ErrorHandler is similar to http.HandlerFunc, but it can also return an error
type ErrorHandler func(http.ResponseWriter, *http.Request) error

type ServerError struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

// Error wraps the custom handler, which is an extension of http.HandlerFunc, that can return an error.
// When the error is not nil, Error encodes it into a standard json form, as follows:
//
// { "error": true, "message": "..." }
func Error(handler ErrorHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			log.Printf("error in %v handler: %v", r.URL.Path, err)

			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			res, _ := json.Marshal(ServerError{
				Error:   true,
				Message: err.Error(),
			})

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(res)
		}
	}
}
