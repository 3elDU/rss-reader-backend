package database

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type Subscription struct {
	ID          int64          `db:"id"`
	Type        string         `db:"type"`
	Url         string         `db:"url"`
	Title       string         `db:"title"`
	Description sql.NullString `db:"description"`
	Thumbnail   sql.NullString `db:"thumbnail"`
}

type SubscriptionRepository struct {
	db *sqlx.DB
}

func NewSubscriptionRepository(db *sqlx.DB) SubscriptionRepository {
	return SubscriptionRepository{db}
}

func (r SubscriptionRepository) All() ([]Subscription, error) {
	rows, err := r.db.Queryx("SELECT * FROM subscriptions")
	if err != nil {
		return nil, err
	}

	s := []Subscription{}
	for rows.Next() {
		sub := Subscription{}
		if err := rows.StructScan(&sub); err != nil {
			return nil, err
		}
		s = append(s, sub)
	}

	return s, nil
}

func (r SubscriptionRepository) Find(id int64) (*Subscription, error) {
	row := r.db.QueryRowx(`SELECT *
		FROM subscriptions
		WHERE subscriptions.id = ?`,
		id,
	)

	f := &Subscription{}
	if err := row.StructScan(f); err != nil {
		return nil, err
	}

	return f, nil
}

func (r SubscriptionRepository) FindByUrl(url string) (*Subscription, error) {
	row := r.db.QueryRowx(`SELECT *
		FROM subscriptions
		WHERE subscriptions.url = ?`,
		url,
	)

	f := &Subscription{}
	if err := row.StructScan(f); err != nil {
		return nil, err
	}

	return f, nil
}

func (r SubscriptionRepository) SubscriptionExists(url string) (exists bool, id int64, err error) {
	res, err := r.db.Query("SELECT id FROM subscriptions WHERE subscriptions.url = ?", url)
	if errors.Is(err, sql.ErrNoRows) {
		return false, 0, nil
	} else if err != nil {
		return
	}

	if !res.Next() {
		err = res.Err()
		return
	}
	err = res.Scan(&id)
	if err != nil {
		return
	}

	return true, id, nil
}

// InsertSubscription inserts the given structure into the database, and sets the ID property on the Subscription.
func (r SubscriptionRepository) InsertSubscription(s *Subscription) (err error) {
	res, err := r.db.NamedExec(`INSERT INTO subscriptions
		(type, url, title, description, thumbnail)
		VALUES (:type, :url, :title, :description, :thumbnail)`,
		s,
	)
	if err != nil {
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		return
	}

	s.ID = id
	return
}

func (r SubscriptionRepository) UpdateSubscription(s Subscription) (err error) {
	_, err = r.db.NamedExec(`UPDATE subscriptions SET
		type = :type, url = :url, title = :title, description = :description, thumbnail = :thumbnail
	WHERE subscriptions.id = :id`,
		s,
	)
	return
}
