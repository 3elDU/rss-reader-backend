CREATE TABLE IF NOT EXISTS subscriptions (
  id INTEGER PRIMARY KEY ASC,
  -- "youtube", "rss", "atom" or "telegram"
  type TEXT NOT NULL,
  -- url to the feed
  url TEXT NOT NULL,
  -- title that will be shown to the user
  title TEXT NOT NULL,
  -- description of the feed
  description TEXT,
  -- thumbnail URL, null if there's no thumbnail
  thumbnail TEXT
);
CREATE TABLE IF NOT EXISTS articles (
  id INTEGER PRIMARY KEY ASC,
  -- feed that this article belongs to
  subscription_id INTEGER NOT NULL,
  -- whether the article was shown to the user
  new INT,
  -- URL to the article
  url TEXT NOT NULL,
  -- title of the article
  title TEXT NOT NULL,
  -- description of the article
  description TEXT,
  -- thumbnail URL, null if there's no thumbnail
  thumbnail TEXT,
  -- date when the article was written
  created TEXT NOT NULL,
  FOREIGN KEY(subscription_id) REFERENCES subscriptions(id)
);
CREATE TABLE IF NOT EXISTS readlater (
  id INTEGER PRIMARY KEY ASC,
  article_id INTEGER NOT NULL,
  FOREIGN KEY(article_id) REFERENCES articles(id)
);