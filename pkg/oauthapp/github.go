package oauthapp

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

var _ OAuthApp = (*github)(nil)

func NewGitHub(httpClient *http.Client, randStr func() string) OAuthApp {
	if randStr == nil {
		randStr = func() string {
			return uuid.NewString()
		}
	}
	return &github{
		config: &oauth2.Config{
			ClientID:     viper.GetString("OAUTH_GITHUB_CLIENT_ID"),
			ClientSecret: viper.GetString("OAUTH_GITHUB_CLIENT_SECRET"),
			Endpoint:     endpoints.GitHub,
			RedirectURL:  viper.GetString("OAUTH_GITHUB_CALLBACK_URL"),
			Scopes:       []string{"user"}, // https://docs.github.com/ko/apps/oauth-apps/building-oauth-apps/scopes-for-oauth-apps#available-scopes
		},
		randStr:    randStr,
		httpClient: httpClient,
	}
}

type github struct {
	config     *oauth2.Config
	httpClient *http.Client
	randStr    func() string
}

// query params
// "client_id"
// "client_secret"
// "code"
// "state"
// "scope"
// "redirect_uri"

// https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#1-request-a-users-github-identity
func (g *github) GetAuthURL(option *GetAuthURLOption) (string, error) {
	params := url.Values{}
	params.Add("client_id", g.config.ClientID)
	if option != nil && option.RedirectURL != nil {
		r := *option.RedirectURL
		u, err := url.Parse(r)
		if err != nil {
			return "", fmt.Errorf("invalid redirect url: %v", err)
		}
		if _, err := g.isValidRedirectURL(*u); err != nil {
			return "", fmt.Errorf("invalid redirect url: %v", err)
		}
		params.Add("redirect_uri", r)
	}
	params.Add("scope", strings.Join(g.config.Scopes, " "))
	params.Add("state", g.randStr())
	authUrl := g.config.Endpoint.AuthURL + "?" + params.Encode()
	return authUrl, nil
}

func (g *github) ExchangeCode(code string, option *ExchangeCodeOption) (string, error) {
	return "", nil
}

func (g *github) GetUserProfile(accessToken string) (UserProfile, error) {
	return UserProfile{}, nil
}

func (g *github) isValidRedirectURL(u url.URL) (bool, error) {
	d, _ := url.Parse(g.config.RedirectURL) // default redirect url configured in GitHub oauth app setting
	if (d.Hostname() != u.Hostname()) || (d.Port() != u.Port()) {
		return false, fmt.Errorf("redirect url %q's host and port must match configured oauth app settings", u.String())
	}
	return true, nil
}
