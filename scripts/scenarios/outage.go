//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// Simulates a service outage scenario:
// 1. Normal operations
// 2. payment-service starts throwing errors (degradation)
// 3. Full outage — cascading failures across services
// 4. Recovery with deploy event
func main() {
	baseURL := envOr("BASE_URL", "http://localhost:8080")
	token := login(baseURL)

	fmt.Println("=== Service Outage Scenario ===")

	fmt.Println("Phase 1: Normal operations (1 min)")
	phase(baseURL, token, 1*time.Minute, func() []map[string]interface{} {
		return healthyEvents()
	})

	fmt.Println("Phase 2: Degradation — payment-service errors (1 min)")
	phase(baseURL, token, 1*time.Minute, func() []map[string]interface{} {
		events := healthyEvents()
		events = append(events, errorEvent("payment-service", "connection_timeout", "Database connection pool exhausted"))
		events = append(events, errorEvent("payment-service", "db_query", "Connection refused to postgres:5432"))
		return events
	})

	fmt.Println("Phase 3: Cascade — multiple services failing (1 min)")
	phase(baseURL, token, 1*time.Minute, func() []map[string]interface{} {
		return []map[string]interface{}{
			errorEvent("payment-service", "connection_timeout", "Database connection pool exhausted"),
			errorEvent("order-service", "dependency_failure", "payment-service returned 503"),
			errorEvent("api-gateway", "upstream_error", "order-service circuit breaker open"),
			errorEvent("notification-service", "queue_backlog", "Message queue depth exceeds 10000"),
			{
				"source": "api-gateway", "event_type": "http_request", "severity": "error",
				"data": map[string]interface{}{"status_code": 503, "path": "/api/v1/checkout", "error": "service unavailable"},
				"tags": []string{"demo", "outage"},
			},
		}
	})

	fmt.Println("Phase 4: Deploy fix + recovery (1 min)")
	// Deployment event
	ingest(baseURL, token, []map[string]interface{}{{
		"source": "deployment-service", "event_type": "deployment", "severity": "info",
		"data": map[string]interface{}{
			"service":     "payment-service",
			"version":     "v2.3.2",
			"change":      "Fix connection pool leak",
			"deployed_by": "oncall-engineer",
		},
		"tags": []string{"demo", "outage", "deploy"},
	}})

	phase(baseURL, token, 1*time.Minute, func() []map[string]interface{} {
		events := healthyEvents()
		// Occasional residual errors clearing
		if rand.Float64() > 0.7 {
			events = append(events, map[string]interface{}{
				"source": "payment-service", "event_type": "recovery", "severity": "warn",
				"data": map[string]interface{}{"message": "Connection pool recovering", "active_connections": rand.Intn(20)},
				"tags": []string{"demo", "outage"},
			})
		}
		return events
	})

	fmt.Println("=== Scenario Complete ===")
}

func phase(baseURL, token string, duration time.Duration, eventFn func() []map[string]interface{}) {
	end := time.Now().Add(duration)
	for time.Now().Before(end) {
		ingest(baseURL, token, eventFn())
		time.Sleep(200 * time.Millisecond)
	}
}

func healthyEvents() []map[string]interface{} {
	services := []string{"api-gateway", "user-service", "payment-service", "order-service"}
	events := make([]map[string]interface{}, 3)
	for i := range events {
		events[i] = map[string]interface{}{
			"source":     services[rand.Intn(len(services))],
			"event_type": "http_request",
			"severity":   "info",
			"data": map[string]interface{}{
				"status_code": 200,
				"duration_ms": 20 + rand.Float64()*80,
			},
			"tags": []string{"demo", "outage"},
		}
	}
	return events
}

func errorEvent(source, eventType, message string) map[string]interface{} {
	return map[string]interface{}{
		"source":     source,
		"event_type": eventType,
		"severity":   "error",
		"data":       map[string]interface{}{"error": message, "duration_ms": 5000 + rand.Float64()*25000},
		"tags":       []string{"demo", "outage"},
	}
}

func ingest(baseURL, token string, events []map[string]interface{}) {
	body, _ := json.Marshal(map[string]interface{}{"events": events})
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	http.DefaultClient.Do(req)
}

func login(baseURL string) string {
	email := fmt.Sprintf("demo-%d@atlas.dev", time.Now().UnixMilli())
	body, _ := json.Marshal(map[string]string{"email": email, "password": "demo123456"})
	http.Post(baseURL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	resp, err := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	var result struct{ AccessToken string `json:"access_token"` }
	json.NewDecoder(resp.Body).Decode(&result)
	return result.AccessToken
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
