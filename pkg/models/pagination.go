package models

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type CursorData struct {
	Timestamp time.Time `json:"ts"`
	EventID   string    `json:"eid"`
}

func EncodeCursor(ts time.Time, eventID string) string {
	data := CursorData{Timestamp: ts, EventID: eventID}
	b, _ := json.Marshal(data)
	return base64.URLEncoding.EncodeToString(b)
}

func DecodeCursor(cursor string) (*CursorData, error) {
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}
	var data CursorData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

type PaginatedResponse[T any] struct {
	Data    []T    `json:"data"`
	Cursor  string `json:"cursor,omitempty"`
	HasMore bool   `json:"has_more"`
	Total   *int64 `json:"total,omitempty"`
}
