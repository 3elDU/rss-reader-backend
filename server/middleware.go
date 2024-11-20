package server

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/3elDU/rss-reader-backend/auth"
	"github.com/jmoiron/sqlx"
)

func withRequestValidation(db *sqlx.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if auth.NoAuth {
			request := r.WithContext(context.WithValue(
				r.Context(),
				auth.TokenContextKey,
				auth.Token{
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

		token, err := auth.FindTokenByString(db, tokenHeader)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if err != nil {
			log.Printf("authentication: unable to find token: %v", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Check if the token is still valid
		if token.Expired() {
			log.Printf("authentication: token expired on %v", token.ValidUntil)
			db.MustExec("DELETE FROM auth_tokens WHERE auth_tokens.id = $1", token.ID)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		request := r.WithContext(context.WithValue(
			r.Context(),
			auth.TokenContextKey,
			token,
		))
		next(w, request)
	}
}
