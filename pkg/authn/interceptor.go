package authn

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/authn"
	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
)

type Authenticator interface {
	Authenticate(ctx context.Context, req *http.Request) (any, error)
}

func Authenticate(ctx context.Context, req *http.Request) (any, error) {
	tokenStr, ok := authn.BearerToken(req)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("bearer token required"))
	}
	keyFunc := func(t *jwt.Token) (any, error) {
		return nil, nil
	}
	_, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, keyFunc)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("parsing token: %v", err))
	}
	return nil, nil
}
