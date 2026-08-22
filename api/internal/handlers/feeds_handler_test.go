package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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

// putJSON builds a PUT request to path with body encoded as JSON, and the
// {id} path value pre-set - mirroring how the real mux populates it via its
// "PUT /api/feeds/{id}" pattern.
func putJSON(t *testing.T, id, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/api/feeds/"+id, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", id)
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

func TestUpdateFeed_UpdatesFeedAndReturns200(t *testing.T) {
	// given: an existing feed
	conn := newTestDB(t)
	id, err := repos.CreateFeed(conn, "Old Name", "https://old.example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	// when: we PUT a corrected name/url
	req := putJSON(t, strconv.FormatInt(id, 10), `{"name":"New Name","url":"https://new.example.com/feed.xml"}`)
	rec := httptest.NewRecorder()
	UpdateFeed(conn)(rec, req)

	// then: it responds 200 with the updated feed
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got FeedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.Name != "New Name" || got.URL != "https://new.example.com/feed.xml" {
		t.Errorf("expected updated name/url in response, got %+v", got)
	}

	// and: it's actually persisted, not just claimed in the response
	stored, err := repos.GetFeed(conn, id)
	if err != nil {
		t.Fatalf("expected feed to still exist: %v", err)
	}
	if stored.Name != "New Name" || stored.URL != "https://new.example.com/feed.xml" {
		t.Errorf("expected persisted name/url to be updated, got %+v", stored)
	}
}

func TestUpdateFeed_ReturnsNotFoundForNonexistentID(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we PUT an id that doesn't exist
	req := putJSON(t, "999", `{"name":"New Name","url":"https://example.com/feed.xml"}`)
	rec := httptest.NewRecorder()
	UpdateFeed(conn)(rec, req)

	// then: it responds 404
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestUpdateFeed_ReturnsBadRequestForInvalidID(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we PUT with a non-numeric id
	req := putJSON(t, "abc", `{"name":"New Name","url":"https://example.com/feed.xml"}`)
	rec := httptest.NewRecorder()
	UpdateFeed(conn)(rec, req)

	// then: it responds 400
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateFeed_ReturnsBadRequestForMissingName(t *testing.T) {
	// given: an existing feed
	conn := newTestDB(t)
	id, err := repos.CreateFeed(conn, "Old Name", "https://old.example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	// when: we PUT with an empty name
	req := putJSON(t, strconv.FormatInt(id, 10), `{"name":"","url":"https://old.example.com/feed.xml"}`)
	rec := httptest.NewRecorder()
	UpdateFeed(conn)(rec, req)

	// then: it responds 400
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateFeed_ReturnsBadRequestForMalformedURL(t *testing.T) {
	// given: an existing feed
	conn := newTestDB(t)
	id, err := repos.CreateFeed(conn, "Old Name", "https://old.example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	// when: we PUT a url with no scheme/host
	req := putJSON(t, strconv.FormatInt(id, 10), `{"name":"Old Name","url":"not-a-url"}`)
	rec := httptest.NewRecorder()
	UpdateFeed(conn)(rec, req)

	// then: it responds 400
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateFeed_ReturnsBadRequestForInvalidJSON(t *testing.T) {
	// given: an existing feed
	conn := newTestDB(t)
	id, err := repos.CreateFeed(conn, "Old Name", "https://old.example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	// when: we PUT malformed JSON
	req := putJSON(t, strconv.FormatInt(id, 10), `{not valid json`)
	rec := httptest.NewRecorder()
	UpdateFeed(conn)(rec, req)

	// then: it responds 400
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestUpdateFeed_ReturnsConflictWhenURLCollidesWithAnotherFeed(t *testing.T) {
	// given: two existing feeds
	conn := newTestDB(t)
	if _, err := repos.CreateFeed(conn, "Feed A", "https://a.example.com/feed.xml"); err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	idB, err := repos.CreateFeed(conn, "Feed B", "https://b.example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	// when: we PUT feed B with feed A's url
	req := putJSON(t, strconv.FormatInt(idB, 10), `{"name":"Feed B","url":"https://a.example.com/feed.xml"}`)
	rec := httptest.NewRecorder()
	UpdateFeed(conn)(rec, req)

	// then: it responds 409
	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

func TestEnableFeed_ClearsDisabledAtAndReturns200(t *testing.T) {
	// given: a disabled feed
	conn := newTestDB(t)
	id, err := repos.CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}
	if err := repos.DeleteFeed(conn, id); err != nil {
		t.Fatalf("DeleteFeed() returned error: %v", err)
	}

	// when: we POST to enable it
	req := httptest.NewRequest(http.MethodPost, "/api/feeds/"+strconv.FormatInt(id, 10)+"/enable", nil)
	req.SetPathValue("id", strconv.FormatInt(id, 10))
	rec := httptest.NewRecorder()
	EnableFeed(conn)(rec, req)

	// then: it responds 200 with disabled_at cleared
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got FeedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.DisabledAt != nil {
		t.Errorf("expected disabled_at to be nil in response, got %v", got.DisabledAt)
	}

	// and: it's actually persisted, not just claimed in the response
	stored, err := repos.GetFeed(conn, id)
	if err != nil {
		t.Fatalf("expected feed to still exist: %v", err)
	}
	if stored.DisabledAt != nil {
		t.Errorf("expected persisted disabled_at to be nil, got %v", stored.DisabledAt)
	}
}

func TestEnableFeed_ReturnsNotFoundForNonexistentID(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we POST to enable an id that doesn't exist
	req := httptest.NewRequest(http.MethodPost, "/api/feeds/999/enable", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()
	EnableFeed(conn)(rec, req)

	// then: it responds 404
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestEnableFeed_ReturnsBadRequestForInvalidID(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we POST to enable with a non-numeric id
	req := httptest.NewRequest(http.MethodPost, "/api/feeds/abc/enable", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()
	EnableFeed(conn)(rec, req)

	// then: it responds 400
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestDeleteFeed_SoftDeletesAndReturns204(t *testing.T) {
	// given: an existing feed
	conn := newTestDB(t)
	id, err := repos.CreateFeed(conn, "Sample Blog", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("CreateFeed() returned error: %v", err)
	}

	// when: we DELETE it
	req := httptest.NewRequest(http.MethodDelete, "/api/feeds/"+strconv.FormatInt(id, 10), nil)
	req.SetPathValue("id", strconv.FormatInt(id, 10))
	rec := httptest.NewRecorder()
	DeleteFeed(conn)(rec, req)

	// then: it responds 204
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}

	// and: the feed row still exists, soft-deleted rather than removed
	feed, err := repos.GetFeed(conn, id)
	if err != nil {
		t.Fatalf("expected feed to still exist after delete: %v", err)
	}
	if feed.DisabledAt == nil {
		t.Error("expected DisabledAt to be set after delete")
	}
}

func TestDeleteFeed_ReturnsNotFoundForNonexistentID(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we DELETE an ID that doesn't exist
	req := httptest.NewRequest(http.MethodDelete, "/api/feeds/999", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()
	DeleteFeed(conn)(rec, req)

	// then: it responds 404
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestDeleteFeed_ReturnsBadRequestForInvalidID(t *testing.T) {
	// given: an empty database
	conn := newTestDB(t)

	// when: we DELETE with a non-numeric id
	req := httptest.NewRequest(http.MethodDelete, "/api/feeds/abc", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()
	DeleteFeed(conn)(rec, req)

	// then: it responds 400
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}
