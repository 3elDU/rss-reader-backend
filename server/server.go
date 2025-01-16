package server

import (
	"net/http"

	"github.com/3elDU/rss-reader-backend/middleware"
	"github.com/3elDU/rss-reader-backend/refresh"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/mmcdole/gofeed"
)

type Server struct {
	*http.ServeMux

	db     *sqlx.DB
	v      *validator.Validate
	Parser *gofeed.Parser

	r *refresh.Task
}

func NewServer(db *sqlx.DB, refresher *refresh.Task) *Server {
	s := &Server{
		ServeMux: http.NewServeMux(),
		db:       db,
		v:        validator.New(),
		Parser:   gofeed.NewParser(),
		r:        refresher,
	}
	s.registerRoutes()

	return s
}

func (s *Server) registerRoutes() {
	// Route to test that the token is valid and that the backend is working properly
	s.Handle("GET /ping",
		middleware.Auth(s.db, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("pong"))
		}),
	)

	// All those routes use the same set of middlewares (auth + json response)
	routes := map[string]http.HandlerFunc{
		"GET /subscriptions/{id}":          s.getSingleSubscription,
		"GET /subscriptions":               s.getSubscriptions,
		"GET /feedinfo":                    s.fetchFeedInfo,
		"POST /subscribe":                  s.subscribe,
		"GET /subscriptions/{id}/articles": s.getArticles,
		"GET /articles/{id}":               s.getSingleArticle,
		"POST /articles/{id}/markread":     s.markArticleAsRead,
		"POST /articles/{id}/readlater":    s.addToReadLater,
		"DELETE /articles/{id}/readlater":  s.removeFromReadLater,
		"GET /readlater":                   s.showReadLater,
		"GET /unread":                      s.getUnreadArticles,
		"POST /refresh":                    s.refresh,
	}

	for p, r := range routes {
		s.Handle(p,
			middleware.Json(middleware.Auth(s.db, r)),
		)
	}
}
