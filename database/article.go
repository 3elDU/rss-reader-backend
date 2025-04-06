package database

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// Extracted query used to query articles along with their subscriptions
const articleJoinQuery = `SELECT 
	a.*,
	s.id AS "sub.id", s.type as "sub.type", s.url as "sub.url", s.title as "sub.title", s.description as "sub.description", s.thumbnail as "sub.thumbnail"
FROM articles a INNER JOIN subscriptions s ON s.id = a.subscription_id`

type Article struct {
	ID               int64          `db:"id"`
	SubscriptionId   int64          `db:"subscription_id"`
	New              bool           `db:"new"`
	Url              string         `db:"url"`
	Title            string         `db:"title"`
	Description      sql.NullString `db:"description"`
	Thumbnail        sql.NullString `db:"thumbnail"`
	Created          sql.NullString `db:"created"`
	ReadLater        bool           `db:"readlater"`
	CreatedReadLater sql.NullString `db:"created_readlater"`
}

type ArticleWithSubscription struct {
	Article
	Subscription Subscription `db:"sub"`
}

type ArticleRepository struct {
	db *sqlx.DB
}

func NewArticleRepository(db *sqlx.DB) ArticleRepository {
	return ArticleRepository{db}
}

func (r ArticleRepository) All() (out []Article, err error) {
	rows, err := r.db.Queryx("SELECT * FROM articles")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		a := Article{}
		if err := rows.StructScan(&a); err != nil {
			return nil, err
		}
		out = append(out, a)
	}

	return
}

func (r ArticleRepository) Find(id int64) (*Article, error) {
	row := r.db.QueryRowx(
		"SELECT * FROM articles WHERE articles.id = ?",
		id,
	)

	a := &Article{}
	if err := row.StructScan(a); err != nil {
		return nil, err
	}

	return a, nil
}

func (r ArticleRepository) Unread() (out []ArticleWithSubscription, err error) {
	res, err := r.db.Queryx(articleJoinQuery + " WHERE new = TRUE")
	if err != nil {
		return nil, err
	}

	for res.Next() {
		var aws ArticleWithSubscription
		if err := res.StructScan(&aws); err != nil {
			return nil, err
		}
		out = append(out, aws)
	}

	return
}

func (r ArticleRepository) Exists(url string) (exists bool, err error) {
	res, err := r.db.Query(`SELECT EXISTS(
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

// ArticlesInSubscription fetches all the articles that belong to the subscription with the specified id.
func (r ArticleRepository) ArticlesInSubscription(id int64) ([]Article, error) {
	rows, err := r.db.Queryx(`SELECT *
		FROM articles
		WHERE articles.subscription_id = ?
		ORDER BY articles.created DESC`,
		id,
	)
	if err != nil {
		return nil, err
	}

	ais := []Article{}
	for rows.Next() {
		a := Article{}
		if err := rows.StructScan(&a); err != nil {
			return nil, err
		}
		ais = append(ais, a)
	}

	return ais, nil
}

// InReadLater returns all articles flagged as read later
func (r ArticleRepository) InReadLater() ([]ArticleWithSubscription, error) {
	rows, err := r.db.Queryx(articleJoinQuery + " WHERE readlater = TRUE")
	if err != nil {
		return nil, err
	}

	rl := []ArticleWithSubscription{}
	for rows.Next() {
		a := ArticleWithSubscription{}
		if err := rows.StructScan(&a); err != nil {
			return nil, err
		}
		rl = append(rl, a)
	}

	return rl, nil
}

// AddToReadLater adds the article to the read later list.
func (r ArticleRepository) AddToReadLater(a *Article) (err error) {
	now := time.Now().UTC().Format(time.DateTime)

	_, err = r.db.Exec(`
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
	a.CreatedReadLater = sql.NullString{
		Valid: true, String: now,
	}
	return
}

func (r ArticleRepository) RemoveFromReadLater(a *Article) (err error) {
	res, err := r.db.NamedExec(
		"UPDATE articles SET readlater = FALSE, created_readlater = NULL WHERE id = :id AND readlater = TRUE",
		a,
	)
	if err != nil {
		return
	}

	// Return an error if an article was not in the read later list in the first place
	if aff, _ := res.RowsAffected(); aff == 0 {
		return sql.ErrNoRows
	}

	return
}

func (r ArticleRepository) MarkRead(a *Article) error {
	// Also delete article from read later
	_, err := r.db.NamedExec(
		"UPDATE articles SET new = FALSE, readlater = FALSE, created_readlater = NULL WHERE id = :id",
		a,
	)
	if err != nil {
		return err
	}

	a.New = false

	return nil
}

func (r ArticleRepository) InsertArticle(a *Article) (err error) {
	res, err := r.db.NamedExec(`INSERT INTO articles
		(subscription_id, new, url, title, description, thumbnail, created, readlater, created_readlater)
		VALUES (:subscription_id, :new, :url, :title, :description, :thumbnail, :created, :readlater, :created_readlater)`,
		a,
	)
	if err != nil {
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		return
	}

	a.ID = id
	return
}

func (r ArticleRepository) UpdateArticle(db *sqlx.DB, a Article) (err error) {
	_, err = r.db.NamedExec(`UPDATE articles SET
		new = :new, url = :url, title = :title, description = :description, thumbnail = :thumbnail, created = :created, readlater = :readlater, created_readlater = :created_readlater
	WHERE articles.id = :id`,
		a,
	)
	return
}

func (r ArticleRepository) BulkAddArticles(a []Article) (err error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareNamed(`INSERT INTO articles 
		(subscription_id, new, url, title, description, thumbnail, created, readlater, created_readlater)
		VALUES 
		(:subscription_id, :new, :url, :title, :description, :thumbnail, :created, :readlater, :created_readlater)`,
	)
	if err != nil {
		tx.Rollback()
		return
	}

	for _, art := range a {
		if res, err := stmt.Exec(art); err != nil {
			tx.Rollback()
			return err
		} else {
			art.ID, _ = res.LastInsertId()
		}
	}

	return tx.Commit()
}
