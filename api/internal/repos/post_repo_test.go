package repos

import (
	"testing"
	"time"
)

func TestCreatePost_InsertsNewPost(t *testing.T) {
	// given: a feed to attach the post to
	conn := newTestDB(t)
	feedID, err := CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	post := newTestPost(feedID, "urn:uuid:sample-post")

	// when: we create the post
	id, err := CreatePost(conn, post)
	if err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// then: it returns a positive, newly assigned ID
	if id <= 0 {
		t.Fatalf("expected a positive ID for a new post, got %d", id)
	}

	// and: a row actually exists in the database under that ID - proving
	// the post was really persisted, not just that CreatePost claimed so
	var storedTitle string
	if err := conn.QueryRow(`SELECT title FROM posts WHERE id = ?`, id).Scan(&storedTitle); err != nil {
		t.Fatalf("expected a post row to exist for id %d: %v", id, err)
	}
	if storedTitle != post.Title {
		t.Errorf("expected stored title %q, got %q", post.Title, storedTitle)
	}
}

func TestCreatePost_RoundTripsImageURL(t *testing.T) {
	// given: a post with an image URL
	conn := newTestDB(t)
	feedID, err := CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	post := newTestPost(feedID, "urn:uuid:image-post")

	// when: we create it and read it back via ListPostsByFeed (scanPosts)
	if _, err := CreatePost(conn, post); err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}
	posts, err := ListPostsByFeed(conn, feedID)
	if err != nil {
		t.Fatalf("ListPostsByFeed() returned error: %v", err)
	}

	// then: the image url survives the round trip through both the INSERT
	// and the SELECT/Scan
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
	if posts[0].ImageURL != post.ImageURL {
		t.Errorf("expected image url %q, got %q", post.ImageURL, posts[0].ImageURL)
	}
}

func TestCreatePost_SkipsDuplicateGUID(t *testing.T) {
	// given: a post already created with a given GUID
	conn := newTestDB(t)
	feedID, err := CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	first := newTestPost(feedID, "urn:uuid:duplicate-post")
	if _, err := CreatePost(conn, first); err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// when: we try to create another post with the same GUID, but a later
	// fetched_at (simulating the fetcher re-seeing this post on a later run)
	second := newTestPost(feedID, "urn:uuid:duplicate-post")
	second.FetchedAt = time.Date(2024, 2, 1, 10, 5, 0, 0, time.UTC)
	id, err := CreatePost(conn, second)
	if err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// then: it returns 0, reporting no new row was inserted
	if id != 0 {
		t.Errorf("expected id 0 for a duplicate guid, got %d", id)
	}

	// and: only one row exists, with the original fetched_at preserved -
	// past issues must never change once a post is first stored
	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM posts WHERE guid = ?`, first.GUID).Scan(&count); err != nil {
		t.Fatalf("failed to count posts: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row for guid %q, got %d", first.GUID, count)
	}
	var fetchedAt time.Time
	if err := conn.QueryRow(`SELECT fetched_at FROM posts WHERE guid = ?`, first.GUID).Scan(&fetchedAt); err != nil {
		t.Fatalf("failed to read fetched_at: %v", err)
	}
	if !fetchedAt.Equal(first.FetchedAt) {
		t.Errorf("expected fetched_at to stay %v, got %v", first.FetchedAt, fetchedAt)
	}
}

func TestListPostsByDate_ReturnsOnlyPostsFetchedThatDay(t *testing.T) {
	// given: posts fetched on two different days
	conn := newTestDB(t)
	feedID, err := CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	onTarget := newTestPost(feedID, "urn:uuid:on-target-day")
	onTarget.FetchedAt = time.Date(2024, 3, 15, 9, 0, 0, 0, time.UTC)
	if _, err := CreatePost(conn, onTarget); err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	onOtherDay := newTestPost(feedID, "urn:uuid:other-day")
	onOtherDay.FetchedAt = time.Date(2024, 3, 16, 9, 0, 0, 0, time.UTC)
	if _, err := CreatePost(conn, onOtherDay); err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// when: we list posts for the target day
	posts, err := ListPostsByDate(conn, time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ListPostsByDate() returned error: %v", err)
	}

	// then: only the post fetched on that day is returned
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
	if posts[0].GUID != onTarget.GUID {
		t.Errorf("expected guid %q, got %q", onTarget.GUID, posts[0].GUID)
	}
}

func TestListPostsByDate_ReturnsEmptySliceWhenNoneMatch(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we list posts for any day
	posts, err := ListPostsByDate(conn, time.Now())
	if err != nil {
		t.Fatalf("ListPostsByDate() returned error: %v", err)
	}

	// then: it returns an empty slice, not an error
	if len(posts) != 0 {
		t.Errorf("expected 0 posts, got %d", len(posts))
	}
}

func TestListPostsByFeed_ReturnsOnlyThatFeedsPosts(t *testing.T) {
	// given: posts belonging to two different feeds
	conn := newTestDB(t)
	feedA, err := CreateFeed(conn, "Feed A", "https://a.example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	feedB, err := CreateFeed(conn, "Feed B", "https://b.example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	if _, err := CreatePost(conn, newTestPost(feedA, "urn:uuid:feed-a-post")); err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}
	if _, err := CreatePost(conn, newTestPost(feedB, "urn:uuid:feed-b-post")); err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// when: we list posts for feed A only
	posts, err := ListPostsByFeed(conn, feedA)
	if err != nil {
		t.Fatalf("ListPostsByFeed() returned error: %v", err)
	}

	// then: only feed A's post is returned
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
	if posts[0].FeedID != feedA {
		t.Errorf("expected feed_id %d, got %d", feedA, posts[0].FeedID)
	}
}

func TestDeletePostsByDate_RemovesOnlyOlderPosts(t *testing.T) {
	// given: an older post and a newer post
	conn := newTestDB(t)
	feedID, err := CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	older := newTestPost(feedID, "urn:uuid:older-post")
	older.FetchedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := CreatePost(conn, older); err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}
	newer := newTestPost(feedID, "urn:uuid:newer-post")
	newer.FetchedAt = time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	if _, err := CreatePost(conn, newer); err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// when: we delete posts fetched before March 2024
	cutoff := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	if err := DeletePostsByDate(conn, cutoff); err != nil {
		t.Fatalf("DeletePostsByDate() returned error: %v", err)
	}

	// then: the older post is gone, the newer post remains
	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM posts WHERE guid = ?`, older.GUID).Scan(&count); err != nil {
		t.Fatalf("failed to count posts: %v", err)
	}
	if count != 0 {
		t.Errorf("expected the older post to be deleted, but %d row(s) remain", count)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM posts WHERE guid = ?`, newer.GUID).Scan(&count); err != nil {
		t.Fatalf("failed to count posts: %v", err)
	}
	if count != 1 {
		t.Errorf("expected the newer post to still exist, got count %d", count)
	}
}
