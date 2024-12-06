package server

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/3elDU/rss-reader-backend/token"
	"github.com/jmoiron/sqlx"
)

var (
	// Disable authentication entirely, allowing all requests
	NoAuth = false
)

func withRequestValidation(db *sqlx.DB, next http.HandlerFunc) http.HandlerFunc {
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

		tokenHeader := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if tokenHeader == "" {
			log.Printf("authentication: unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		t, err := token.FindByString(db.DB, tokenHeader)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if err != nil {
			log.Printf("authentication: unable to find token: %v", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Check if the token is still valid
		if t.Expired() {
			log.Printf("authentication: token expired on %v", t.ValidUntil)
			db.MustExec("DELETE FROM auth_tokens WHERE auth_tokens.id = $1", t.ID)
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
