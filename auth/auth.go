package auth

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	// Disable authentication entirely, allowing all requests
	NoAuth = false
)

type ContextKey string

const TokenContextKey ContextKey = "token"

type Token struct {
	ID         int       `db:"id"`
	Token      string    `db:"token"`
	CreatedAt  time.Time `db:"created_at"`
	ValidUntil time.Time `db:"valid_until"`
}

func (t Token) Expired() bool {
	return !t.ValidUntil.IsZero() && t.ValidUntil.Before(time.Now().UTC())
}

// FindTokenByString finds a token in the database and returns the token structure
func FindTokenByString(db *sqlx.DB, tokenString string) (*Token, error) {
	row := db.QueryRowx("SELECT * FROM auth_tokens WHERE auth_tokens.token = $1", tokenString)

	token := Token{}
	createdAt := ""
	var validUntil sql.NullString

	if err := row.Scan(&token.ID, &token.Token, &createdAt, &validUntil); err != nil {
		return nil, err
	}

	token.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	if validUntil.Valid {
		token.ValidUntil, _ = time.Parse(time.DateTime, validUntil.String)
	}

	return &token, nil
}
