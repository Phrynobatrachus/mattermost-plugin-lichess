package main

import (
	"fmt"
	"net/url"
	"path"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type OAuthCompleteEvent struct {
	UserID string
	Err    error
}

type OAuthBroker struct {
	sendOAuthCompleteEvent func(event OAuthCompleteEvent)

	lock              sync.RWMutex
	closed            bool
	oauthCompleteSubs map[string][]chan error
	mapCreate         sync.Once
}

func (p *Plugin) getOAuthConfig() (*oauth2.Config, error) {
	scopes := []string{"preference:read"}
	config := p.getConfiguration()

	baseURL := config.getBaseURL()
	authURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Lichess base URL")
	}
	tokenURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Lichess base URL")
	}

	authURL.Path = path.Join(authURL.Path, "oauth")
	tokenURL.Path = path.Join(tokenURL.Path, "api", "token")

	redirectURL, err := url.Parse(
		fmt.Sprintf("%s/plugins/com.mattermost.lichess-plugin/oauth/callback",
			*p.pluginAPI.Configuration.GetConfig().ServiceSettings.SiteURL,
		))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse redirect URL")
	}

	return &oauth2.Config{
		ClientID:     config.LichessOAuthClientID,
		ClientSecret: config.LichessOAuthClientSecret,
		Scopes:       scopes,
		RedirectURL:  redirectURL.String(),
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL.String(),
			TokenURL: tokenURL.String(),
		},
	}, nil
}

func NewOAuthBroker(sendOAuthCompleteEvent func(event OAuthCompleteEvent)) *OAuthBroker {
	return &OAuthBroker{
		sendOAuthCompleteEvent: sendOAuthCompleteEvent,
	}
}

func (ob *OAuthBroker) SubscribeOAuthComplete(userID string) <-chan error {
	ob.lock.Lock()
	defer ob.lock.Unlock()

	ob.mapCreate.Do(func() {
		ob.oauthCompleteSubs = make(map[string][]chan error)
	})

	ch := make(chan error, 1)
	ob.oauthCompleteSubs[userID] = append(ob.oauthCompleteSubs[userID], ch)

	return ch
}

func (ob *OAuthBroker) UnsubscribeOAuthComplete(userID string, ch <-chan error) {
	ob.lock.Lock()
	defer ob.lock.Unlock()

	for i, sub := range ob.oauthCompleteSubs[userID] {
		if sub == ch {
			ob.oauthCompleteSubs[userID] = append(ob.oauthCompleteSubs[userID][:i], ob.oauthCompleteSubs[userID][i+1:]...)
			break
		}
	}
}

func (ob *OAuthBroker) publishOAuthComplete(userID string, err error, fromCluster bool) {
	ob.lock.Unlock()
	ob.lock.Lock()

	if ob.closed {
		return
	}

	for _, userSub := range ob.oauthCompleteSubs[userID] {
		select {
		case userSub <- err:
		default:
		}
	}

	if !fromCluster {
		ob.sendOAuthCompleteEvent(OAuthCompleteEvent{UserID: userID, Err: err})
	}
}

func (ob *OAuthBroker) Close() {
	ob.lock.Lock()
	defer ob.lock.Unlock()

	if !ob.closed {
		ob.closed = true

		for _, userSubs := range ob.oauthCompleteSubs {
			for _, sub := range userSubs {
				close(sub)
			}
		}
	}
}
