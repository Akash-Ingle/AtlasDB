package models

import (
	"encoding/json"
	"time"
)

type Severity string

const (
	SeverityDebug Severity = "debug"
	SeverityInfo  Severity = "info"
	SeverityWarn  Severity = "warn"
	SeverityError Severity = "error"
	SeverityFatal Severity = "fatal"
)

func (s Severity) Valid() bool {
	switch s {
	case SeverityDebug, SeverityInfo, SeverityWarn, SeverityError, SeverityFatal:
		return true
	}
	return false
}

type Event struct {
	EventID    string          `json:"event_id"`
	Source     string          `json:"source"`
	EventType  string          `json:"event_type"`
	Severity   Severity        `json:"severity"`
	Timestamp  time.Time       `json:"timestamp"`
	ReceivedAt time.Time       `json:"received_at"`
	Data       json.RawMessage `json:"data"`
	Tags       []string        `json:"tags"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
}

type EventInput struct {
	Source    string          `json:"source"`
	EventType string          `json:"event_type"`
	Severity  Severity        `json:"severity"`
	Timestamp *time.Time      `json:"timestamp,omitempty"`
	Data      json.RawMessage `json:"data"`
	Tags      []string        `json:"tags,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
}

type IngestRequest struct {
	Events []EventInput `json:"events"`
}

type IngestResponse struct {
	Accepted int      `json:"accepted"`
	EventIDs []string `json:"event_ids"`
}

type EventQuery struct {
	Source    string     `json:"source,omitempty"`
	EventType string     `json:"event_type,omitempty"`
	Severity  Severity   `json:"severity,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Cursor    string     `json:"cursor,omitempty"`
}

const (
	DefaultPageSize = 50
	MaxPageSize     = 1000
)
