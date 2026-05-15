package network

import (
	"math"
	"net/http"
	"time"
)

type RetryClient struct {
	MaxRetries int
	BaseDelay  time.Duration
	Client     http.Client
}

func (c *RetryClient) CalculateBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	multiplier := math.Pow(2, float64(attempt-1))
	return time.Duration(float64(c.BaseDelay) * multiplier)
}

func (c *RetryClient) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i < c.MaxRetries; i++ {
		if i > 0 {
			time.Sleep(c.CalculateBackoff(i))
		}

		resp, err = c.Client.Do(req)
		if !c.ShouldRetry(resp, err) {
			return resp, err
		}
	}
	return resp, err
}

func (c *RetryClient) ShouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	if resp == nil {
		return false
	}
	return resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode >= 500 && resp.StatusCode <= 599)
}
