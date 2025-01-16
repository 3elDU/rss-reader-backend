package middleware

import (
	"encoding/json"
	"net/http"
)

// ErrorHandler is similar to http.HandlerFunc, but it can also return an error
type ErrorHandler func(http.ResponseWriter, *http.Request) error

type ServerError struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

// Error accepts a handler different from a regular http.HandleFunc, because it can now return an error.
// When the error is not nil, Error encodes it into a standard json form, as follows:
//
// { "error": true, "message": "..." }
//
// Errors should only be returned if it's an internal server error.
// If this is a client-side error, say an invalid form, the error should be nil.
func Error(handler ErrorHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			data, _ := json.Marshal(ServerError{
				Error:   true,
				Message: err.Error(),
			})
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write(data)
		}
	}
}
