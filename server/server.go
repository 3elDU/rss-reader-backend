package server

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	*http.ServeMux

	db       *sqlx.DB
	validate *validator.Validate
}

func NewServer(db *sqlx.DB) *Server {
	s := &Server{
		ServeMux: http.NewServeMux(),
		db:       db,
		validate: validator.New(),
	}
	s.registerRoutes()

	return s
}

func (s *Server) registerRoutes() {
	s.Handle("GET /subscriptions/{id}",
		withRequestValidation(s.db, s.getSingleSubscription),
	)
	s.Handle("GET /subscriptions",
		withRequestValidation(s.db, s.getSubscriptions),
	)
	s.Handle("POST /subscribe", withRequestValidation(s.db, s.subscribe))

	s.Handle("GET /subscriptions/{id}/articles",
		withRequestValidation(s.db, s.getArticles),
	)
}
