package resource

import (
	"database/sql"

	"github.com/3elDU/rss-reader-backend/database"
	"github.com/mmcdole/gofeed"
)

type Subscription struct {
	Id    int64  `json:"id,omitzero"`
	Type  string `json:"type"`
	Url   string `json:"url"`
	Title string `json:"title"`
	// Description can be empty.
	Description string `json:"description,omitempty"`
	// Thumbnail can be empty.
	Thumbnail string `json:"thumbnail,omitempty"`
}

func (s Subscription) ToModel() database.Subscription {

	return database.Subscription{
		Type:  s.Type,
		Url:   s.Url,
		Title: s.Title,
		Description: sql.NullString{
			Valid:  s.Description != "",
			String: s.Description,
		},
		Thumbnail: sql.NullString{
			Valid:  s.Thumbnail != "",
			String: s.Thumbnail,
		},
	}
}

func NewSubscription(m database.Subscription) Subscription {
	return Subscription{
		Id:          m.ID,
		Type:        m.Type,
		Url:         m.Url,
		Title:       m.Title,
		Description: m.Description.String,
		Thumbnail:   m.Thumbnail.String,
	}
}

func NewSubscriptionFromGofeed(feed gofeed.Feed) Subscription {
	t := ""
	if feed.Image != nil {
		t = feed.Image.URL
	}

	return Subscription{
		Type:        feed.FeedType,
		Url:         feed.FeedLink,
		Title:       feed.Title,
		Description: feed.Description,
		Thumbnail:   t,
	}
}
