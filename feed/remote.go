package feed

import "github.com/mmcdole/gofeed"

// FetchRemoteFeed fetches a remote feed through rss
func FetchRemoteFeed(parser *gofeed.Parser, url string) (*Feed, error) {
	gf, err := parser.ParseURL(url)
	if err != nil {
		return nil, err
	}

	f, err := NewFeedFromGofeed(gf)
	f.URL = url
	if err != nil {
		return nil, err
	}

	return f, nil
}

// FetchRemote fetches a remote feed through rss along with all articles in it.
func FetchRemote(parser *gofeed.Parser, url string) (*Feed, []Article, error) {
	gf, err := parser.ParseURL(url)
	if err != nil {
		return nil, nil, err
	}

	f, err := NewFeedFromGofeed(gf)
	f.URL = url
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
