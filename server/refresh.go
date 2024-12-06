package server

import (
	"encoding/json"
	"log"
	"net/http"
)

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	new, err := s.r.Refresh()
	if err != nil {
		log.Printf("error whilst refreshing articles: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(new)
	w.Write(data)
}
