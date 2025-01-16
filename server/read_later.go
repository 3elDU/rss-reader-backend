package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/3elDU/rss-reader-backend/feed"
)

func (s *Server) addToReadLater(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	article, err := feed.FindArticleByID(s.db, id)

	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		return nil
	} else if err != nil {
		log.Printf("error fetching article: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	if _, err := article.AddToReadLater(s.db); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error adding article to read later: %v", err)
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) removeFromReadLater(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	article, err := feed.FindArticleByID(s.db, id)

	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		return nil
	} else if err != nil {
		log.Printf("error fetching article: %v", article)
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	_, err = s.db.Exec(
		`UPDATE articles SET
			readlater = FALSE,
			created_readlater = NULL
		WHERE articles.id = ?`,
		article.ID,
	)
	return err
}

func (s *Server) showReadLater(w http.ResponseWriter, r *http.Request) error {
	rows, err := s.db.Queryx(`SELECT a.id, a.subscription_id, a.url, a.new, a.title, a.description, a.thumbnail, a.created, a.readlater, a.created_readlater, s.title
	FROM articles a
		INNER JOIN subscriptions s on a.subscription_id = s.id
	WHERE a.readlater = TRUE
	ORDER BY a.created_readlater DESC`)
	if err != nil {
		log.Printf("failed fetching articles in read later: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	articles := []feed.Article{}
	for rows.Next() {
		a := feed.Article{
			Subscription: &feed.Feed{},
		}

		if err := rows.Scan(
			&a.ID,
			&a.SubscriptionID,
			&a.URL,
			&a.New,
			&a.Title,
			&a.Description,
			&a.Thumbnail,
			&a.Created,
			&a.ReadLater,
			&a.AddedToReadLater,
			&a.Subscription.Title,
		); err != nil {
			log.Printf("failed fetching articles in read later: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

		a.Subscription.ID = a.SubscriptionID

		articles = append(articles, a)
	}
	if err := rows.Err(); err != nil {
		log.Printf("failed fetching articles in read later: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	data, _ := json.Marshal(articles)
	w.Write(data)
	return nil
}
