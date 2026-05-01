package ui

import "testing"

func TestParsePositiveMinutesAcceptsDecimals(t *testing.T) {
	value, err := parsePositiveMinutes(".5", "rest minutes")
	if err != nil {
		t.Fatalf("expected decimal minutes to parse, got error: %v", err)
	}

	if value != 0.5 {
		t.Fatalf("expected 0.5, got %v", value)
	}
}

func TestParsePositiveMinutesRejectsZeroOrLess(t *testing.T) {
	if _, err := parsePositiveMinutes("0", "work minutes"); err == nil {
		t.Fatal("expected zero minutes to be rejected")
	}
}
