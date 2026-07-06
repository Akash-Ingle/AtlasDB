//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
)

var baseURL string

func TestMain(m *testing.M) {
	baseURL = os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	os.Exit(m.Run())
}

func TestHealthEndpoints(t *testing.T) {
	t.Run("liveness", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/healthz")
		if err != nil {
			t.Fatalf("healthz request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("readiness", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/readyz")
		if err != nil {
			t.Fatalf("readyz request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestAuthFlow(t *testing.T) {
	email := fmt.Sprintf("inttest-%d@atlas.dev", os.Getpid())
	password := "integrationtest123"

	t.Run("register", func(t *testing.T) {
		body := jsonBody(map[string]string{"email": email, "password": password})
		resp, err := http.Post(baseURL+"/api/v1/auth/register", "application/json", body)
		if err != nil {
			t.Fatalf("register failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 201 {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	var token string
	t.Run("login", func(t *testing.T) {
		body := jsonBody(map[string]string{"email": email, "password": password})
		resp, err := http.Post(baseURL+"/api/v1/auth/login", "application/json", body)
		if err != nil {
			t.Fatalf("login failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var result struct {
			AccessToken string `json:"access_token"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		token = result.AccessToken
		if token == "" {
			t.Fatal("empty access token")
		}
	})

	t.Run("ingest_events", func(t *testing.T) {
		events := []map[string]interface{}{
			{
				"source":     "integration-test",
				"event_type": "test_event",
				"severity":   "info",
				"data":       map[string]interface{}{"test": true, "run_id": os.Getpid()},
				"tags":       []string{"integration-test"},
			},
		}
		body := jsonBody(map[string]interface{}{"events": events})
		req, _ := http.NewRequest("POST", baseURL+"/api/v1/events", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("ingest failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 202 {
			t.Fatalf("expected 202, got %d", resp.StatusCode)
		}

		var result struct {
			Accepted int `json:"accepted"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		if result.Accepted != 1 {
			t.Fatalf("expected 1 accepted, got %d", result.Accepted)
		}
	})

	t.Run("list_events", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/api/v1/events?limit=5", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("list events failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("analytics_summary", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/api/v1/analytics/summary?range=1h", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("analytics failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("alert_rules", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/api/v1/alerts/rules", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("alert rules failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthenticated_rejected", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/api/v1/events", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 401 {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})
}

func jsonBody(v interface{}) *bytes.Buffer {
	data, _ := json.Marshal(v)
	return bytes.NewBuffer(data)
}
