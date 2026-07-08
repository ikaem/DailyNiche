package handlers

import (
	"testing"
	"time"
)

func TestParseDateParam_DefaultsToTodayWhenEmpty(t *testing.T) {
	// given: no date string
	// when: we parse an empty string
	before := time.Now().UTC()
	got, err := parseDateParam("")
	after := time.Now().UTC()

	// then: it returns "now", in UTC
	if err != nil {
		t.Fatalf("parseDateParam() returned error: %v", err)
	}
	if got.Before(before) || got.After(after) {
		t.Errorf("expected a time between %v and %v, got %v", before, after, got)
	}
	if got.Location() != time.UTC {
		t.Errorf("expected UTC location, got %v", got.Location())
	}
}

func TestParseDateParam_ParsesValidDate(t *testing.T) {
	// given: a valid YYYY-MM-DD string
	// when: we parse it
	got, err := parseDateParam("2026-07-08")
	if err != nil {
		t.Fatalf("parseDateParam() returned error: %v", err)
	}

	// then: it matches the expected date, in UTC
	want := time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestParseDateParam_ReturnsErrorForInvalidDate(t *testing.T) {
	// given: a malformed date string
	// when: we parse it
	_, err := parseDateParam("not-a-date")

	// then: it returns an error
	if err == nil {
		t.Fatal("expected an error for an invalid date, got nil")
	}
}
