// Subscription-related routes

package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/3elDU/rss-reader-backend/database"
	"github.com/3elDU/rss-reader-backend/resource"
	"github.com/mmcdole/gofeed"
)

func (s *Server) getSubscriptions(w http.ResponseWriter, r *http.Request) error {
	sms, err := s.sr.All()
	if err != nil {
		return err
	}

	srs := make([]resource.Subscription, len(sms))
	for i, sm := range sms {
		srs[i] = resource.NewSubscription(sm)
	}

	enc, _ := json.Marshal(srs)
	w.Write(enc)
	return nil
}

func (s *Server) getSingleSubscription(w http.ResponseWriter, r *http.Request) error {
	// Validate the id
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	sm, err := s.sr.Find(int64(id))
	if err != nil {
		return err
	}

	sr := resource.NewSubscription(*sm)
	enc, _ := json.Marshal(sr)
	w.Write(enc)
	return nil
}

type SubscribeRequest struct {
	URL string `json:"url" validate:"required,http_url"`
	// Optional title and description which override those from the feed
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (s *Server) subscribe(w http.ResponseWriter, r *http.Request) error {
	body := SubscribeRequest{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Printf("invalid json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	if err := s.v.Struct(&body); err != nil {
		log.Printf("validate error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	url := body.URL

	ex, id, err := s.sr.SubscriptionExists(url)
	if err != nil {
		return err
	}

	// Return early if the subscription already exists in the database
	if ex {
		w.Header().Set(
			"Location",
			fmt.Sprintf("/subscriptions/%v", id),
		)
		w.WriteHeader(http.StatusFound)
		return nil
	}

	gf, err := s.Parser.ParseURL(url)
	if err != nil {
		log.Printf("failed to fetch remote feed: %v", err)

		if err, ok := err.(gofeed.HTTPError); ok && err.StatusCode == 404 {
			http.Error(
				w,
				`{"error":true,"message":"404 when fetching a remote feed"}`,
				http.StatusBadRequest,
			)
			return nil
		} else {
			return err
		}
	}

	sr := resource.NewSubscriptionFromGofeed(*gf)

	// Some feeds return an empty url, or an invalid one
	// Overwrite the URL to the one pointing at the actual feed
	sr.Url = url

	// Override title and description with those from the request, if they are set
	if body.Title != "" {
		sr.Title = body.Title
	}
	if body.Description != "" {
		sr.Description = body.Description
	}

	sm := sr.ToModel()
	if err := s.sr.InsertSubscription(&sm); err != nil {
		return err
	}
	sr.Id = sm.ID

	articles := resource.NewArticlesFromGofeed(gf.Items, sm.ID)
	aModels := []database.Article{}
	for _, article := range articles {
		aModels = append(aModels, article.ToModel())
	}

	if err := s.ar.BulkAddArticles(aModels); err != nil {
		return err
	}

	enc, _ := json.Marshal(sr)
	w.WriteHeader(http.StatusCreated)
	w.Write(enc)
	return nil
}

func (s *Server) fetchFeedInfo(w http.ResponseWriter, r *http.Request) error {
	feedUrl := r.URL.Query().Get("url")
	if _, err := url.Parse(feedUrl); err != nil || feedUrl == "" {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	f, err := s.Parser.ParseURL(feedUrl)
	switch err.(type) {
	case gofeed.HTTPError:
		http.Error(w, "error while fetching remote feed", http.StatusBadRequest)
		return nil
	case nil:
	default:
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	res := resource.NewSubscriptionFromGofeed(*f)
	// Some feeds return an empty url, or an invalid one
	// Overwrite the URL to the one pointing at the actual feed
	res.Url = feedUrl

	// Check if feed with the specified URL already exists in the database
	if exists, id, err := s.sr.SubscriptionExists(f.Link); err == nil && exists {
		// Populate feed with it's ID in the database
		// Clients then can check, if the id != 0, then the feed already exists in the database
		res.Id = id
	} else if err != nil {
		return err
	}

	enc, _ := json.Marshal(res)
	w.Write(enc)
	return nil
}
