package ingest

import (
	"encoding/json"
	"testing"

	"github.com/atlasdb/atlasdb/pkg/models"
)

func TestValidateEventInput_Valid(t *testing.T) {
	e := models.EventInput{
		Source:    "payment-service",
		EventType: "http_request",
		Severity:  models.SeverityInfo,
		Data:      json.RawMessage(`{"status": 200}`),
		Tags:      []string{"production"},
	}

	errs := ValidateEventInput(0, e, 1048576)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateEventInput_MissingSource(t *testing.T) {
	e := models.EventInput{
		EventType: "http_request",
	}

	errs := ValidateEventInput(0, e, 1048576)
	if len(errs) == 0 {
		t.Error("expected error for missing source")
	}

	found := false
	for _, err := range errs {
		if err.Field == "source" {
			found = true
		}
	}
	if !found {
		t.Error("expected source field error")
	}
}

func TestValidateEventInput_MissingEventType(t *testing.T) {
	e := models.EventInput{
		Source: "test-service",
	}

	errs := ValidateEventInput(0, e, 1048576)
	found := false
	for _, err := range errs {
		if err.Field == "event_type" {
			found = true
		}
	}
	if !found {
		t.Error("expected event_type field error")
	}
}

func TestValidateEventInput_InvalidSeverity(t *testing.T) {
	e := models.EventInput{
		Source:    "test-service",
		EventType: "test",
		Severity:  "critical",
	}

	errs := ValidateEventInput(0, e, 1048576)
	found := false
	for _, err := range errs {
		if err.Field == "severity" {
			found = true
		}
	}
	if !found {
		t.Error("expected severity field error")
	}
}

func TestValidateEventInput_OversizedData(t *testing.T) {
	bigData := make([]byte, 100)
	for i := range bigData {
		bigData[i] = 'x'
	}

	e := models.EventInput{
		Source:    "test-service",
		EventType: "test",
		Data:      json.RawMessage(bigData),
	}

	errs := ValidateEventInput(0, e, 50) // 50 byte limit
	found := false
	for _, err := range errs {
		if err.Field == "data" {
			found = true
		}
	}
	if !found {
		t.Error("expected data field error for oversized payload")
	}
}

func TestValidateEventInput_TooManyTags(t *testing.T) {
	tags := make([]string, 51)
	for i := range tags {
		tags[i] = "tag"
	}

	e := models.EventInput{
		Source:    "test-service",
		EventType: "test",
		Tags:      tags,
	}

	errs := ValidateEventInput(0, e, 1048576)
	found := false
	for _, err := range errs {
		if err.Field == "tags" {
			found = true
		}
	}
	if !found {
		t.Error("expected tags field error")
	}
}
