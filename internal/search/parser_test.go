package search

import (
	"testing"
)

func TestParseSimpleKeyword(t *testing.T) {
	pq := Parse("timeout error")
	if pq.FreeText == "" {
		t.Error("expected free text")
	}
	if len(pq.Filters) != 0 {
		t.Errorf("expected no filters, got %v", pq.Filters)
	}
}

func TestParseFieldFilter(t *testing.T) {
	pq := Parse("source:payment-service error")
	if pq.Filters["source"] != "payment-service" {
		t.Errorf("source filter = %q, want payment-service", pq.Filters["source"])
	}
	if pq.FreeText != "error" {
		t.Errorf("free text = %q, want error", pq.FreeText)
	}
}

func TestParseMultipleFilters(t *testing.T) {
	pq := Parse("source:auth-service severity:error login failed")
	if pq.Filters["source"] != "auth-service" {
		t.Errorf("source = %q", pq.Filters["source"])
	}
	if pq.Filters["severity"] != "error" {
		t.Errorf("severity = %q", pq.Filters["severity"])
	}
	if pq.FreeText != "login failed" {
		t.Errorf("free text = %q, want 'login failed'", pq.FreeText)
	}
}

func TestParseBooleanOperators(t *testing.T) {
	pq := Parse("timeout OR refused")
	if pq.FreeText != "timeout | refused" {
		t.Errorf("free text = %q, want 'timeout | refused'", pq.FreeText)
	}
}

func TestParseQuotedPhrase(t *testing.T) {
	pq := Parse(`"connection refused" source:db`)
	if pq.Filters["source"] != "db" {
		t.Errorf("source = %q", pq.Filters["source"])
	}
	// Quoted phrase should be in free text tokens
	if pq.FreeText == "" {
		t.Error("expected free text from quoted phrase")
	}
}

func TestParseUnknownFieldTreatedAsFreeText(t *testing.T) {
	pq := Parse("unknown_field:value test")
	if _, ok := pq.Filters["unknown_field"]; ok {
		t.Error("unknown field should not be in filters")
	}
}

func TestBuildSQL(t *testing.T) {
	pq := Parse("source:payment-service error")
	conditions, args, nextArg := pq.BuildSQL(1)

	if len(conditions) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(conditions))
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if nextArg != 3 {
		t.Errorf("nextArg = %d, want 3", nextArg)
	}
	if args[0] != "payment-service" {
		t.Errorf("args[0] = %v, want payment-service", args[0])
	}
}

func TestTokenize(t *testing.T) {
	tokens := tokenize(`hello "world foo" bar`)
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0] != "hello" {
		t.Errorf("tokens[0] = %q", tokens[0])
	}
	if tokens[1] != `"world foo"` {
		t.Errorf("tokens[1] = %q", tokens[1])
	}
	if tokens[2] != "bar" {
		t.Errorf("tokens[2] = %q", tokens[2])
	}
}
