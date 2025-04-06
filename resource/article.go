package resource

import (
	"database/sql"
	"time"

	"github.com/3elDU/rss-reader-backend/database"
	"github.com/mmcdole/gofeed"
)

type Article struct {
	Id             int64  `json:"id"`
	SubscriptionId int64  `json:"subscriptionId"`
	New            bool   `json:"new"`
	Url            string `json:"url"`
	Title          string `json:"title"`
	Description    string `json:"description,omitempty"`
	Thumbnail      string `json:"thumbnail,omitempty"`
	// Time in time.DateTime format.
	Created   string `json:"created"`
	ReadLater bool   `json:"readLater"`
	// Time in time.DateTime format. Can be empty.
	CreatedReadLater string `json:"createdReadLater,omitempty"`
}

func (a Article) ToModel() database.Article {
	return database.Article{
		ID:             a.Id,
		SubscriptionId: a.SubscriptionId,
		New:            a.New,
		Url:            a.Url,
		Title:          a.Title,
		Description: sql.NullString{
			Valid:  a.Description != "",
			String: a.Description,
		},
		Thumbnail: sql.NullString{
			Valid:  a.Thumbnail != "",
			String: a.Thumbnail,
		},
		Created: sql.NullString{
			Valid:  true,
			String: a.Created,
		},
		ReadLater: a.ReadLater,
		CreatedReadLater: sql.NullString{
			Valid:  a.CreatedReadLater != "",
			String: a.CreatedReadLater,
		},
	}
}

func NewArticle(a database.Article) Article {
	return Article{
		Id:               a.ID,
		SubscriptionId:   a.SubscriptionId,
		New:              a.New,
		Url:              a.Url,
		Title:            a.Title,
		Description:      a.Description.String,
		Thumbnail:        a.Thumbnail.String,
		Created:          a.Created.String,
		ReadLater:        a.ReadLater,
		CreatedReadLater: a.CreatedReadLater.String,
	}
}

func NewArticleFromGofeed(article gofeed.Item, subscriptionId int64) Article {
	var thmb string
	if article.Image != nil {
		thmb = article.Image.URL
	}

	var c = time.Now().UTC().Format(time.DateTime)
	if article.PublishedParsed != nil {
		c = article.PublishedParsed.Format(time.DateTime)
	}

	return Article{
		SubscriptionId:   subscriptionId,
		New:              true,
		Url:              article.Link,
		Title:            article.Title,
		Description:      article.Description,
		Thumbnail:        thmb,
		Created:          c,
		ReadLater:        false,
		CreatedReadLater: "",
	}
}

func NewArticlesFromGofeed(articles []*gofeed.Item, subscriptionId int64) (out []Article) {
	for _, item := range articles {
		out = append(out, NewArticleFromGofeed(*item, subscriptionId))
	}

	return
}

type ArticleWithSubscription struct {
	Article
	Subscription Subscription `json:"subscription"`
}

func NewArticleWithSubscription(a database.ArticleWithSubscription) ArticleWithSubscription {
	return ArticleWithSubscription{
		Article:      NewArticle(a.Article),
		Subscription: NewSubscription(a.Subscription),
	}
}

// NewArticleWithSubscriptionTwopart creates an ArticleWithSubscription resource given an article model, and a subscription model
func NewArticleWithSubscriptionTwopart(a database.Article, s database.Subscription) ArticleWithSubscription {
	return ArticleWithSubscription{
		Article:      NewArticle(a),
		Subscription: NewSubscription(s),
	}
}
