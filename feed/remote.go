package feed

import "github.com/mmcdole/gofeed"

// Fetches a remote feed using gofeed, returns the feed object and articles in it
func FetchRemote(url string) (*Feed, []Article, error) {
	gf, err := gofeed.NewParser().ParseURL(url)
	if err != nil {
		return nil, nil, err
	}

	f, err := NewFeedFromGofeed(gf)
	if err != nil {
		return nil, nil, err
	}

	a := []Article{}
	for _, gart := range gf.Items {
		art, err := NewArticleFromGofeed(gart)
		if err != nil {
			return nil, nil, err
		}
		a = append(a, *art)
	}

	return f, a, nil
}
