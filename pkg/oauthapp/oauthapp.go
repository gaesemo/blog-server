package oauthapp

type OAuthApp interface {
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
