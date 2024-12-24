package server_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type mockResponse struct {
	statusCode int
	body       string
}

// RoundTripper implementation that mocks the requests to get RSS feeds
type mockRoundTripper struct {
	t *testing.T

	// A map with URLs to mock, where key is the URL and value is the body to respond with
	MockedURLs map[string]mockResponse
}

func (rt *mockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, ok := rt.MockedURLs[r.URL.String()]
	if !ok {
		return nil, fmt.Errorf("url not in list of mocked urls: %v", r.URL)
	}

	return &http.Response{
		StatusCode:    resp.statusCode,
		Body:          io.NopCloser(strings.NewReader(resp.body)),
		ContentLength: int64(len(resp.body)),
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Request:       r,
		Header:        make(http.Header),
	}, nil
}

func TestSubscriptionRoutes(t *testing.T) {
	ServerStruct.Parser.Client = &http.Client{
		Transport: &mockRoundTripper{t,
			map[string]mockResponse{
				"https://example.com/rss.xml": {200, `<?xml version="1.0" encoding="UTF-8"?>
					<rss version="2.0">
						<channel>
							<title>Test Feed</title>
							<link>https://example.com</link>
							<description>Test feed for testing</description>
							<item>
								<title>Test Article</title>
								<pubDate>
									Tue, 24 Dec 2024 00:00:00 +0000
								</pubDate>
								<link>https://example.com/test-article</link>
								<description>Test article description</description>
							</item>
						</channel>
					</rss>`,
				},
				"https://example.com/404.xml": {404, "404 not found"},
			},
		},
	}

	tests := []struct {
		name   string
		method string
		path   string
		// Used only when method == "POST"
		postBody io.Reader
		// Status code to expect
		statusCode int
		// Response body to expect
		responseBody string
	}{
		{
			"subscribe to feed",
			"POST",
			"/subscribe",
			strings.NewReader(`{"url": "https://example.com/rss.xml"}`),
			http.StatusCreated,
			`{"id":1,"type":"rss","title":"Test Feed","description":"Test feed for testing"}`,
		},
		{
			"get feed by id",
			"GET",
			"/subscriptions/1",
			nil,
			200,
			`{"id":1,"type":"rss","title":"Test Feed","description":"Test feed for testing"}`,
		},
		{
			"get all feeds",
			"GET",
			"/subscriptions",
			nil,
			200,
			`[{"id":1,"type":"rss","title":"Test Feed","description":"Test feed for testing"}]`,
		},
		{
			"get subscriptions articles",
			"GET",
			"/subscriptions/1/articles",
			nil,
			200,
			`[{"id":1,"subscription":{"id":1,"type":"rss","title":"Test Feed","description":"Test feed for testing"},"subscriptionId":1,"url":"https://example.com/test-article","new":true,"title":"Test Article","description":"Test article description","created":"2024-12-24 00:00:00","readLater":false}]`,
		},
		{
			"proper 404 handling",
			"POST",
			"/subscribe",
			strings.NewReader(`{"url": "https://example.com/404.xml"}`),
			400,
			"404 when fetching remote feed\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var resp *http.Response
			var err error

			if test.method == "GET" {
				resp, err = http.Get(TestServer.URL + test.path)
			} else if test.method == "POST" {
				resp, err = http.Post(
					TestServer.URL+test.path,
					"application/json",
					test.postBody,
				)
			}
			if err != nil {
				t.Fatalf("http request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != test.statusCode {
				t.Errorf(
					"bad http status code: want %v, got %v",
					test.statusCode, resp.StatusCode,
				)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			if diff := cmp.Diff(test.responseBody, string(body)); diff != "" {
				t.Errorf("unexpected response body (-want +got):\n%v", diff)
			}
		})
	}
}
