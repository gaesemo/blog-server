package authn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/authn"
	"connectrpc.com/connect"
	"github.com/gaesemo/blog-server/pkg/token"
	"github.com/golang-jwt/jwt/v5"
)

type Authenticator interface {
	Authenticate(ctx context.Context, req *http.Request) (any, error)
}

func Authenticate(ctx context.Context, req *http.Request) (any, error) {
	str, ok := authn.BearerToken(req)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("extracting token: not found"))
	}
	claims := token.NewUserClaims()
	tok, err := token.ParseWithClaims(str, claims)
	if err != nil && errors.Is(err, jwt.ErrTokenMalformed) {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("token malformed: %v", err))
	}
	if err != nil && errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid signature"))
	}
	if !tok.Valid {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid token"))
	}
	now := time.Now()
	nbf, _ := tok.Claims.GetNotBefore()
	if now.Before(nbf.Time) {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("token not yet active"))
	}

	exp, _ := tok.Claims.GetExpirationTime()
	if now.After(exp.Time) {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("token expired"))
	}

	return claims.UserID, nil
}

func NewLoggingInterceptor() connect.UnaryInterceptorFunc {
	return loggingInterceptor
}

func loggingInterceptor(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		procedure := req.Spec().Procedure

		slog.InfoContext(ctx, "request", slog.String("calling", procedure), slog.String("method", req.HTTPMethod()), slog.Any("headers", req.Header()), slog.Any("any", req.Any()))
		resp, err := next(ctx, req)
		if err != nil {
			slog.ErrorContext(ctx, "response", slog.String("calling", procedure), slog.Any("error", err))
		} else {
			slog.InfoContext(ctx, "response", slog.String("calling", procedure), slog.Any("headers", resp.Header()), slog.Any("any", resp.Any()))
		}
		return resp, nil
	}
}
