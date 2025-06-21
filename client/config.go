package client

import (
	"net/http"
	"time"
)

type Option func(*Config)

func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) {
		c.HTTPClient = client
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		if c.HTTPClient == nil {
			c.HTTPClient = &http.Client{}
		}
		c.HTTPClient.Timeout = timeout
	}
}

func WithBaseURL(baseURL string) Option {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

func NewConfig(opts ...Option) *Config {
	config := &Config{
		BaseURL:    "http://localhost:8080",
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
