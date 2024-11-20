// Subscription-related routes

package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/3elDU/rss-reader-backend/types"
	"github.com/jmoiron/sqlx"
	"github.com/mmcdole/gofeed"
)

func (s *Server) getSubscriptions(w http.ResponseWriter, r *http.Request) {
	subscriptions := []types.Subscription{}

	rows, err := s.db.Queryx("SELECT id, type, url, title, description FROM subscriptions")
	if err != nil {
		log.Printf("error fetching subscriptions from db: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		subscription := types.Subscription{}
		if err := rows.StructScan(&subscription); err != nil {
			log.Printf("failed to scan subscription struct: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		subscriptions = append(subscriptions, subscription)
	}

	w.Header().Add("Content-Type", "application/json")
	encoded, _ := json.Marshal(subscriptions)
	w.Write(encoded)
}

func (s *Server) getSingleSubscription(w http.ResponseWriter, r *http.Request) {
	// Validate the id
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	subscription := types.Subscription{}

	row := s.db.QueryRowx(`SELECT id, type, url, title, description
		FROM subscriptions
		WHERE subscriptions.id = $1`,
		id,
	)

	err = row.StructScan(&subscription)

	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("failed to scan subscription struct: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoded, _ := json.Marshal(subscription)
	w.Write(encoded)
}

func (s *Server) subscribe(w http.ResponseWriter, r *http.Request) {
	subscription := types.Subscription{}

	err := json.NewDecoder(r.Body).Decode(&subscription)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.validate.Struct(&subscription); err != nil {
		log.Printf("validate error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	subscription.URL = simplifyURL(subscription.URL)

	exists, err := subscription.ExistsInDB(s.db)
	if err != nil {
		log.Printf(
			"failed to check whether the subscription (%v) exists in db: %v",
			subscription.URL,
			err,
		)
	}

	// Return early if the subscription already exists in the database
	if exists {
		w.Header().Set(
			"Location",
			fmt.Sprintf("/subscriptions/%v", subscription.ID),
		)
		w.WriteHeader(http.StatusFound)
		return
	}

	parser := gofeed.NewParser()

	feed, err := parser.ParseURL(subscription.URL)
	if err != nil {
		log.Printf("feed parse error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	subscription.Type = feed.FeedType
	subscription.Title = feed.Title
	subscription.Description = feed.Description

	favicon, _ := fetchFavicon(subscription.URL)
	subscription.Thumbnail = favicon

	const query = `INSERT INTO subscriptions
		(type, url, title, description, thumbnail)
		VALUES
		(:type, :url, :title, :description, :thumbnail)`

	res, err := s.db.NamedExec(query, subscription)
	if err != nil {
		log.Printf("failed to insert into db: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set row ID in the subscription stuct
	id, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	subscription.ID = int(id)

	if err := populateArticles(s.db, subscription, feed.Items); err != nil {
		log.Printf("failed to populate articles in db: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, _ := json.Marshal(subscription)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}

// Adds articles from the feed to the database
func populateArticles(
	db *sqlx.DB,
	subscription types.Subscription,
	items []*gofeed.Item,
) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(fmt.Sprintf(`INSERT INTO articles 
		(subscription_id, new, title, description, thumbnail)
		VALUES 
		(%v, TRUE, $1, $2, $3)`,
		subscription.ID,
	))
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, article := range items {
		var thumbnail []byte
		if article.Image != nil {
			thumbnail, err = fetchImage(article.Image.URL)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		if _, err := stmt.Exec(
			article.Title, article.Description, thumbnail,
		); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
