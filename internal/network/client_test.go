package network

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
