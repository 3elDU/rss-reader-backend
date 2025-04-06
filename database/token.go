package database

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Token struct {
	ID         int64          `db:"id"`
	Token      string         `db:"token"`
	CreatedAt  sql.NullString `db:"created_at"`
	ValidUntil sql.NullString `db:"valid_until"`
}

type TokenRepository struct {
	db *sqlx.DB
}

func NewTokenRepository(db *sqlx.DB) TokenRepository {
	return TokenRepository{db}
}

// Find finds a token by it's value.
func (r TokenRepository) Find(tokenStr string) (*Token, error) {
	row := r.db.QueryRowx(`SELECT * FROM auth_tokens
		WHERE auth_tokens.token = ?`,
		tokenStr,
	)

	token := Token{}
	if err := row.StructScan(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

// Insert inserts a token into the database, and sets the id property on a token to the newly created row id.
func (r TokenRepository) Insert(t *Token) (err error) {
	res, err := r.db.NamedExec(`INSERT INTO auth_tokens
		(token, created_at, valid_until)
		VALUES (:token, :created_at, :valid_until)`,
		t,
	)
	if err != nil {
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		return
	}

	t.ID = id
	return
}

func (r TokenRepository) Delete(t Token) (err error) {
	_, err = r.db.NamedExec(`DELETE FROM auth_tokens WHERE id = :id OR token = :token`, t)
	return
}
