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
	a, err := feed.UnreadArticles(s.db)
	if err != nil {
		log.Printf("failed to fetch unread articles: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoded, _ := json.Marshal(a)
	w.Write(encoded)
}
