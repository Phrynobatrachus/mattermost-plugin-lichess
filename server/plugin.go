package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sync"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type Plugin struct {
	plugin.MattermostPlugin
	pluginAPI *pluginapi.Client
	//lichessAPI *lichessapi.Client

	configurationLock sync.RWMutex
	configuration     *Configuration
}

var (
	globalToken    *oauth2.Token
	globalVerifier string
	globalState    string
)

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

	redirectURL, err := url.Parse(fmt.Sprintf("%s/plugins/com.mattermost.lichess-plugin/callback", *p.pluginAPI.Configuration.GetConfig().ServiceSettings.SiteURL))
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

func (p *Plugin) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	verifier, err := generateSecret(*base64.RawURLEncoding, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	globalVerifier = verifier

	state, err := generateSecret(*base64.RawURLEncoding, 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	globalState = state

	config, err := p.getOAuthConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	u := config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge", genCodeChallengeS256(verifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	http.Redirect(w, r, u, http.StatusFound)
}

func (p *Plugin) handleCallback(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	s := qs.Get("state")
	c := qs.Get("code")

	if s != globalState {
		http.Error(w, "State invalid", http.StatusBadRequest)
		return
	}

	if c == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	config, err := p.getOAuthConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := config.Exchange(context.Background(), c, oauth2.SetAuthURLParam("code_verifier", globalVerifier))
	if err != nil {
		fmt.Fprint(w, globalVerifier)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	globalToken = token

	e := json.NewEncoder(w)
	e.SetIndent("", " ")
	e.Encode(token)
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("MMUSERID")
	if err != nil {
		http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.URL.Path {
	case "/login":
		p.handleLogin(w, r)
	case "/callback":
		p.handleCallback(w, r)
	default:
		http.NotFound(w, r)
		return
	}

}

func (p *Plugin) OnActivate() error {
	pluginAPIClient := pluginapi.NewClient(p.API, p.Driver)
	p.pluginAPI = pluginAPIClient

	siteURL := p.pluginAPI.Configuration.GetConfig().ServiceSettings.SiteURL
	if siteURL == nil || *siteURL == "" {
		return errors.New("siteURL must be set")
	}

	err := p.setDefaultConfiguration()
	if err != nil {
		return errors.Wrap(err, "failed to set default configuration")
	}

	return nil
}

func (p *Plugin) setDefaultConfiguration() error {
	config := p.getConfiguration()

	changed, err := config.setDefaults()
	if err != nil {
		return err
	}

	if changed {
		configMap, err := config.ToMap()
		if err != nil {
			return err
		}

		err = p.pluginAPI.Configuration.SavePluginConfig(configMap)
		if err != nil {
			return err
		}
	}
	return nil
}
