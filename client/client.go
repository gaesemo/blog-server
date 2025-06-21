package client

import (
	"net/http"

	"github.com/gaesemo/tech-blog-api/go/service/auth/v1/authv1connect"
	"github.com/gaesemo/tech-blog-api/go/service/user/v1/userv1connect"
)

type Client struct {
	Auth authv1connect.AuthServiceClient
	User userv1connect.UserServiceClient
}

type Config struct {
	BaseURL    string
	HTTPClient *http.Client
}

func New(config Config) *Client {
	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}

	return &Client{
		Auth: authv1connect.NewAuthServiceClient(config.HTTPClient, config.BaseURL),
		User: userv1connect.NewUserServiceClient(config.HTTPClient, config.BaseURL),
	}
}
