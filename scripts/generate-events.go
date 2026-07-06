// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

var (
	sources = []string{
		"payment-service", "user-service", "order-service",
		"catalog-service", "auth-service", "notification-service",
		"gateway", "search-service",
	}

	eventTypes = []string{
		"http_request", "error", "deployment", "authentication",
		"database_query", "cache_operation", "queue_operation",
	}

	severities = []string{"debug", "info", "warn", "error"}

	endpoints = []string{
		"/api/users", "/api/orders", "/api/payments",
		"/api/catalog", "/api/search", "/api/auth/login",
		"/api/auth/register", "/api/notifications",
	}

	methods   = []string{"GET", "POST", "PUT", "DELETE"}
	statuses  = []int{200, 201, 204, 400, 401, 403, 404, 500, 502, 503}
)

type event struct {
	Source    string                 `json:"source"`
	EventType string                `json:"event_type"`
	Severity  string                `json:"severity"`
	Timestamp string                `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Tags      []string               `json:"tags"`
}

type batch struct {
	Events []event `json:"events"`
}

func main() {
	baseURL := "http://localhost:8080"

	// Register a user and get a token
	token := register(baseURL)
	if token == "" {
		fmt.Println("Failed to get auth token, sending without auth")
	}

	fmt.Println("Generating events...")

	totalSent := 0
	for i := 0; i < 20; i++ {
		events := make([]event, 50)
		for j := range events {
			events[j] = randomEvent()
		}

		b := batch{Events: events}
		body, _ := json.Marshal(b)

		req, _ := http.NewRequest("POST", baseURL+"/api/v1/events", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		resp.Body.Close()

		totalSent += len(events)
		fmt.Printf("Batch %d: sent %d events (total: %d), status: %d\n",
			i+1, len(events), totalSent, resp.StatusCode)

		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("\nDone. Sent %d events total.\n", totalSent)
}

func register(baseURL string) string {
	body, _ := json.Marshal(map[string]string{
		"email":    fmt.Sprintf("demo-%d@atlasdb.dev", rand.Intn(100000)),
		"password": "demodemo123",
	})

	resp, err := http.Post(baseURL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if t, ok := result["access_token"].(string); ok {
		return t
	}
	return ""
}

func randomEvent() event {
	source := sources[rand.Intn(len(sources))]
	evtType := eventTypes[rand.Intn(len(eventTypes))]
	severity := weightedSeverity()
	ts := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)

	data := map[string]interface{}{}

	switch evtType {
	case "http_request":
		status := statuses[rand.Intn(len(statuses))]
		data["method"] = methods[rand.Intn(len(methods))]
		data["path"] = endpoints[rand.Intn(len(endpoints))]
		data["status"] = status
		data["duration_ms"] = rand.Intn(2000) + 5
		data["user_id"] = fmt.Sprintf("usr_%06d", rand.Intn(10000))
		if status >= 500 {
			severity = "error"
		} else if status >= 400 {
			severity = "warn"
		}
	case "error":
		severity = "error"
		data["error"] = randomError()
		data["stack_trace"] = "at handler.go:42\nat router.go:118"
	case "authentication":
		success := rand.Float32() > 0.1
		data["result"] = map[bool]string{true: "success", false: "failure"}[success]
		data["method"] = "password"
		data["ip"] = fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255))
		if !success {
			severity = "warn"
		}
	case "database_query":
		data["query"] = "SELECT * FROM orders WHERE user_id = $1"
		data["duration_ms"] = rand.Intn(500) + 1
		data["rows_affected"] = rand.Intn(100)
	default:
		data["detail"] = "operational event"
	}

	tags := []string{"production"}
	if rand.Float32() > 0.5 {
		tags = append(tags, "us-east-1")
	} else {
		tags = append(tags, "eu-west-1")
	}

	return event{
		Source:    source,
		EventType: evtType,
		Severity:  severity,
		Timestamp: ts.Format(time.RFC3339),
		Data:      data,
		Tags:      tags,
	}
}

func weightedSeverity() string {
	r := rand.Float32()
	switch {
	case r < 0.05:
		return "error"
	case r < 0.15:
		return "warn"
	case r < 0.30:
		return "debug"
	default:
		return "info"
	}
}

func randomError() string {
	errors := []string{
		"connection refused: payment-gateway:5432",
		"timeout waiting for response from inventory-service",
		"null pointer exception in OrderProcessor.process",
		"rate limit exceeded for external API",
		"circuit breaker open for downstream service",
		"out of memory: heap allocation failed",
		"deadlock detected in transaction",
		"TLS handshake timeout",
	}
	return errors[rand.Intn(len(errors))]
}
