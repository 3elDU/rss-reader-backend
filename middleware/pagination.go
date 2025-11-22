package middleware

import (
	"context"
	"net/http"
	"strconv"
)

type ContextKey string

const (
	// DefaultLimit is the default count of items per page
	DefaultLimit             = 20
	PaginationKey ContextKey = "pagination"
)

type PaginationData struct {
	// Offset represents the number of items to skip, equivalent to OFFSET in SQL
	Offset int
	// Limit represents the maximum number of items to return, equivalent to LIMIT in SQL
	Limit int
}

// Pagination takes the optional query parameters `page` and `limit` from the request,
// and exposes them as `Page` and `Limit` fields in the context.
func WithPagination(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Query().Get("page")
		l := r.URL.Query().Get("limit")

		pn, err := strconv.Atoi(p)
		if err != nil || pn < 1 {
			pn = 1
		}
		ln, err := strconv.Atoi(l)
		if err != nil || ln < 1 {
			ln = DefaultLimit
		}

		// Calculate the offset
		off := (pn - 1) * ln

		r = r.WithContext(context.WithValue(r.Context(), PaginationKey, PaginationData{
			Offset: off,
			Limit:  ln,
		}))

		next.ServeHTTP(w, r)
	})
}

// GetPagination gets the PaginationData struct from the request, falling
// back to the default value, if PaginationData does not exist in the context
func GetPagination(ctx context.Context) PaginationData {
	pg, ok := ctx.Value(PaginationKey).(PaginationData)
	if ok {
		return pg
	} else {
		return PaginationData{
			Offset: 0,
			Limit:  DefaultLimit,
		}
	}
}
