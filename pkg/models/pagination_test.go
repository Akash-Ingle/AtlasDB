package models

import (
	"testing"
	"time"
)

func TestCursorRoundTrip(t *testing.T) {
	ts := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	eventID := "01J5M2KABC123"

	encoded := EncodeCursor(ts, eventID)
	if encoded == "" {
		t.Fatal("encoded cursor is empty")
	}

	decoded, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("decode cursor: %v", err)
	}

	if !decoded.Timestamp.Equal(ts) {
		t.Errorf("timestamp = %v, want %v", decoded.Timestamp, ts)
	}

	if decoded.EventID != eventID {
		t.Errorf("event_id = %q, want %q", decoded.EventID, eventID)
	}
}

func TestDecodeCursorInvalid(t *testing.T) {
	cases := []string{
		"",
		"not-base64!!!",
		"bm90LWpzb24=", // "not-json" in base64
	}

	for _, c := range cases {
		_, err := DecodeCursor(c)
		if err == nil {
			t.Errorf("DecodeCursor(%q) expected error, got nil", c)
		}
	}
}
