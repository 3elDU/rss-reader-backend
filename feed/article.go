package feed

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type Article struct {
	ID int `json:"id" db:"id"`

	Subscription   *Feed `json:"-"`
	SubscriptionID int   `json:"subscription_id" db:"subscription_id"`

	URL         string `json:"url" db:"url"`
	New         bool   `json:"new" db:"new"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
	Thumbnail   []byte `json:"-"`

	// Time string in [time.DateTime] format
	Created       string    `json:"-" db:"created"`
	CreatedParsed time.Time `json:"created"`
}

// Fetch article from the database by its ID
func FindArticleByID(db *sqlx.DB, id int) (*Article, error) {
	row := db.QueryRowx(`SELECT subscription_id, url, new, title, description, thumbnail, created
		FROM articles
		WHERE articles.id = ?`,
		id,
	)

	a := &Article{}
	if err := row.StructScan(a); err != nil {
		return nil, err
	}

	a.CreatedParsed, _ = time.Parse(time.DateTime, a.Created)

	return a, nil
}

// Checks if the article with specified URL already exists in the database
func ArticleExistsInDB(db *sqlx.DB, url string) (exists bool, err error) {
	res, err := db.Query(`SELECT EXISTS(
			SELECT 1 FROM articles WHERE articles.url = ?
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

func (a Article) AddToReadLater(db *sqlx.DB) (sql.Result, error) {
	return db.Exec("INSERT INTO readlater (article_id) VALUES ($1)", a.ID)
}

// Add or update the article object in the database.
// If the article was not yet present in the database, sets the ID to the newly created row's ID
func (a *Article) Write(db *sqlx.DB) error {
	exists, err := ArticleExistsInDB(db, a.URL)
	if err != nil {
		return err
	}

	if exists {
		_, err := db.Exec(`UPDATE articles SET
			(new, title, description, thumbnail, created, url) = (?, ?, ?, ?, ?, ?)
			WHERE articles.url = ?`,
			a.New, a.Title, a.Description, a.Thumbnail, a.Created,
			a.URL,
		)
		if err != nil {
			return err
		}
	} else {
		res, err := db.Exec(`INSERT INTO articles
			(subscription_id, url, new, title, description, thumbnail, created) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			a.SubscriptionID, a.URL, a.New, a.Title, a.Description, a.Thumbnail, a.Created,
		)
		if err != nil {
			return err
		}

		id, err := res.LastInsertId()
		if err != nil {
			return err
		}

		// Set ID to the row created in the database
		a.ID = int(id)
	}

	return nil
}
