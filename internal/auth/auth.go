package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// 共通の認証インターフェース
type Authenticator interface {
	Apply(req *http.Request) error
}

// 認証なし(プロトタイプ用)
type NoAuth struct{}

func (n *NoAuth) Apply(req *http.Request) error {
	return nil
}

// ベーシック認証
type BearerAuth struct {
	Token string
}

func (b *BearerAuth) Apply(req *http.Request) error {
	if b.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", b.Token))
	}
	return nil
}

type Profile struct {
	Username string
	Password string
	UserID   int
	Token    string
	Client   *http.Client
}

func NewProfile(username, password string) *Profile {
	return &Profile{
		Username: username,
		Password: password,
		Client:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (p *Profile) Login(baseURL string) error {
	loginURL := fmt.Sprintf("%s/login", baseURL)

	payload := map[string]string{
		"username": p.Username,
		"password": p.Password,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := p.Client.Post(loginURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: status %d", resp.StatusCode)
	}

	var result struct {
		Token  string `json:"token"`
		UserID int    `json:"user_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	p.Token = result.Token

	if result.UserID != 0 {
		p.UserID = result.UserID
	}
	return nil
}

func (p *Profile) GetAuthenticator() Authenticator {
	if p.Token != "" {
		return &BearerAuth{Token: p.Token}
	}
	return &NoAuth{}
}
