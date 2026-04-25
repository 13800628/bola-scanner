package auth

import "net/http"

type Authenticator interface {
	Apply(req *http.Request) error
}

// のちに実装
type NoAuth struct{}

func (n *NoAuth) Apply(req *http.Request) error {
	return nil
}
