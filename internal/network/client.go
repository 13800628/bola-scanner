package network

import (
	"net/http"
	"time"
)

type RetryClient struct {
	MaxRetries int
	Client     http.Client
}

func (c *RetryClient) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i < c.MaxRetries; i++ {
		resp, err = c.Client.Do(req)

		if err == nil && resp.StatusCode != http.StatusServiceUnavailable && resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		// 簡易的な指数バックオフの実装(今後はここだけで一つの関数に切り出す)
		time.Sleep(time.Duration(i) * 10 * time.Millisecond)
	}
	return resp, err
}
