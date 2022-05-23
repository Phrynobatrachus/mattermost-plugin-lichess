package main

import (
	"encoding/json"
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
	UserID          string
	Token           *oauth2.Token
	LichessUsername string
}

type OAuthState struct {
	UserID   string
	State    string
	Verifier string
}

const (
	lichessOauthKey = "lichessoauthkey_"
	lichessTokenKey = "_lichesstoken"
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

func (p *Plugin) storeLichessUserInfo(info *LichessUserInfo) error {
	config := p.getConfiguration()

	encryptedToken, err := encrypt([]byte(config.EncryptionKey), info.Token.AccessToken)
	if err != nil {
		return errors.Wrap(err, "error occured while encrypting access token")
	}

	info.Token.AccessToken = encryptedToken

	jsonInfo, err := json.Marshal(info)
	if err != nil {
		return errors.Wrap(err, "error while converting user info to json")
	}

	if err := p.API.KVSet(info.UserID+lichessTokenKey, jsonInfo); err != nil {
		return errors.Wrap(err, "failed to store user info in kv store")
	}

	return nil
}

func (p *Plugin) getLichessUserInfo(userID string) (*LichessUserInfo, error) {
	config := p.getConfiguration()

	info, appErr := p.API.KVGet(userID + lichessTokenKey)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "failed  to get Lichess user info from kv store")
	}

	var userInfo LichessUserInfo
	if err := json.Unmarshal(info, &userInfo); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal user info")
	}

	unencryptedToken, err := decrypt([]byte(config.EncryptionKey), userInfo.Token.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt lichess token")
	}

	userInfo.Token.AccessToken = unencryptedToken

	return &userInfo, nil
}
