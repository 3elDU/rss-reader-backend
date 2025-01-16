package middleware

import "net/http"

// Json sets the "Content-Type" header of the response to "application/json; charset=utf-8"
func Json(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next(w, r)
	}
}
