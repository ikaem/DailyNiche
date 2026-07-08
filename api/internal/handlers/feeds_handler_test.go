package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/karlo/dailyniche/internal/repos"
)

// postJSON builds a POST request to path with body encoded as JSON.
func postJSON(t *testing.T, path, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestFeeds_ReturnsAllFeeds(t *testing.T) {
	// given: two feeds, one of them disabled
	conn := newTestDB(t)
	activeID, err := repos.CreateFeed(conn, "Active Blog", "https://active.example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	disabledID, err := repos.CreateFeed(conn, "Disabled Blog", "https://disabled.example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	if err := repos.DeleteFeed(conn, disabledID); err != nil {
		t.Fatalf("DeleteFeed() returned error: %v", err)
	}

	// when: we request /api/feeds
	req := httptest.NewRequest(http.MethodGet, "/api/feeds", nil)
	rec := httptest.NewRecorder()
	Feeds(conn)(rec, req)

	// then: it responds 200 with both feeds, active and disabled
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var got []FeedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 feeds, got %d", len(got))
	}

	byID := make(map[int64]FeedResponse, len(got))
	for _, f := range got {
		byID[f.ID] = f
	}
	if byID[activeID].DisabledAt != nil {
		t.Errorf("expected active feed's disabled_at to be nil, got %v", byID[activeID].DisabledAt)
	}
	if byID[disabledID].DisabledAt == nil {
		t.Error("expected disabled feed's disabled_at to be set")
	}
}

func TestFeeds_ReturnsEmptyArrayWhenNoFeeds(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we request /api/feeds
	req := httptest.NewRequest(http.MethodGet, "/api/feeds", nil)
	rec := httptest.NewRecorder()
	Feeds(conn)(rec, req)

	// then: it responds 200 with an empty JSON array, not null
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Errorf("expected empty JSON array \"[]\", got %q", body)
	}
}

func TestCreateFeed_CreatesFeedAndReturns201(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we POST a valid feed
	req := postJSON(t, "/api/feeds", `{"name":"Tech Blog","url":"https://example.com/feed.xml"}`)
	rec := httptest.NewRecorder()
	CreateFeed(conn)(rec, req)

	// then: it responds 201 with the created feed
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var got FeedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.ID <= 0 {
		t.Errorf("expected a positive ID, got %d", got.ID)
	}
	if got.Name != "Tech Blog" {
		t.Errorf("expected name %q, got %q", "Tech Blog", got.Name)
	}
	if got.URL != "https://example.com/feed.xml" {
		t.Errorf("expected url %q, got %q", "https://example.com/feed.xml", got.URL)
	}
	if got.DisabledAt != nil {
		t.Errorf("expected disabled_at to be nil for a new feed, got %v", got.DisabledAt)
	}

	// and: it's actually persisted, not just claimed in the response
	stored, err := repos.GetFeed(conn, got.ID)
	if err != nil {
		t.Fatalf("expected feed to exist in the database: %v", err)
	}
	if stored.Name != "Tech Blog" {
		t.Errorf("expected stored name %q, got %q", "Tech Blog", stored.Name)
	}
}

func TestCreateFeed_ReturnsBadRequestForMissingName(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we POST without a name
	req := postJSON(t, "/api/feeds", `{"name":"","url":"https://example.com/feed.xml"}`)
	rec := httptest.NewRecorder()
	CreateFeed(conn)(rec, req)

	// then: it responds 400
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestCreateFeed_ReturnsBadRequestForMissingURL(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we POST without a url
	req := postJSON(t, "/api/feeds", `{"name":"Tech Blog","url":""}`)
	rec := httptest.NewRecorder()
	CreateFeed(conn)(rec, req)

	// then: it responds 400
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestCreateFeed_ReturnsBadRequestForMalformedURL(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we POST a url with no scheme/host
	req := postJSON(t, "/api/feeds", `{"name":"Tech Blog","url":"not-a-url"}`)
	rec := httptest.NewRecorder()
	CreateFeed(conn)(rec, req)

	// then: it responds 400
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestCreateFeed_ReturnsBadRequestForInvalidJSON(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we POST malformed JSON
	req := postJSON(t, "/api/feeds", `{not valid json`)
	rec := httptest.NewRecorder()
	CreateFeed(conn)(rec, req)

	// then: it responds 400
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestCreateFeed_ReturnsConflictForDuplicateURL(t *testing.T) {
	// given: a feed already created with a given URL
	conn := newTestDB(t)
	if _, err := repos.CreateFeed(conn, "Existing Blog", "https://example.com/feed.xml"); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	// when: we POST a new feed with the same URL
	req := postJSON(t, "/api/feeds", `{"name":"Different Name","url":"https://example.com/feed.xml"}`)
	rec := httptest.NewRecorder()
	CreateFeed(conn)(rec, req)

	// then: it responds 409
	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}
