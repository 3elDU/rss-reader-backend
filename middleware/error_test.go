package middleware_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/3elDU/rss-reader-backend/middleware"
	"github.com/google/go-cmp/cmp"
)

func TestErrorMiddleware(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/noerror",
		middleware.Error(func(w http.ResponseWriter, r *http.Request) error {
			w.Write([]byte("hello"))
			return nil
		}),
	)
	mux.HandleFunc("/error",
		middleware.Error(func(w http.ResponseWriter, r *http.Request) error {
			return errors.New("test error")
		}),
	)

	s := httptest.NewServer(mux)
	defer s.Close()

	tests := map[string]struct {
		Route       string
		ContentType string
		Want        string
	}{
		"no error": {
			"/noerror",
			"text/plain; charset=utf-8",
			"hello",
		},
		"error": {
			"/error",
			"application/json; charset=utf-8",
			`{"error":true,"message":"test error"}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := http.Get(s.URL + test.Route)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}

			if ct := res.Header.Get("Content-Type"); ct != test.ContentType {
				t.Errorf("content-type mismatch: want '%v' got '%v'",
					test.ContentType, ct,
				)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			if diff := cmp.Diff(test.Want, string(body)); diff != "" {
				t.Errorf("response mismatch (-want +got):\n%v", diff)
			}
		})
	}
}
