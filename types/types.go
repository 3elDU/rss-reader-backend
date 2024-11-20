package types

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type Subscription struct {
	ID          int    `json:"id" db:"id"`
	Type        string `json:"type" db:"type"`
	URL         string `json:"url" db:"url" validate:"required,http_url"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description,omitempty" db:"description"`
	Thumbnail   []byte `db:"thumbnail" json:"-"`
}

// Fetch missing subscription fields from the database. Requires ID to be set
func (s *Subscription) Fetch(db *sqlx.DB) error {
	row := db.QueryRowx(`SELECT id, type, url, title, description, thumbnail
		FROM subscriptions
		WHERE subscriptions.id = ?`,
		s.ID,
	)

	return row.StructScan(s)
}

func (s Subscription) ExistsInDB(db *sqlx.DB) (exists bool, err error) {
	res, err := db.Query(`SELECT EXISTS(
			SELECT 1 FROM subscriptions WHERE subscriptions.url = ?
	)`, s.URL)
	if err != nil {
		return false, err
	}

	if !res.Next() {
		return false, res.Err()
	}

	err = res.Scan(&exists)
	res.Close()
	return
}

func (s *Subscription) Articles(db *sqlx.DB) (articles []Article, err error) {
	rows, err := db.Queryx(`SELECT id, url, new, title, description, thumbnail, created
		FROM articles
		WHERE articles.subscription_id = ?`,
		s.ID,
	)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		article := Article{Subscription: s}

		if err := rows.StructScan(&article); err != nil {
			return nil, err
		}

		article.CreatedParsed, _ = time.Parse(time.DateTime, article.Created)

		articles = append(articles, article)
	}

	return
}

type Article struct {
	ID           int           `json:"id" db:"id"`
	Subscription *Subscription `json:"-"`

	URL         string `json:"url" db:"url"`
	New         bool   `json:"new" db:"new"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
	Thumbnail   []byte `json:"-"`

	Created       string    `json:"-" db:"created"`
	CreatedParsed time.Time `json:"created"`
}

// Fetch missing fields from DB, requires ID to be set
func (a *Article) Fetch(db *sqlx.DB) error {
	row := db.QueryRowx(`SELECT id, url, new, title, description, thumbnail, created
		FROM articles
		WHERE articles.id = ?`,
		a.ID,
	)

	if err := row.StructScan(a); err != nil {
		return err
	}

	a.CreatedParsed, _ = time.Parse(time.DateTime, a.Created)

	return nil
}

func (a Article) AddToReadLater(db *sqlx.DB) (sql.Result, error) {
	return db.Exec("INSERT INTO readlater (article_id) VALUES ($1)", a.ID)
}
