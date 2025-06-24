package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	authv1 "github.com/gaesemo/tech-blog-api/go/service/auth/v1"
	"github.com/gaesemo/tech-blog-api/go/service/auth/v1/authv1connect"
	typesv1 "github.com/gaesemo/tech-blog-api/go/types/v1"
	"github.com/gaesemo/tech-blog-server/pkg/oauthapp"
	"github.com/jackc/pgx/v5"
)

var _ authv1connect.AuthServiceHandler = (*service)(nil)

func New(
	logger *slog.Logger,
	db *pgx.Conn,
	httpClient *http.Client,
	timeNow func() time.Time,
	randStr func() string,
	opts ...OAuthAppOption,
) authv1connect.AuthServiceHandler {
	svc := &service{
		logger: logger,

		db:         db,
		httpClient: httpClient,
		oauthApps:  map[string]oauthapp.OAuthApp{},

		timeNow: timeNow,
		randStr: randStr,
	}

	for _, o := range opts {
		o(svc.oauthApps)
	}

	return svc
}

const (
	github = "github"
	google = "google"
)

type OAuthAppOption func(cfgs map[string]oauthapp.OAuthApp)

type service struct {
	logger *slog.Logger

	db         *pgx.Conn
	httpClient *http.Client
	oauthApps  map[string]oauthapp.OAuthApp

	timeNow func() time.Time
	randStr func() string
}

func (svc *service) GetAuthURL(ctx context.Context, req *connect.Request[authv1.GetAuthURLRequest]) (*connect.Response[authv1.GetAuthURLResponse], error) {

	identityProvider := req.Msg.IdentityProvider
	redirectUrl := req.Msg.RedirectUrl

	oauthApp, err := svc.getOAuthApp(identityProvider)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	authURL, err := oauthApp.GetAuthURL(&oauthapp.GetAuthURLOption{
		RedirectURL: redirectUrl,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&authv1.GetAuthURLResponse{
		AuthUrl: authURL,
	}), nil
}

func (svc *service) Login(ctx context.Context, req *connect.Request[authv1.LoginRequest]) (*connect.Response[authv1.LoginResponse], error) {
	identityProvider := req.Msg.IdentityProvider
	code := req.Msg.AuthCode

	oauthApp, err := svc.getOAuthApp(identityProvider)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	accessToken, err := oauthApp.ExchangeCode(code, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	profile, err := oauthApp.GetUserProfile(accessToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	slog.InfoContext(ctx, "user profile", slog.String("email", profile.Email), slog.String("avatar", profile.AvatarURL))

	return nil, nil
}

func (svc *service) Logout(ctx context.Context, req *connect.Request[authv1.LogoutRequest]) (*connect.Response[authv1.LogoutResponse], error) {
	return nil, nil
}

func (svc *service) getOAuthApp(identityProvider typesv1.IdentityProvider) (oauthapp.OAuthApp, error) {
	switch identityProvider {
	case typesv1.IdentityProvider_IDENTITY_PROVIDER_GITHUB:
		oa, exist := svc.oauthApps[github]
		if !exist {
			return nil, fmt.Errorf("unsupported identity provider: github")
		}
		return oa, nil
	case typesv1.IdentityProvider_IDENTITY_PROVIDER_UNSPECIFIED:
		return nil, fmt.Errorf("identity provider unspecified")
	default:
		return nil, fmt.Errorf("unsupported identity provider")
	}
}

func WithGitHubOAuthApp(app oauthapp.OAuthApp) OAuthAppOption {
	return func(oa map[string]oauthapp.OAuthApp) {
		_, exists := oa[github]
		if !exists {
			oa[github] = app
		}
	}
}

func WithOAuthApp(provider string, app oauthapp.OAuthApp) OAuthAppOption {
	return func(oa map[string]oauthapp.OAuthApp) {
		_, exists := oa[provider]
		if !exists {
			oa[provider] = app
		}
	}
}
