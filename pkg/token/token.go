package token

import (
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

var (
	signingMethod = jwt.SigningMethodHS256
	signingSecret = viper.GetString("JWT_SIGNING_SECRET")
)

var _ jwt.Claims = (*UserClaims)(nil)

type UserClaims struct {
	Audience       []string  `json:"aud"`
	Issuer         string    `json:"iss"`
	IssuedAt       time.Time `json:"iat"`
	ExpirationTime time.Time `json:"exp"`
	NotBefore      time.Time `json:"nbf"`
	UserID         int64     `json:"uid"`
}

func NewUserClaims() *UserClaims {
	return &UserClaims{
		Audience:       []string{},
		Issuer:         "",
		IssuedAt:       time.Time{},
		ExpirationTime: time.Time{},
		NotBefore:      time.Time{},
		UserID:         0,
	}
}

func NewWithUserClaims(claims UserClaims) *jwt.Token {
	return jwt.NewWithClaims(signingMethod, claims)
}

func ParseWithClaims(tok string, claims *UserClaims) (*jwt.Token, error) {
	keyFunc := func(t *jwt.Token) (any, error) {
		return signingSecret, nil
	}
	t, err := jwt.ParseWithClaims(tok, claims, keyFunc)
	if err != nil {
		slog.Error("parsing jwt token", slog.Any("error", err))
		return nil, err
	}
	return t, nil
}

// GetAudience implements jwt.Claims.
func (u UserClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings(u.Audience), nil
}

// GetExpirationTime implements jwt.Claims.
func (u UserClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(u.ExpirationTime), nil
}

// GetIssuedAt implements jwt.Claims.
func (u UserClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(u.IssuedAt), nil
}

// GetIssuer implements jwt.Claims.
func (u UserClaims) GetIssuer() (string, error) {
	return u.Issuer, nil
}

// GetNotBefore implements jwt.Claims.
func (u UserClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(u.NotBefore), nil
}

// GetSubject implements jwt.Claims.
func (u UserClaims) GetSubject() (string, error) {
	return "", nil
}
