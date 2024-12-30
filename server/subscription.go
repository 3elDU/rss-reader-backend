// Subscription-related routes

package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/3elDU/rss-reader-backend/feed"
	"github.com/mmcdole/gofeed"
)

func (s *Server) getSubscriptions(w http.ResponseWriter, r *http.Request) {
	subscriptions := []feed.Feed{}

	rows, err := s.db.Queryx("SELECT id, type, url, title, description FROM subscriptions")
	if err != nil {
		log.Printf("error fetching subscriptions from db: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		subscription := feed.Feed{}
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

	sub, err := feed.FindByID(s.db, id)

	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("failed to scan subscription struct: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoded, _ := json.Marshal(sub)
	w.Write(encoded)
}

type SubscribeRequest struct {
	URL string `json:"url" validate:"required,http_url"`
	// Optional title and description which override those from the feed
	Title       string
	Description string
}

func (s *Server) subscribe(w http.ResponseWriter, r *http.Request) {
	body := SubscribeRequest{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Printf("invalid json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.v.Struct(&body); err != nil {
		log.Printf("validate error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	url := body.URL

	exists, err := feed.ExistsInDB(s.db, url)
	if err != nil {
		log.Printf(
			"failed to check whether the subscription (%v) exists in db: %v",
			url,
			err,
		)
	}

	// Return early if the subscription already exists in the database
	if exists {
		f, err := feed.FindByURL(s.db, url)
		if err != nil {
			log.Printf("failed to fetch existing feed from db: %v", err)
		}

		w.Header().Set(
			"Location",
			fmt.Sprintf("/subscriptions/%v", f.ID),
		)
		w.WriteHeader(http.StatusFound)
		return
	}

	sub, articles, err := feed.FetchRemote(s.Parser, url)
	if err != nil {
		log.Printf("failed to fetch remote feed: %v", err)

		if err, ok := err.(gofeed.HTTPError); ok && err.StatusCode == 404 {
			http.Error(w, "404 when fetching remote feed", http.StatusBadRequest)
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Override title and description with those from the request, if they are set
	if body.Title != "" {
		sub.Title = body.Title
	}
	if body.Description != "" {
		sub.Description = body.Description
	}

	if err := sub.Write(s.db); err != nil {
		log.Printf("failed to insert into db: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := sub.BulkAddArticles(s.db, articles); err != nil {
		log.Printf("failed to populate articles in db: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, _ := json.Marshal(sub)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(j)
}

func (s *Server) fetchFeedInfo(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Query().Get("url")
	if _, err := url.Parse(u); err != nil || u == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	f, err := feed.FetchRemoteFeed(s.Parser, u)
	switch err.(type) {
	case gofeed.HTTPError:
		http.Error(w, "error while fetching remote feed", http.StatusBadRequest)
		return
	case nil:
	default:
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Check if feed with the specified URL already exists in the database
	if exists, err := feed.ExistsInDB(s.db, u); err == nil && exists {
		var id int
		if err := s.db.Get(&id,
			"SELECT id FROM subscriptions s WHERE s.url = ?", u,
		); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Populate feed with it's ID in the database
		// Clients then can check, if the id != 0, then the feed already exists in the database
		f.ID = id
	} else if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(f)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
