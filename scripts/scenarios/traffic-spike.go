//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// Simulates a traffic spike scenario: normal baseline → sudden 10x spike → recovery.
func main() {
	baseURL := envOr("BASE_URL", "http://localhost:8080")
	token := login(baseURL)

	fmt.Println("=== Traffic Spike Scenario ===")
	fmt.Println("Phase 1: Normal traffic (2 min)")
	generateTraffic(baseURL, token, 5, 2*time.Minute)

	fmt.Println("Phase 2: SPIKE — 10x traffic (1 min)")
	generateTraffic(baseURL, token, 50, 1*time.Minute)

	fmt.Println("Phase 3: Recovery (2 min)")
	generateTraffic(baseURL, token, 5, 2*time.Minute)

	fmt.Println("=== Scenario Complete ===")
}

func generateTraffic(baseURL, token string, eventsPerSec int, duration time.Duration) {
	end := time.Now().Add(duration)
	interval := time.Second / time.Duration(eventsPerSec)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	total := 0
	for time.Now().Before(end) {
		<-ticker.C
		go ingest(baseURL, token, randomEvents(1+rand.Intn(5)))
		total++
	}
	fmt.Printf("  Sent %d batches\n", total)
}

func randomEvents(count int) []map[string]interface{} {
	sources := []string{"api-gateway", "user-service", "payment-service", "order-service"}
	types := []string{"http_request", "db_query", "cache_hit", "authentication"}
	severities := []string{"info", "info", "info", "warn", "error"}

	events := make([]map[string]interface{}, count)
	for i := range events {
		events[i] = map[string]interface{}{
			"source":     sources[rand.Intn(len(sources))],
			"event_type": types[rand.Intn(len(types))],
			"severity":   severities[rand.Intn(len(severities))],
			"data": map[string]interface{}{
				"status_code": 200,
				"duration_ms": math.Round(rand.Float64()*200*100) / 100,
				"path":        fmt.Sprintf("/api/v1/items/%d", rand.Intn(10000)),
			},
			"tags": []string{"demo", "traffic-spike"},
		}
	}
	return events
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

	var result struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.AccessToken
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
