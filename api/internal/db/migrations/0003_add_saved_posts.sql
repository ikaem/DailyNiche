-- saved_posts: per-post favorite/read-later state - kept as its own table
-- rather than columns on posts, since these represent the user's
-- relationship to a post, not a property of the post's fetched content
-- itself. post_id is the primary key (not a separate surrogate id) since
-- this is a true 1:1 extension of posts, not an independent entity - see
-- CLAUDE.md's Saved Posts design notes.
-- Both columns are independent and nullable, matching feeds.disabled_at's
-- existing nullable-timestamp-as-flag convention - a post can be both
-- favorited and read-later at once.
CREATE TABLE IF NOT EXISTS saved_posts (
    post_id       INTEGER PRIMARY KEY REFERENCES posts(id),
    favorited_at  TIMESTAMP,
    read_later_at TIMESTAMP
);
