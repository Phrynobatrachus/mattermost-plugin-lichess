package main

import (
	"sync"

	"github.com/gorilla/mux"
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
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

	router *mux.Router

	oauthBroker *OAuthBroker
}

type LichessUserInfo struct {
	UserID string
	Token  *oauth2.Token
}

type OAuthState struct {
	UserID   string
	State    string
	Verifier string
}

const (
	lichessOauthKey = "lichessoauthkey_"
)

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

	p.initializeAPI()

	p.oauthBroker = NewOAuthBroker(p.sendOAuthCompleteEvent)

	return nil
}

func (p *Plugin) OnDeactivate() error {
	p.oauthBroker.Close()

	return nil
}

func (p *Plugin) OnPluginClusterEvent(c *plugin.Context, ev model.PluginClusterEvent) {
	p.HandleClusterEvent(ev)
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
