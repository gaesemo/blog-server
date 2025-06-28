package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

var _ App = (*github)(nil)

func NewGitHub(httpClient *http.Client, randStrFunc func() string) App {
	if randStrFunc == nil {
		randStrFunc = func() string {
			return uuid.NewString()
		}
	}
	return &github{
		config: &oauth2.Config{
			ClientID:     viper.GetString("OAUTH_GITHUB_CLIENT_ID"),
			ClientSecret: viper.GetString("OAUTH_GITHUB_CLIENT_SECRET"),
			Endpoint:     endpoints.GitHub,
			RedirectURL:  viper.GetString("OAUTH_GITHUB_REDIRECT_URL"),
			Scopes:       []string{"user"}, // https://docs.github.com/ko/apps/oauth-apps/building-oauth-apps/scopes-for-oauth-apps#available-scopes
		},
		randStrFunc: randStrFunc,
		httpClient:  httpClient,
	}
}

type github struct {
	config      *oauth2.Config
	httpClient  *http.Client
	randStrFunc func() string
}

// https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#1-request-a-users-github-identity
func (g *github) GetAuthURL() (string, error) {
	params := url.Values{}
	params.Add("client_id", g.config.ClientID)
	params.Add("scope", strings.Join(g.config.Scopes, " "))
	params.Add("state", g.randStrFunc())
	authUrl := g.config.Endpoint.AuthURL + "?" + params.Encode()
	return authUrl, nil
}

func (g *github) ExchangeCode(code string) (string, error) {

	reqBody := map[string]string{
		"client_id":     g.config.ClientID,
		"client_secret": g.config.ClientSecret,
		"code":          code,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, _ := http.NewRequest("POST", g.config.Endpoint.TokenURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("exchanging token status: %d body: %s", resp.StatusCode, string(body))
	}
	slog.Info("resp body", slog.Any("response", string(body)))

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %v", err)
	}

	return tokenResp.AccessToken, nil
}

func (g *github) GetUserProfile(accessToken string) (*UserProfile, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting github user: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("requesting user status: %d body: %s", resp.StatusCode, string(body))
	}

	var user struct {
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("unmarshalling json: %v", err)
	}

	return &UserProfile{
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
	}, nil
}
