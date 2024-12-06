package token

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"time"
)

type ContextKey string

const TokenContextKey ContextKey = "token"

type Token struct {
	ID         int        `db:"id"`
	Token      string     `db:"token"`
	CreatedAt  time.Time  `db:"created_at"`
	ValidUntil *time.Time `db:"valid_until"`
}

func (t Token) Expired() bool {
	return !t.ValidUntil.IsZero() && t.ValidUntil.Before(time.Now().UTC())
}

// Write token to the database.
// Sets its ID property to the row created in the database.
func (t *Token) Write(db *sql.DB) (err error) {
	res, err := db.Exec(
		`INSERT INTO auth_tokens (token, created_at, valid_until)
		VALUES ($1, $2, $3)`,
		t.Token, t.CreatedAt, t.ValidUntil,
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	t.ID = int(id)

	return
}

// Finds a token in the database by its value
func FindByString(db *sql.DB, tokenString string) (*Token, error) {
	row := db.QueryRow("SELECT * FROM auth_tokens WHERE auth_tokens.token = $1", tokenString)

	token := Token{}
	createdAt := ""
	var validUntil sql.NullString

	if err := row.Scan(&token.ID, &token.Token, &createdAt, &validUntil); err != nil {
		return nil, err
	}

	token.CreatedAt, _ = time.Parse(time.DateTime, createdAt)
	if validUntil.Valid {
		t, _ := time.Parse(time.DateTime, validUntil.String)
		token.ValidUntil = &t
	}

	return &token, nil
}

// Generate a new token.
// A token is just 256-bit random number encoded as base64 string
func New(validFor *time.Duration) *Token {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}

	t := base64.URLEncoding.EncodeToString(buf)

	now := time.Now().UTC()
	var validUntil *time.Time
	if validFor != nil {
		added := now.Add(*validFor)
		validUntil = &added
	}

	return &Token{
		Token:      t,
		CreatedAt:  now,
		ValidUntil: validUntil,
	}
}
