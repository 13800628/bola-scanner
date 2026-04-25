package scanner

import (
	"net/http"
	"testing"
)

type MockAuthenticator struct{}

func (m *MockAuthenticator) Apply(req *http.Request) error {
	req.Header.Set("X-Test-Auth", "authenticated")
	return nil
}

func TestScanner_AuthApplication(t *testing.T) {
	s := &Scanner{
		Auth: &MockAuthenticator{},
	}
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	err := s.Auth.Apply(req)

	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if req.Header.Get("X-Test-Auth") != "authenticated" {
		t.Errorf("Auth header was not set correnctly")
	}
}
