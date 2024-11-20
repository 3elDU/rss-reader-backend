// Article-related routes

package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/3elDU/rss-reader-backend/types"
)

func (s *Server) getArticles(w http.ResponseWriter, r *http.Request) {
	// Validate the subscription id
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	subscription := types.Subscription{ID: id}
	err = subscription.Fetch(s.db)

	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("failed to fetch subscription from db: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	articles, err := subscription.Articles(s.db)
	if err != nil {
		log.Printf("failed to fetch articles from db: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoded, _ := json.Marshal(articles)
	w.Write(encoded)
}
