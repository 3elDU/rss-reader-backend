package token

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/3elDU/rss-reader-backend/database"
)

type ContextKey string

const TokenContextKey ContextKey = "token"

type Token struct {
	ID         int64
	Token      string
	CreatedAt  time.Time
	ValidUntil time.Time
}

func (t Token) Expired() bool {
	return !t.ValidUntil.IsZero() && t.ValidUntil.Before(time.Now().UTC())
}

func (t Token) ToModel() database.Token {
	vs := ""
	if !t.ValidUntil.IsZero() {
		vs = t.ValidUntil.Format(time.DateTime)
	}

	return database.Token{
		ID:    t.ID,
		Token: t.Token,
		CreatedAt: sql.NullString{
			String: t.CreatedAt.Format(time.DateTime),
			Valid:  true,
		},
		ValidUntil: sql.NullString{
			String: vs,
			Valid:  !t.ValidUntil.IsZero(),
		},
	}
}

func FromModel(t database.Token) Token {
	ca, _ := time.Parse(time.DateTime, t.CreatedAt.String)

	vu := time.Time{}
	if t.ValidUntil.Valid {
		vu, _ = time.Parse(time.DateTime, t.ValidUntil.String)
	}

	return Token{
		ID:         t.ID,
		Token:      t.Token,
		CreatedAt:  ca,
		ValidUntil: vu,
	}
}

// Generate a new token.
// A token is just 256-bit random number encoded as base64 string
func New(validFor *time.Duration) *Token {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}

	t := base64.URLEncoding.EncodeToString(buf)

	now := time.Now().UTC().Truncate(time.Second)
	validUntil := time.Time{}

	if validFor != nil && validFor.Nanoseconds() != 0 {
		added := now.Add(*validFor)
		validUntil = added
	}

	return &Token{
		Token:      t,
		CreatedAt:  now,
		ValidUntil: validUntil,
	}
}
