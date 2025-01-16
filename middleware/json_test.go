package middleware_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/3elDU/rss-reader-backend/middleware"
	"github.com/google/go-cmp/cmp"
)

func TestJsonMiddleware(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /data",
		middleware.Json(func(w http.ResponseWriter, r *http.Request) {
			data := struct {
				A string
				B int
				C float64
			}{
				A: "hello",
				B: 42,
				C: 3.14,
			}

			enc, _ := json.Marshal(data)
			w.Write(enc)
		}),
	)

	s := httptest.NewServer(mux)
	defer s.Close()

	res, err := http.Get(s.URL + "/data")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("unable to read response body: %v", err)
	}

	want := `{"A":"hello","B":42,"C":3.14}`
	got := string(body)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("respone body mismatch (-want +got):\n%v", diff)
	}
}
