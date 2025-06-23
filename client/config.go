package client

import (
	"net/http"
	"time"
)

type ConfigOption func(*Config)

type Config struct {
	BaseURL    string
	HTTPClient *http.Client
}

const (
	BaseUrl           = "http://localhost:8080"
	ConnectionTimeout = 10 * time.Second
)

func NewConfig(opts ...ConfigOption) *Config {
	config := &Config{
		BaseURL:    BaseUrl,
		HTTPClient: &http.Client{Timeout: ConnectionTimeout},
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

func WithHTTPClient(client *http.Client) ConfigOption {
	return func(c *Config) {
		c.HTTPClient = client
	}
}

func WithTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		if c.HTTPClient == nil {
			c.HTTPClient = &http.Client{}
		}
		c.HTTPClient.Timeout = timeout
	}
}

func WithBaseURL(baseURL string) ConfigOption {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}
