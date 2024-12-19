package server

import (
	"net/http"

	"github.com/3elDU/rss-reader-backend/refresh"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	*http.ServeMux

	db       *sqlx.DB
	validate *validator.Validate

	r *refresh.Task
}

func NewServer(db *sqlx.DB, refresher *refresh.Task) *Server {
	s := &Server{
		ServeMux: http.NewServeMux(),
		db:       db,
		validate: validator.New(),
		r:        refresher,
	}
	s.registerRoutes()

	return s
}

func (s *Server) registerRoutes() {
	// Route to test that the token is valid and that the backend is working properly
	s.Handle("GET /ping",
		withRequestValidation(s.db, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("pong"))
		}),
	)

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

	s.Handle("GET /articles/{id}",
		withRequestValidation(s.db, s.getSingleArticle),
	)
	s.Handle("POST /articles/{id}/markread",
		withRequestValidation(s.db, s.markArticleAsRead),
	)
	s.Handle("POST /articles/{id}/readlater",
		withRequestValidation(s.db, s.addToReadLater),
	)
	s.Handle("DELETE /articles/{id}/readlater",
		withRequestValidation(s.db, s.removeFromReadLater),
	)
	s.Handle("GET /readlater",
		withRequestValidation(s.db, s.showReadLater),
	)

	s.Handle("GET /unread",
		withRequestValidation(s.db, s.getUnreadArticles),
	)

	s.Handle("POST /refresh",
		withRequestValidation(s.db, s.refresh),
	)
}
