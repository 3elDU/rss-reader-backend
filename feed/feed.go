package feed

import (
	"github.com/jmoiron/sqlx"
)

type Feed struct {
	ID          int    `json:"id" db:"id"`
	Type        string `json:"type,omitempty" db:"type"`
	URL         string `json:"url,omitempty" db:"url"`
	Title       string `json:"title,omitempty" db:"title"`
	Description string `json:"description,omitempty" db:"description"`
	Thumbnail   []byte `db:"thumbnail" json:"-"`
}

// Read feed from the database by its ID
func FindByID(db *sqlx.DB, id int) (*Feed, error) {
	row := db.QueryRowx(`SELECT id, type, url, title, description, thumbnail
		FROM subscriptions
		WHERE subscriptions.id = ?`,
		id,
	)

	f := &Feed{}
	if err := row.StructScan(f); err != nil {
		return nil, err
	}

	return f, nil
}

// Read feed from the database by its URL
func FindByURL(db *sqlx.DB, url string) (*Feed, error) {
	row := db.QueryRowx(`SELECT id, type, url, title, description, thumbnail
		FROM subscriptions
		WHERE subscriptions.url = ?`,
		url,
	)

	f := &Feed{}
	if err := row.StructScan(f); err != nil {
		return nil, err
	}

	return f, nil
}

// Checks if the feed with specified URL already exists in the database
func ExistsInDB(db *sqlx.DB, url string) (exists bool, err error) {
	res, err := db.Query(`SELECT EXISTS(
			SELECT 1 FROM subscriptions WHERE subscriptions.url = ?
	)`, url)
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

// Add or update this feed in the database.
// If the feed was not yet present in the database, sets ID to the newly created row's ID
func (s *Feed) Write(db *sqlx.DB) error {
	exists, err := ExistsInDB(db, s.URL)
	if err != nil {
		return err
	}

	if exists {
		_, err = db.Exec(`UPDATE subscriptions SET
			(title, description, thumbnail) = (?, ?, ?)
			WHERE subscriptions.url = ?`,
			s.Title, s.Description, s.Thumbnail,
			s.URL,
		)

		return err
	} else {
		res, err := db.Exec(`INSERT INTO subscriptions
		(url, type, title, description, thumbnail) VALUES (?, ?, ?, ?, ?)`,
			s.URL, s.Type, s.Title, s.Description, s.Thumbnail,
		)
		if err != nil {
			return err
		}

		id, err := res.LastInsertId()
		if err != nil {
			return err
		}

		// Grab newly created row ID from the database
		s.ID = int(id)

		return nil
	}
}

func (s *Feed) Articles(db *sqlx.DB) (articles []Article, err error) {
	rows, err := db.Queryx(`SELECT id, url, new, title, description, thumbnail, created, readlater, created_readlater
		FROM articles
		WHERE articles.subscription_id = ?
		ORDER BY articles.created DESC`,
		s.ID,
	)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		article := Article{Subscription: s, SubscriptionID: s.ID}

		if err := rows.StructScan(&article); err != nil {
			return nil, err
		}

		articles = append(articles, article)
	}

	return
}

// Efficiently add multiple articles to the database
func (s *Feed) BulkAddArticles(db *sqlx.DB, a []Article) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO articles 
		(subscription_id, url, new, title, description, thumbnail, created, readlater)
		VALUES 
		($1, $2, TRUE, $3, $4, $5, $6, FALSE)`,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, art := range a {
		if _, err := stmt.Exec(
			s.ID,
			art.URL,
			art.Title, art.Description,
			art.Thumbnail,
			art.Created,
		); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
