package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var _ jwt.Claims = (*UserClaims)(nil)

type UserClaims struct {
	UserID   int64
	Username string
	// standard claims
	Audience       []string
	Issuer         string
	IssuedAt       time.Time
	ExpirationTime time.Time
	NotBefore      time.Time
}

func NewClaim() *UserClaims {
	return &UserClaims{
		Audience:       []string{},
		Issuer:         "",
		IssuedAt:       time.Time{},
		ExpirationTime: time.Time{},
		NotBefore:      time.Time{},
		UserID:         0,
		Username:       "",
	}
}

// GetAudience implements jwt.Claims.
func (u *UserClaims) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings(u.Audience), nil
}

// GetExpirationTime implements jwt.Claims.
func (u *UserClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(u.ExpirationTime), nil
}

// GetIssuedAt implements jwt.Claims.
func (u *UserClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(u.IssuedAt), nil
}

// GetIssuer implements jwt.Claims.
func (u *UserClaims) GetIssuer() (string, error) {
	return u.Issuer, nil
}

// GetNotBefore implements jwt.Claims.
func (u *UserClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(u.NotBefore), nil
}

// GetSubject implements jwt.Claims.
func (u *UserClaims) GetSubject() (string, error) {
	return "", nil
}
