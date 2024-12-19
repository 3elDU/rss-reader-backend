package feed

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type Article struct {
	ID int `json:"id" db:"id"`

	Subscription   *Feed `json:"subscription,omitempty"`
	SubscriptionID int   `json:"subscriptionId" db:"subscription_id"`

	URL         string `json:"url" db:"url"`
	New         bool   `json:"new" db:"new"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`

	Thumbnail *string `json:"thumbnail,omitempty"`

	// Time string in [time.DateTime] format
	Created string `json:"created" db:"created"`

	ReadLater bool `json:"readLater" db:"readlater"`
	// This field is nil when ReadLater is false
	AddedToReadLater *string `json:"addedToReadLater,omitempty" db:"created_readlater"`
}

func (a Article) CreatedTime() time.Time {
	t, _ := time.Parse(time.DateTime, a.Created)
	return t
}

// This function will return nil pointer, if the ReadLater field is false
func (a Article) AddedToReadLaterTime() *time.Time {
	if a.AddedToReadLater == nil {
		return nil
	}
	t, _ := time.Parse(time.DateTime, a.Created)
	return &t
}

// Fetch article from the database by its ID
func FindArticleByID(db *sqlx.DB, id int) (*Article, error) {
	row := db.QueryRowx(`SELECT subscription_id, url, new, title, description, thumbnail, created, readlater, created_readlater
		FROM articles
		WHERE articles.id = ?`,
		id,
	)

	a := &Article{ID: id}
	if err := row.StructScan(a); err != nil {
		return nil, err
	}

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

func (a *Article) AddToReadLater(db *sqlx.DB) (res sql.Result, err error) {
	now := time.Now().UTC().Format(time.DateTime)

	res, err = db.Exec(`
		UPDATE articles
		SET
			readlater = TRUE,
			created_readlater = ?
		WHERE articles.id = ?`,
		now, a.ID,
	)
	if err != nil {
		return
	}

	a.ReadLater = true
	a.AddedToReadLater = &now
	return
}

func (a *Article) MarkAsRead(db *sqlx.DB) (res sql.Result, err error) {
	res, err = db.Exec(`
		UPDATE articles
		SET new = FALSE
		WHERE articles.id = ?
	`, a.ID)
	if err != nil {
		return
	}

	a.New = false
	return
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
