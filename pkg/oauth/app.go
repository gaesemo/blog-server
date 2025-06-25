package oauth

import "strings"

type App interface {
	GetAuthURL(option *GetAuthURLOption) (string, error)
	ExchangeCode(code string, option *ExchangeCodeOption) (string, error)
	GetUserProfile(accessToken string) (UserProfile, error)
}

type GetAuthURLOption struct {
	RedirectURL *string
}

type ExchangeCodeOption struct {
	RedirectURL *string
}

type UserProfile struct {
	Email     string
	AvatarURL string
}

func (p UserProfile) TempName() string {
	return strings.Split(p.Email, "@")[0]
}
