package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/3elDU/rss-reader-backend/middleware"
	"github.com/3elDU/rss-reader-backend/resource"
)

func (s *Server) addToReadLater(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	a, err := s.ar.Find(int64(id))
	if err != nil {
		return err
	}

	if err := s.ar.AddToReadLater(a); err != nil {
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

	a, err := s.ar.Find(int64(id))
	if err != nil {
		return err
	}

	if err := s.ar.RemoveFromReadLater(a); err == nil {
		w.WriteHeader(http.StatusNoContent)
	}

	return nil
}

func (s *Server) showReadLater(w http.ResponseWriter, r *http.Request) error {
	pg := middleware.GetPagination(r.Context())

	arl, err := s.ar.InReadLater(pg.Limit, pg.Offset)
	if err != nil {
		return err
	}

	ars := make([]resource.ArticleWithSubscription, len(arl))
	for i, arl := range arl {
		ars[i] = resource.NewArticleWithSubscription(arl)
	}

	enc, _ := json.Marshal(ars)
	w.Write(enc)
	return nil
}
