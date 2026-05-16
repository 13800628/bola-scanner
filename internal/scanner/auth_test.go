package scanner

import (
	"net/http"
	"testing"

	"github.com/yuto-isayama/bola-scanner/internal/auth"
)

type MockAuthenticator struct{}

func (m *MockAuthenticator) Apply(req *http.Request) error {
	req.Header.Set("X-Test-Auth", "authenticated")
	return nil
}

func TestScanner_AuthApplication(t *testing.T) {
	// テスト用のMockAuthenticatorをマップに
	mockAuth := &MockAuthenticator{}
	authMap := map[string]auth.Authenticator{
		"attacker-1": mockAuth,
	}

	s := &Scanner{
		AuthMap: authMap,
	}
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	// マップから特定の認証情報を取り出す
	targetAuth, exists := s.AuthMap["attacker-1"]
	if !exists {
		t.Fatalf("Expected authenticator for 'attacker-1' to exist")
	}

	err := targetAuth.Apply(req)

	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if req.Header.Get("X-Test-Auth") != "authenticated" {
		t.Errorf("Auth header was not set correnctly")
	}
}
