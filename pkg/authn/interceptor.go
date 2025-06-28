package authn

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/gaesemo/blog-server/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type Authenticator interface {
	Authenticate(ctx context.Context, req *http.Request) (any, error)
}

func Authenticate(ctx context.Context, req *http.Request) (any, error) {
	var tok string
	if cookie, err := req.Cookie("tok"); err == nil {
		tok = cookie.Value
	} else {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("auth token not provided"))
	}

	claims := token.NewClaim()
	token, err := jwt.ParseWithClaims(tok, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(viper.GetString("JWT_SIGNING_SECRET")), nil
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("parsing token: %v", err))
	}
	if !token.Valid {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid token"))
	}
	return nil, nil
}
