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
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthService interface {
	authv1connect.AuthServiceHandler
	GitHubCallback(ctx context.Context, code string, redirectURL string) (string, error)
}

var _ AuthService = (*service)(nil)

func New(
	logger *slog.Logger,
	httpClient *http.Client,
	db *pgx.Conn,
	timeNow func() time.Time,
	randStr func() string,
	opts ...OAuthAppOption,
) AuthService {
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

func (svc *service) GitHubCallback(ctx context.Context, code string, redirectURL string) (string, error) {

	oauthApp, _ := svc.oauthApps[github]
	accessToken, err := oauthApp.ExchangeCode(code, &oauth.ExchangeCodeOption{RedirectURL: &redirectURL})
	if err != nil {
		return "", fmt.Errorf("denied: %v", err)
	}

	profile, err := oauthApp.GetUserProfile(accessToken)
	if err != nil {
		return "", fmt.Errorf("denied: %v", err)
	}

	tx := transaction.New[postgres.User](
		svc.db,
		pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadWrite},
		svc.queries,
	)

	user, txErr := tx.Exec(ctx, func(c context.Context, q *postgres.Queries) (*postgres.User, error) {
		u, err := q.GetUserByEmailAndIDP(ctx, postgres.GetUserByEmailAndIDPParams{
			Email:            profile.Email,
			IdentityProvider: typesv1.IdentityProvider_name[int32(typesv1.IdentityProvider_IDENTITY_PROVIDER_GITHUB)],
		})
		if err == nil {
			return &u, nil
		}
		if errors.Is(err, pgx.ErrNoRows) {
			u, err := q.CreateUser(ctx, postgres.CreateUserParams{
				IdentityProvider: typesv1.IdentityProvider_name[int32(typesv1.IdentityProvider_IDENTITY_PROVIDER_GITHUB)],
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
		return "", fmt.Errorf("internal: %v", err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user.ID,
		"ttl":  svc.timeNow().Add(time.Hour).Unix(),
	})
	gsmAuthToken, err := token.SignedString("some-secret")
	if err != nil {
		return "", fmt.Errorf("internal: %v", err)
	}
	return gsmAuthToken, nil
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
