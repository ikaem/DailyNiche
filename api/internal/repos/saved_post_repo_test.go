package repos

import "testing"

func TestFavoritePost_SetsFavoritedAt(t *testing.T) {
	// given: a post with no saved-state row yet
	conn := newTestDB(t)
	feedID, err := CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	postID, err := CreatePost(conn, newTestPost(feedID, "urn:uuid:favorite-post"))
	if err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// when: we favorite it
	if err := FavoritePost(conn, postID); err != nil {
		t.Fatalf("FavoritePost() returned error: %v", err)
	}

	// then: GetSavedPost reports it as favorited, read-later untouched
	sp, err := GetSavedPost(conn, postID)
	if err != nil {
		t.Fatalf("GetSavedPost() returned error: %v", err)
	}
	if sp.FavoritedAt == nil {
		t.Error("expected FavoritedAt to be set")
	}
	if sp.ReadLaterAt != nil {
		t.Errorf("expected ReadLaterAt to remain nil, got %v", sp.ReadLaterAt)
	}
}

func TestFavoritePost_LeavesReadLaterAtUntouchedWhenAlreadySet(t *testing.T) {
	// given: a post already marked read-later
	conn := newTestDB(t)
	feedID, err := CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	postID, err := CreatePost(conn, newTestPost(feedID, "urn:uuid:both-post"))
	if err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}
	if err := MarkReadLater(conn, postID); err != nil {
		t.Fatalf("MarkReadLater() returned error: %v", err)
	}

	// when: we also favorite it
	if err := FavoritePost(conn, postID); err != nil {
		t.Fatalf("FavoritePost() returned error: %v", err)
	}

	// then: a post can be both favorited and read-later at once
	sp, err := GetSavedPost(conn, postID)
	if err != nil {
		t.Fatalf("GetSavedPost() returned error: %v", err)
	}
	if sp.FavoritedAt == nil {
		t.Error("expected FavoritedAt to be set")
	}
	if sp.ReadLaterAt == nil {
		t.Error("expected ReadLaterAt to still be set")
	}
}

func TestUnfavoritePost_ClearsFavoritedAt(t *testing.T) {
	// given: a favorited post
	conn := newTestDB(t)
	feedID, err := CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	postID, err := CreatePost(conn, newTestPost(feedID, "urn:uuid:unfavorite-post"))
	if err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}
	if err := FavoritePost(conn, postID); err != nil {
		t.Fatalf("FavoritePost() returned error: %v", err)
	}

	// when: we unfavorite it
	if err := UnfavoritePost(conn, postID); err != nil {
		t.Fatalf("UnfavoritePost() returned error: %v", err)
	}

	// then: FavoritedAt is cleared
	sp, err := GetSavedPost(conn, postID)
	if err != nil {
		t.Fatalf("GetSavedPost() returned error: %v", err)
	}
	if sp.FavoritedAt != nil {
		t.Errorf("expected FavoritedAt to be nil, got %v", sp.FavoritedAt)
	}
}

func TestUnfavoritePost_NoOpWhenNeverFavorited(t *testing.T) {
	// given: a post with no saved-state row at all
	conn := newTestDB(t)
	feedID, err := CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	postID, err := CreatePost(conn, newTestPost(feedID, "urn:uuid:never-favorited"))
	if err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// when: we unfavorite it anyway
	// then: it's a harmless no-op, not an error
	if err := UnfavoritePost(conn, postID); err != nil {
		t.Fatalf("UnfavoritePost() returned error: %v", err)
	}
}

func TestMarkReadLater_SetsReadLaterAt(t *testing.T) {
	// given: a post with no saved-state row yet
	conn := newTestDB(t)
	feedID, err := CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	postID, err := CreatePost(conn, newTestPost(feedID, "urn:uuid:read-later-post"))
	if err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// when: we mark it read-later
	if err := MarkReadLater(conn, postID); err != nil {
		t.Fatalf("MarkReadLater() returned error: %v", err)
	}

	// then: GetSavedPost reports it as read-later, favorited untouched
	sp, err := GetSavedPost(conn, postID)
	if err != nil {
		t.Fatalf("GetSavedPost() returned error: %v", err)
	}
	if sp.ReadLaterAt == nil {
		t.Error("expected ReadLaterAt to be set")
	}
	if sp.FavoritedAt != nil {
		t.Errorf("expected FavoritedAt to remain nil, got %v", sp.FavoritedAt)
	}
}

func TestUnmarkReadLater_ClearsReadLaterAt(t *testing.T) {
	// given: a read-later post
	conn := newTestDB(t)
	feedID, err := CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	postID, err := CreatePost(conn, newTestPost(feedID, "urn:uuid:unmark-post"))
	if err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}
	if err := MarkReadLater(conn, postID); err != nil {
		t.Fatalf("MarkReadLater() returned error: %v", err)
	}

	// when: we unmark it
	if err := UnmarkReadLater(conn, postID); err != nil {
		t.Fatalf("UnmarkReadLater() returned error: %v", err)
	}

	// then: ReadLaterAt is cleared
	sp, err := GetSavedPost(conn, postID)
	if err != nil {
		t.Fatalf("GetSavedPost() returned error: %v", err)
	}
	if sp.ReadLaterAt != nil {
		t.Errorf("expected ReadLaterAt to be nil, got %v", sp.ReadLaterAt)
	}
}

func TestGetSavedPost_ReturnsZeroValueWhenNeverSaved(t *testing.T) {
	// given: a post that has never been favorited or read-later'd
	conn := newTestDB(t)
	feedID, err := CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	postID, err := CreatePost(conn, newTestPost(feedID, "urn:uuid:untouched-post"))
	if err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// when: we ask for its saved state
	sp, err := GetSavedPost(conn, postID)

	// then: no error - "never saved" is the default, expected state, not
	// an exceptional one
	if err != nil {
		t.Fatalf("GetSavedPost() returned error: %v", err)
	}
	if sp.FavoritedAt != nil || sp.ReadLaterAt != nil {
		t.Errorf("expected a zero-value SavedPost, got %+v", sp)
	}
	if sp.PostID != postID {
		t.Errorf("expected PostID %d, got %d", postID, sp.PostID)
	}
}
