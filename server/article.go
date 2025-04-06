// Article-related routes

package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/3elDU/rss-reader-backend/resource"
)

func (s *Server) getArticles(w http.ResponseWriter, r *http.Request) error {
	// Validate the subscription id
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	adb, err := s.ar.ArticlesInSubscription(int64(id))
	if err != nil {
		return err
	}

	sub, err := s.sr.Find(int64(id))
	if err != nil {
		return err
	}

	a := make([]resource.ArticleWithSubscription, len(adb))
	for i, adb := range adb {
		a[i] = resource.NewArticleWithSubscriptionTwopart(adb, *sub)
	}

	enc, _ := json.Marshal(a)
	w.Write(enc)
	return nil
}

func (s *Server) getSingleArticle(w http.ResponseWriter, r *http.Request) error {
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	am, err := s.ar.Find(int64(id))
	if err != nil {
		return err
	}

	sm, err := s.sr.Find(am.SubscriptionId)
	if err != nil {
		return err
	}

	res := resource.NewArticleWithSubscriptionTwopart(*am, *sm)

	encoded, _ := json.Marshal(res)
	w.Write(encoded)
	return nil
}

func (s *Server) getUnreadArticles(w http.ResponseWriter, r *http.Request) error {
	unr, err := s.ar.Unread()
	if err != nil {
		return err
	}

	ars := make([]resource.ArticleWithSubscription, len(unr))
	for i, m := range unr {
		ars[i] = resource.NewArticleWithSubscription(m)
	}

	encoded, _ := json.Marshal(ars)
	w.Write(encoded)
	return nil
}

func (s *Server) markArticleAsRead(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	article, err := s.ar.Find(int64(id))
	if err != nil {
		return nil
	}

	if err := s.ar.MarkRead(article); err != nil {
		return err
	}

	return nil
}
