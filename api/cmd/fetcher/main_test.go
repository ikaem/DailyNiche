package main

import "testing"

func TestParseFlags_DefaultsToAllFalse(t *testing.T) {
	// given: no flags
	// when: we parse an empty argument list
	cfg, err := parseFlags([]string{})
	if err != nil {
		t.Fatalf("parseFlags() returned error: %v", err)
	}

	// then: every option defaults to false
	if cfg.Once || cfg.Verbose || cfg.DryRun {
		t.Errorf("expected all-false defaults, got %+v", cfg)
	}
}

func TestParseFlags_SetsFieldsFromFlags(t *testing.T) {
	// given: all three flags passed
	// when: we parse them
	cfg, err := parseFlags([]string{"-once", "-verbose", "-dry-run"})
	if err != nil {
		t.Fatalf("parseFlags() returned error: %v", err)
	}

	// then: each corresponding field is true
	if !cfg.Once {
		t.Error("expected Once to be true")
	}
	if !cfg.Verbose {
		t.Error("expected Verbose to be true")
	}
	if !cfg.DryRun {
		t.Error("expected DryRun to be true")
	}
}

func TestParseFlags_ReturnsErrorForUnknownFlag(t *testing.T) {
	// given: an unrecognized flag
	// when: we parse it
	_, err := parseFlags([]string{"-bogus"})

	// then: it returns an error
	if err == nil {
		t.Fatal("expected an error for an unknown flag, got nil")
	}
}
