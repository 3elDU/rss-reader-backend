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
  -- optional thumbnail (favicon or channel thumbnail)
  thumbnail BLOB
);
CREATE TABLE IF NOT EXISTS articles (
  id INTEGER PRIMARY KEY ASC,
  -- feed that this article belongs to
  subscription_id INTEGER NOT NULL,
  -- whether the article was shown to the user
  new INT,
  -- title of the article
  title TEXT NOT NULL,
  -- description of the article
  description TEXT,
  -- thumbnail
  thumbnail BLOB,
  FOREIGN KEY(subscription_id) REFERENCES subscriptions(id)
);
CREATE TABLE IF NOT EXISTS readlater (
  id INTEGER PRIMARY KEY ASC,
  article_id INTEGER NOT NULL,
  FOREIGN KEY(article_id) REFERENCES articles(id)
);