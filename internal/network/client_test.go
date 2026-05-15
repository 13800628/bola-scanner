package network

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRetryClient_Do(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &RetryClient{
		MaxRetries: 3,
		// 待機時間などはあとで追加
	}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	// 検証
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls (2 retries + 1 success), got %d", callCount)
	}
}

func TestRetryClient_CalculateBackoff(t *testing.T) {
	client := &RetryClient{
		BaseDelay: 100 * time.Millisecond,
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{attempt: 0, expected: 0},
		{attempt: 1, expected: 100 * time.Millisecond},
		{attempt: 2, expected: 200 * time.Millisecond},
		{attempt: 3, expected: 400 * time.Millisecond},
	}

	for _, tt := range tests {
		actual := client.CalculateBackoff(tt.attempt)
		if actual != tt.expected {
			t.Errorf("attempt %d: expected %v, got %v", tt.attempt, tt.expected, actual)
		}
	}
}
