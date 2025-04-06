package middleware

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/3elDU/rss-reader-backend/database"
	"github.com/3elDU/rss-reader-backend/token"
)

var (
	// Disable authentication entirely, allowing all requests
	NoAuth = false
)

// Auth checks the token in the Authorization header against the provided database.
// The token is then provided in the context value
// If `NoAuth` is false - all checks are skipped and the dummy token is set in the context
func Auth(repo database.TokenRepository, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if NoAuth {
			request := r.WithContext(context.WithValue(
				r.Context(),
				token.TokenContextKey,
				token.Token{
					ID:        0,
					Token:     "dummy",
					CreatedAt: time.Now(),
				},
			))
			next(w, request)
			return
		}

		tokenStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if tokenStr == "" {
			log.Printf("authentication: unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tm, err := repo.Find(tokenStr)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if err != nil {
			log.Printf("authentication: unable to find token: %v", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		t := token.FromModel(*tm)

		// Check if the token is still valid
		if t.Expired() {
			log.Printf("authentication: token expired on %v", t.ValidUntil)
			repo.Delete(*tm)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		request := r.WithContext(context.WithValue(
			r.Context(),
			token.TokenContextKey,
			t,
		))
		next(w, request)
	}
}
