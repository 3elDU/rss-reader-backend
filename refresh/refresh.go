// refresh package provides a task that will fetch new articles from the feeds

package refresh

import (
	"log"
	"time"

	"github.com/3elDU/rss-reader-backend/database"
	"github.com/3elDU/rss-reader-backend/resource"
	"github.com/jmoiron/sqlx"
	"github.com/mmcdole/gofeed"
)

// Refresh task that will run with a specified periodicity.
type Task struct {
	Ticker *time.Ticker
	Parser *gofeed.Parser
	db     *sqlx.DB
	sr     database.SubscriptionRepository
	ar     database.ArticleRepository
}

func NewTask(db *sqlx.DB, freq time.Duration) *Task {
	return &Task{
		Ticker: time.NewTicker(freq),
		Parser: gofeed.NewParser(),
		db:     db,
		sr:     database.NewSubscriptionRepository(db),
		ar:     database.NewArticleRepository(db),
	}
}

// Run blocks indefinetely, and runs the Refresh function in repeated intervals
func (t *Task) Run() {
	for {
		<-t.Ticker.C

		if _, err := t.Refresh(); err != nil {
			log.Printf("feed refresh error: %v", err)
		}
	}
}

// Refresh all the feeds. This function can also be called manually.
func (t *Task) Refresh() ([]resource.Article, error) {
	f, err := t.sr.All()
	if err != nil {
		return nil, err
	}

	// Articles currently in the database
	adb, err := t.ar.All()
	if err != nil {
		return nil, err
	}

	// All articles fetched from all feeds
	af, err := t.collectNewArticles(f)
	if err != nil {
		return nil, err
	}

	na := []resource.Article{}

	// Compare two slices, and add new articles to the database
	for _, anew := range af {
		new := true

		for _, aold := range adb {
			if aold.Url == anew.Url {
				new = false
				break
			}
		}

		if new {
			t.ar.InsertArticle(&anew)
			na = append(na, resource.NewArticle(anew))
		}
	}

	return na, nil
}

// Fetch all articles from each feed via gofeed, and put them all into one array
func (t *Task) collectNewArticles(feeds []database.Subscription) (out []database.Article, err error) {
	for _, f := range feeds {
		gf, err := t.Parser.ParseURL(f.Url)
		if err != nil {
			return nil, err
		}

		art := resource.NewArticlesFromGofeed(gf.Items, f.ID)
		for _, a := range art {
			out = append(out, a.ToModel())
		}
	}

	return
}
