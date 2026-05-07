package auth

import (
	"fmt"
	"net/http"
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

func (b *BearerAuth) BasicApply(req *http.Request) error {
	if b.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", b.Token))
	}
	return nil
}
