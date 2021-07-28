package slack

import (
	"net/http"
	"net/url"
)

// AuthTestResponse is response to auth.test - see https://api.slack.com/methods/auth.test
type AuthTestResponse struct {
	slackResponse
	URL    string `json:"url"`
	Team   string `json:"team"`
	User   string `json:"user"`
	TeamID string `json:"team_id"`
	UserID string `json:"user_id"`
}

// OAuthAccessResponse - See https://api.slack.com/methods/oauth.access
type OAuthAccessResponse struct {
	slackResponse
	AccessToken     string                       `json:"access_token"`
	Scope           string                       `json:"scope"`
	TeamName        string                       `json:"team_name"`
	TeamID          string                       `json:"team_id"`
	IncomingWebhook OAuthResponseIncomingWebhook `json:"incoming_webhook"`
	Bot             struct {
		BotUserID      string `json:"bot_user_id"`
		BotAccessToken string `json:"bot_access_token"`
	} `json:"bot"`
}

// OAuthResponseIncomingWebhook ...
type OAuthResponseIncomingWebhook struct {
	URL              string `json:"url"`
	Channel          string `json:"channel"`
	ChannelID        string `json:"channel_id,omitempty"`
	ConfigurationURL string `json:"configuration_url"`
}

// AuthTest tests if the authentication is in place - see https://api.slack.com/methods/auth.test
func (s *Slack) AuthTest() (*AuthTestResponse, error) {
	r := &AuthTestResponse{}
	err := s.do("auth.test", url.Values{}, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// OAuthAccess returns the token for OAuth
func OAuthAccess(clientID, clientSecret, code, redirectURI string) (*OAuthAccessResponse, error) {
	params := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
	}
	if redirectURI != "" {
		params.Set("redirect_uri", redirectURI)
	}
	s := &Slack{
		url: DefaultURL,
		c:   http.DefaultClient,
	}
	r := &OAuthAccessResponse{}
	err := s.do("oauth.access", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
