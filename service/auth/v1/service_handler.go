package auth

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	authv1 "github.com/gaesemo/tech-blog-api/go/service/auth/v1"
	"github.com/gaesemo/tech-blog-api/go/service/auth/v1/authv1connect"
	typesv1 "github.com/gaesemo/tech-blog-api/go/types/v1"
	"golang.org/x/oauth2/endpoints"
)

var _ authv1connect.AuthServiceHandler = (*authServiceHandler)(nil)

func New(logger *slog.Logger) authv1connect.AuthServiceHandler {
	return &authServiceHandler{
		logger,
	}
}

type authServiceHandler struct {
	logger *slog.Logger
}

func (h *authServiceHandler) GetAuthURL(ctx context.Context, req *connect.Request[authv1.GetAuthURLRequest]) (*connect.Response[authv1.GetAuthURLResponse], error) {
	identityProvider := req.Msg.IdentityProvider
	switch identityProvider {
	case typesv1.IdentityProvider_IDENTITY_PROVIDER_UNSPECIFIED:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("identity provider must be specified"))
	case typesv1.IdentityProvider_IDENTITY_PROVIDER_GITHUB:
		return connect.NewResponse(&authv1.GetAuthURLResponse{
			AuthUrl: endpoints.GitHub.AuthURL, // TODO: add ?client_id=xxx&...
		}), nil
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unsupported auth provider: %s", typesv1.IdentityProvider_name[int32(identityProvider)]))
	}
}

func (h *authServiceHandler) Login(ctx context.Context, req *connect.Request[authv1.LoginRequest]) (*connect.Response[authv1.LoginResponse], error) {
	return nil, nil
}

func (h *authServiceHandler) Logout(ctx context.Context, req *connect.Request[authv1.LogoutRequest]) (*connect.Response[authv1.LogoutResponse], error) {
	return nil, nil
}
