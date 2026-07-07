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
