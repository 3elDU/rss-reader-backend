package feed

import (
	"log"
	"time"

	"github.com/3elDU/rss-reader-backend/util"
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
	thumb, err := util.FetchFavicon(f.Link)
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
