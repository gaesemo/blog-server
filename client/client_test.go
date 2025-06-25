package client

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"connectrpc.com/connect"
	authv1 "github.com/gaesemo/blog-api/go/service/auth/v1"
	typesv1 "github.com/gaesemo/blog-api/go/types/v1"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	config := NewConfig()
	client := New(*config)
	resp, err := client.Auth.GetAuthURL(ctx, connect.NewRequest(&authv1.GetAuthURLRequest{
		IdentityProvider: typesv1.IdentityProvider_IDENTITY_PROVIDER_GITHUB,
		RedirectUrl:      nil,
	}))
	if err != nil {
		logger.ErrorContext(ctx, "get auth url", slog.Any("error", err))
		require.NoError(t, err)
	}
	logger.InfoContext(ctx, "get auth url", slog.Any("resp", resp))
}
