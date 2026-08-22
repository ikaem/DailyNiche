package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/karlo/dailyniche/internal/repos"
)

func TestFavoritePost_SetsFavoritedAtAndReturns200(t *testing.T) {
	// given: an existing post
	conn := newTestDB(t)
	feedID, err := repos.CreateFeed(conn, "Tech Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	postID, err := repos.CreatePost(conn, newTestPost(feedID, "urn:uuid:favorite", time.Now().UTC()))
	if err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// when: we POST to favorite it
	req := httptest.NewRequest(http.MethodPost, "/api/posts/"+strconv.FormatInt(postID, 10)+"/favorite", nil)
	req.SetPathValue("id", strconv.FormatInt(postID, 10))
	rec := httptest.NewRecorder()
	FavoritePost(conn)(rec, req)

	// then: it responds 200 with favorited_at set
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got SavedPostResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.FavoritedAt == nil {
		t.Error("expected favorited_at to be set")
	}

	// and: it's actually persisted, not just claimed in the response
	sp, err := repos.GetSavedPost(conn, postID)
	if err != nil {
		t.Fatalf("GetSavedPost() returned error: %v", err)
	}
	if sp.FavoritedAt == nil {
		t.Error("expected persisted FavoritedAt to be set")
	}
}

func TestFavoritePost_ReturnsNotFoundForNonexistentPost(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we POST to favorite an id that doesn't exist
	req := httptest.NewRequest(http.MethodPost, "/api/posts/999/favorite", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()
	FavoritePost(conn)(rec, req)

	// then: it responds 404
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestFavoritePost_ReturnsBadRequestForInvalidID(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we POST with a non-numeric id
	req := httptest.NewRequest(http.MethodPost, "/api/posts/abc/favorite", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()
	FavoritePost(conn)(rec, req)

	// then: it responds 400
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestUnfavoritePost_ClearsFavoritedAtAndReturns200(t *testing.T) {
	// given: a favorited post
	conn := newTestDB(t)
	feedID, err := repos.CreateFeed(conn, "Tech Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	postID, err := repos.CreatePost(conn, newTestPost(feedID, "urn:uuid:unfavorite", time.Now().UTC()))
	if err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}
	if err := repos.FavoritePost(conn, postID); err != nil {
		t.Fatalf("FavoritePost() returned error: %v", err)
	}

	// when: we DELETE to unfavorite it
	req := httptest.NewRequest(http.MethodDelete, "/api/posts/"+strconv.FormatInt(postID, 10)+"/favorite", nil)
	req.SetPathValue("id", strconv.FormatInt(postID, 10))
	rec := httptest.NewRecorder()
	UnfavoritePost(conn)(rec, req)

	// then: it responds 200 with favorited_at cleared
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got SavedPostResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.FavoritedAt != nil {
		t.Errorf("expected favorited_at to be nil, got %v", got.FavoritedAt)
	}
}

func TestUnfavoritePost_ReturnsNotFoundForNonexistentPost(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we DELETE to unfavorite an id that doesn't exist
	req := httptest.NewRequest(http.MethodDelete, "/api/posts/999/favorite", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()
	UnfavoritePost(conn)(rec, req)

	// then: it responds 404
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestMarkReadLater_SetsReadLaterAtAndReturns200(t *testing.T) {
	// given: an existing post
	conn := newTestDB(t)
	feedID, err := repos.CreateFeed(conn, "Tech Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	postID, err := repos.CreatePost(conn, newTestPost(feedID, "urn:uuid:read-later", time.Now().UTC()))
	if err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}

	// when: we POST to mark it read-later
	req := httptest.NewRequest(http.MethodPost, "/api/posts/"+strconv.FormatInt(postID, 10)+"/read-later", nil)
	req.SetPathValue("id", strconv.FormatInt(postID, 10))
	rec := httptest.NewRecorder()
	MarkReadLater(conn)(rec, req)

	// then: it responds 200 with read_later_at set
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got SavedPostResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.ReadLaterAt == nil {
		t.Error("expected read_later_at to be set")
	}
}

func TestMarkReadLater_ReturnsNotFoundForNonexistentPost(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we POST to mark read-later an id that doesn't exist
	req := httptest.NewRequest(http.MethodPost, "/api/posts/999/read-later", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()
	MarkReadLater(conn)(rec, req)

	// then: it responds 404
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestUnmarkReadLater_ClearsReadLaterAtAndReturns200(t *testing.T) {
	// given: a read-later post
	conn := newTestDB(t)
	feedID, err := repos.CreateFeed(conn, "Tech Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	postID, err := repos.CreatePost(conn, newTestPost(feedID, "urn:uuid:unmark", time.Now().UTC()))
	if err != nil {
		t.Fatalf("CreatePost() returned error: %v", err)
	}
	if err := repos.MarkReadLater(conn, postID); err != nil {
		t.Fatalf("MarkReadLater() returned error: %v", err)
	}

	// when: we DELETE to unmark it
	req := httptest.NewRequest(http.MethodDelete, "/api/posts/"+strconv.FormatInt(postID, 10)+"/read-later", nil)
	req.SetPathValue("id", strconv.FormatInt(postID, 10))
	rec := httptest.NewRecorder()
	UnmarkReadLater(conn)(rec, req)

	// then: it responds 200 with read_later_at cleared
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got SavedPostResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.ReadLaterAt != nil {
		t.Errorf("expected read_later_at to be nil, got %v", got.ReadLaterAt)
	}
}

func TestUnmarkReadLater_ReturnsNotFoundForNonexistentPost(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we DELETE to unmark read-later an id that doesn't exist
	req := httptest.NewRequest(http.MethodDelete, "/api/posts/999/read-later", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()
	UnmarkReadLater(conn)(rec, req)

	// then: it responds 404
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}
