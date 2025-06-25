package client

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gaesemo/blog-api/go/service/auth/v1/authv1connect"
	"github.com/gaesemo/blog-api/go/service/user/v1/userv1connect"
)

type Option func(*Client)
type Client struct {
	logger *slog.Logger
	Auth   authv1connect.AuthServiceClient
	User   userv1connect.UserServiceClient
}

func Default() *Client {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
	config := NewConfig() // use default
	return &Client{
		logger: logger,
		Auth:   authv1connect.NewAuthServiceClient(config.HTTPClient, config.BaseURL),
		User:   userv1connect.NewUserServiceClient(config.HTTPClient, config.BaseURL),
	}
}

func New(config Config, opts ...Option) *Client {

	c := &Client{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		})),
	}

	for _, opt := range opts {
		opt(c)
	}

	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}

	c.Auth = authv1connect.NewAuthServiceClient(config.HTTPClient, config.BaseURL)
	c.User = userv1connect.NewUserServiceClient(config.HTTPClient, config.BaseURL)

	return c
}
