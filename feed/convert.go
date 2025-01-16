package feed

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/mmcdole/gofeed"
)

// Converts gofeed.Item into feed.Article.
// Does not add the article to the database, so ID is always zero
func NewArticleFromGofeed(a *gofeed.Item) (*Article, error) {
	var thumb *string
	if a.Image != nil {
		thumb = &a.Image.URL
	}

	var t time.Time
	if a.PublishedParsed != nil {
		t = a.PublishedParsed.UTC()
	} else {
		// Failsafe, when the published date is nil
		t = time.Now().UTC()
		log.Printf("gofeed article has PublishedParsed = nil: %v", a)
	}
	c := t.Format(time.DateTime)

	art := &Article{
		URL:         a.Link,
		New:         true,
		Title:       a.Title,
		Description: a.Description,
		Thumbnail:   thumb,
		Created:     c,
	}

	return art, nil
}

// Converts gofeed.Feed into feed.Feed
// Does not add the feed to the database, so ID is always zero
func NewFeedFromGofeed(f *gofeed.Feed) (*Feed, error) {
	thumb, err := fetchFavicon(f.Link)
	if err != nil {
		return nil, err
	}

	feed := &Feed{
		Type:        f.FeedType,
		URL:         f.FeedLink,
		Title:       f.Title,
		Description: f.Description,
		Thumbnail:   thumb,
	}
	return feed, nil
}

// Fetch the website's favicon using google favicon cache, returns the image blob
func fetchFavicon(website string) ([]byte, error) {
	u := fmt.Sprintf(
		"https://www.google.com/s2/favicons?sz=64&domain=%v",
		url.QueryEscape(website),
	)
	res, err := http.DefaultClient.Get(u)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(res.Body)
}
