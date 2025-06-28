package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	authv1 "github.com/gaesemo/blog-api/go/service/auth/v1"
	"github.com/gaesemo/blog-api/go/service/auth/v1/authv1connect"
	typesv1 "github.com/gaesemo/blog-api/go/types/v1"
	"github.com/gaesemo/blog-server/gen/db/postgres"
	"github.com/gaesemo/blog-server/pkg/oauth"
	"github.com/gaesemo/blog-server/pkg/transaction"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spf13/viper"
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

	oauthApp, err := svc.getOAuthApp(identityProvider)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	authURL, err := oauthApp.GetAuthURL()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&authv1.GetAuthURLResponse{
		AuthUrl: authURL,
	}), nil
}

func (svc *service) Login(ctx context.Context, req *connect.Request[authv1.LoginRequest]) (*connect.Response[authv1.LoginResponse], error) {
	ll := svc.logger.With("login", req.Spec().Procedure)
	identityProvider := req.Msg.IdentityProvider
	code := req.Msg.Code

	oauthApp, _ := svc.getOAuthApp(identityProvider)

	ll.DebugContext(ctx, "exchanging temporary code with access token", slog.String("code", code))
	accessToken, err := oauthApp.ExchangeCode(code)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("exchaning code: %v", err))
	}

	profile, err := oauthApp.GetUserProfile(accessToken)
	if err != nil {
		return nil, fmt.Errorf("denied: %v", err)
	}

	type Result struct {
		User      *postgres.User
		IsNewUser bool
	}

	tx := transaction.New[Result](
		svc.db,
		pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadWrite},
		svc.queries,
	)

	result, txErr := tx.Exec(ctx, func(c context.Context, q *postgres.Queries) (*Result, error) {
		u, err := q.GetUserByEmailAndIDP(ctx, postgres.GetUserByEmailAndIDPParams{
			Email:            profile.Email,
			IdentityProvider: typesv1.IdentityProvider_name[int32(typesv1.IdentityProvider_IDENTITY_PROVIDER_GITHUB)],
		})
		if err == nil {
			return &Result{User: &u, IsNewUser: false}, nil
		}
		if errors.Is(err, pgx.ErrNoRows) {
			u, err := q.CreateUser(ctx, postgres.CreateUserParams{
				IdentityProvider: typesv1.IdentityProvider_name[int32(typesv1.IdentityProvider_IDENTITY_PROVIDER_GITHUB)],
				Email:            profile.Email,
				Username:         profile.Name,
				AvatarUrl:        profile.AvatarURL,
				AboutMe:          "",
				CreatedAt:        pgtype.Timestamptz{Time: svc.timeNow(), Valid: true},
				UpdatedAt:        pgtype.Timestamptz{Time: svc.timeNow(), Valid: true},
			})
			return &Result{User: &u, IsNewUser: true}, err
		}
		return nil, err
	})
	if txErr != nil {
		return nil, fmt.Errorf("in login flow: %v", err)
	}

	user := result.User
	isNewUser := result.IsNewUser
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "gsm",
		"iat": jwt.NewNumericDate(svc.timeNow()),
		"exp": jwt.NewNumericDate(svc.timeNow().Add(time.Hour)),
		"nbf": jwt.NewNumericDate(svc.timeNow()),
		"uid": user.ID,
	})
	gsmAccessToken, err := token.SignedString([]byte(viper.GetString("JWT_SIGNING_SECRET")))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("signing token: %v", err))
	}
	resp := connect.NewResponse(&authv1.LoginResponse{
		Token:     gsmAccessToken,
		IsNewUser: isNewUser,
	})

	cookie := &http.Cookie{
		Name:     "token",
		Value:    gsmAccessToken,
		Expires:  svc.timeNow().Add(time.Hour),
		MaxAge:   3600,
		HttpOnly: true,
		Path:     "/",
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	resp.Header().Set("Set-Cookie", cookie.String())
	return resp, nil
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
