package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/gaesemo/blog-server/pkg/token"
	"github.com/golang-jwt/jwt/v5"
)

func Authorize(ctx context.Context, req *http.Request) (any, error) {
	cookie, err := req.Cookie("token")
	if err != nil {
		slog.InfoContext(ctx, "author not found")
		return nil, nil
	}
	tokenValue := cookie.Value
	claims := token.NewUserClaims()
	tok, err := token.ParseWithClaims(tokenValue, claims)
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
	return &claims.UserID, nil
}
