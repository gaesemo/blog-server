package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	authv1 "github.com/gaesemo/tech-blog-api/go/service/auth/v1"
	"github.com/gaesemo/tech-blog-api/go/service/auth/v1/authv1connect"
	typesv1 "github.com/gaesemo/tech-blog-api/go/types/v1"
	"github.com/gaesemo/tech-blog-server/gen/db/postgres"
	"github.com/gaesemo/tech-blog-server/pkg/oauth"
	"github.com/gaesemo/tech-blog-server/pkg/transaction"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var _ authv1connect.AuthServiceHandler = (*service)(nil)

func New(
	logger *slog.Logger,
	httpClient *http.Client,
	db *pgx.Conn,
	timeNow func() time.Time,
	randStr func() string,
	opts ...OAuthAppOption,
) authv1connect.AuthServiceHandler {
	svc := &service{
		logger:     logger,
		httpClient: httpClient,
		db:         db,
		queries:    postgres.New(db),
		timeNow:    timeNow,
		randStr:    randStr,
		oauthApps:  map[string]oauth.App{},
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

type OAuthAppOption func(cfgs map[string]oauth.App)

type service struct {
	logger     *slog.Logger
	db         *pgx.Conn
	queries    *postgres.Queries
	httpClient *http.Client
	oauthApps  map[string]oauth.App
	timeNow    func() time.Time
	randStr    func() string
}

func (svc *service) GetAuthURL(ctx context.Context, req *connect.Request[authv1.GetAuthURLRequest]) (*connect.Response[authv1.GetAuthURLResponse], error) {

	identityProvider := req.Msg.IdentityProvider
	redirectUrl := req.Msg.RedirectUrl

	oauthApp, err := svc.getOAuthApp(identityProvider)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	authURL, err := oauthApp.GetAuthURL(&oauth.GetAuthURLOption{
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
	redirectURL := req.Msg.RedirectUrl

	oauthApp, err := svc.getOAuthApp(identityProvider)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	accessToken, err := oauthApp.ExchangeCode(code, &oauth.ExchangeCodeOption{RedirectURL: redirectURL})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	profile, err := oauthApp.GetUserProfile(accessToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	tx := transaction.New[postgres.User](
		svc.db,
		pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadWrite},
		svc.queries,
	)

	user, txErr := tx.Exec(ctx, func(c context.Context, q *postgres.Queries) (*postgres.User, error) {
		u, err := q.GetUserByEmailAndIDP(ctx, postgres.GetUserByEmailAndIDPParams{
			Email:            profile.Email,
			IdentityProvider: typesv1.IdentityProvider_name[int32(identityProvider)],
		})
		if err == nil {
			return &u, nil
		}
		if errors.Is(err, pgx.ErrNoRows) {
			u, err := q.CreateUser(ctx, postgres.CreateUserParams{
				IdentityProvider: typesv1.IdentityProvider_name[int32(identityProvider)],
				Email:            profile.Email,
				Username:         profile.TempName(),
				AvatarUrl:        profile.AvatarURL,
				AboutMe:          "",
				CreatedAt:        pgtype.Timestamptz{Time: svc.timeNow(), Valid: true},
				UpdatedAt:        pgtype.Timestamptz{Time: svc.timeNow(), Valid: true},
			})
			return &u, err
		}
		return nil, err
	})
	if txErr != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("executing tx: %v", txErr))
	}
	slog.InfoContext(ctx, "user", slog.Int64("id", user.ID))
	// TODO: issue jwt token, create new session
	return nil, nil
}

func (svc *service) Logout(ctx context.Context, req *connect.Request[authv1.LogoutRequest]) (*connect.Response[authv1.LogoutResponse], error) {
	return nil, nil
}

func (svc *service) getOAuthApp(identityProvider typesv1.IdentityProvider) (oauth.App, error) {
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

func WithGitHubOAuthApp(app oauth.App) OAuthAppOption {
	return func(oa map[string]oauth.App) {
		_, exists := oa[github]
		if !exists {
			oa[github] = app
		}
	}
}

func WithOAuthApp(provider string, app oauth.App) OAuthAppOption {
	return func(oa map[string]oauth.App) {
		_, exists := oa[provider]
		if !exists {
			oa[provider] = app
		}
	}
}
