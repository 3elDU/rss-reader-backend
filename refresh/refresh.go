// refresh package provides a task that will fetch new articles from the feeds

package refresh

import (
	"log"
	"slices"
	"time"

	"github.com/3elDU/rss-reader-backend/feed"
	"github.com/jmoiron/sqlx"
	"github.com/mmcdole/gofeed"
)

// Refresh task that will run with a specified periodicity.
type Task struct {
	Ticker *time.Ticker
	db     *sqlx.DB
}

func NewTask(db *sqlx.DB, freq time.Duration) *Task {
	return &Task{
		db:     db,
		Ticker: time.NewTicker(freq),
	}
}

// Run the task scheduler. This function blocks indefinetely.
func (t *Task) Run() {
	for {
		<-t.Ticker.C

		if _, err := t.Refresh(); err != nil {
			log.Printf("feed refresh error: %v", err)
		}
	}
}

// Refresh all the feeds. This function can also be called manually.
func (t *Task) Refresh() ([]feed.Article, error) {
	feeds, err := t.collectFeeds()
	if err != nil {
		return nil, err
	}

	// Articles currently in the database
	articlesInDB, err := t.collectArticles(feeds)
	if err != nil {
		return nil, err
	}

	// Articles fetched from the feed
	articlesInFeed, err := t.collectNewArticles(feeds)
	if err != nil {
		return nil, err
	}

	newArticles := []feed.Article{}

	// Compare two slices, and add new articles to the database
	for _, a := range articlesInFeed {
		new := !slices.ContainsFunc(articlesInDB, func(el feed.Article) bool {
			return el.URL == a.URL
		})

		if new {
			if err := a.Write(t.db); err != nil {
				return nil, err
			}
			newArticles = append(newArticles, a)
		}
	}

	return newArticles, nil
}

// Grab all feeds from the database with their URLs
func (t *Task) collectFeeds() ([]feed.Feed, error) {
	rows, err := t.db.Queryx("SELECT id, url FROM subscriptions")
	if err != nil {
		return nil, err
	}

	feeds := []feed.Feed{}

	for rows.Next() {
		f := feed.Feed{}
		if err := rows.StructScan(&f); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}

	return feeds, nil
}

// Grab all articles from each feed currently in the database
func (t *Task) collectArticles(feeds []feed.Feed) (articles []feed.Article, err error) {
	for _, f := range feeds {
		fa, err := f.Articles(t.db)
		if err != nil {
			return nil, err
		}

		articles = append(articles, fa...)
	}

	return
}

// Fetch all articles from each feed via gofeed
func (t *Task) collectNewArticles(feeds []feed.Feed) (articles []feed.Article, err error) {
	for _, f := range feeds {
		gf, err := gofeed.NewParser().ParseURL(f.URL)
		if err != nil {
			return nil, err
		}

		for _, ga := range gf.Items {
			a, err := feed.NewArticleFromGofeed(ga)
			if err != nil {
				return nil, err
			}

			a.SubscriptionID = f.ID

			articles = append(articles, *a)
		}
	}

	return
}
