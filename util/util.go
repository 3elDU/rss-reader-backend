package util

import (
	"fmt"
	"io"
	"net/http"
	urllib "net/url"

	_ "image/png"
)

// Fetch the website's favicon using google favicon cache, returns the image blob
func FetchFavicon(website string) ([]byte, error) {
	url := fmt.Sprintf(
		"https://www.google.com/s2/favicons?sz=64&domain=%v",
		urllib.QueryEscape(website),
	)
	res, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(res.Body)
}

func FetchImage(img string) ([]byte, error) {
	res, err := http.DefaultClient.Get(img)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(res.Body)
}

// Replaces scheme to 'https', and strips ?query and #fragment
func SimplifyURL(orig string) string {
	url, _ := urllib.Parse(orig)

	url.Scheme = "https"
	url.RawQuery = ""
	url.RawFragment = ""

	return url.String()
}
