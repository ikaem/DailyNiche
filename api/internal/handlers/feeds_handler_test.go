package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/karlo/dailyniche/internal/repos"
)

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
