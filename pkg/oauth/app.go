package oauth

type App interface {
	GetAuthURL() (string, error)
	ExchangeCode(code string) (string, error)
	GetUserProfile(accessToken string) (*UserProfile, error)
}

type GetAuthURLOption struct {
	RedirectURL *string
}

type ExchangeCodeOption struct {
	RedirectURL *string
}

type UserProfile struct {
	Name      string
	Email     string
	AvatarURL string
}
