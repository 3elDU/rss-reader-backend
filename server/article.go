// Article-related routes

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

func (s *Server) getArticles(w http.ResponseWriter, r *http.Request) {
	// Validate the subscription id
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sub, err := feed.FindByID(s.db, id)

	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("failed to fetch subscription from db: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	articles, err := sub.Articles(s.db)
	if err != nil {
		log.Printf("failed to fetch articles from db: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoded, _ := json.Marshal(articles)
	w.Write(encoded)
}

func (s *Server) getSingleArticle(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
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

	w.Header().Set("Content-Type", "application/json")
	encoded, _ := json.Marshal(article)
	w.Write(encoded)
}

func (s *Server) getUnreadArticles(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Queryx(`SELECT a.id, a.subscription_id, a.url, a.new, a.title, a.description, a.thumbnail, a.created, a.readlater, a.created_readlater, s.title
		FROM articles a
			INNER JOIN subscriptions s ON s.id = a.subscription_id
			WHERE new = TRUE
		ORDER BY a.created DESC
	`)
	if err != nil {
		log.Printf("failed to query db: %v", err)
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
			log.Printf("failed to scan article: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		a.Subscription.ID = a.SubscriptionID

		articles = append(articles, a)
	}

	w.Header().Set("Content-Type", "application/json")
	encoded, _ := json.Marshal(articles)
	w.Write(encoded)
}

func (s *Server) markArticleAsRead(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	article, err := feed.FindArticleByID(s.db, id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if _, err := article.MarkAsRead(s.db); err != nil {
		log.Printf("failed to mark article as read: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
