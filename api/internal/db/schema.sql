-- feeds: subscribed RSS/Atom sources
-- disabled_at is a soft-delete: NULL means active. A non-NULL date means the
-- feed was removed from the dashboard as of that date - the fetcher skips it
-- going forward, but the row is never deleted so past posts (and their
-- feed_id references) keep resolving and past issues never change.
CREATE TABLE IF NOT EXISTS feeds (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    url         TEXT NOT NULL UNIQUE,
    disabled_at TIMESTAMP,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- posts: individual articles pulled from feeds
-- All timestamp columns are stored in UTC.
-- feed_id intentionally has no ON DELETE CASCADE: past issues must stay
-- unchanged even if a feed is later removed, so posts outlive their feed.
CREATE TABLE IF NOT EXISTS posts (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id         INTEGER NOT NULL REFERENCES feeds(id),
    title           TEXT NOT NULL,
    url             TEXT NOT NULL,
    content_summary TEXT,
    published_at    TIMESTAMP NOT NULL,
    fetched_at      TIMESTAMP NOT NULL,
    guid            TEXT NOT NULL UNIQUE,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_posts_feed_id ON posts(feed_id);
CREATE INDEX IF NOT EXISTS idx_posts_published_at ON posts(published_at);
CREATE INDEX IF NOT EXISTS idx_posts_fetched_at ON posts(fetched_at);
