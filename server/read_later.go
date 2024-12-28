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

func (s *Server) addToReadLater(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	article, err := feed.FindArticleByID(s.db, id)

	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("error fetching article: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := article.AddToReadLater(s.db); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error adding article to read later: %v", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) removeFromReadLater(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	article, err := feed.FindArticleByID(s.db, id)

	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("error fetching article: %v", article)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.db.Exec(
		`UPDATE articles SET
			readlater = FALSE,
			created_readlater = NULL
		WHERE articles.id = ?`,
		article.ID,
	)
}

func (s *Server) showReadLater(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Queryx(`SELECT a.id, a.subscription_id, a.url, a.new, a.title, a.description, a.thumbnail, a.created, a.readlater, a.created_readlater, s.title
	FROM articles a
		INNER JOIN subscriptions s on a.subscription_id = s.id
	WHERE a.readlater = TRUE
	ORDER BY a.created_readlater DESC`)
	if err != nil {
		log.Printf("failed fetching articles in read later: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
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
			return
		}

		a.Subscription.ID = a.SubscriptionID

		articles = append(articles, a)
	}
	if err := rows.Err(); err != nil {
		log.Printf("failed fetching articles in read later: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(articles)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
