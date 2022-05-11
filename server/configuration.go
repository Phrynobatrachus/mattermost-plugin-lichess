package main

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type Configuration struct {
	LichessOAuthClientID     string `json:"lichessoauthclientid"`
	LichessOAuthClientSecret string `json:"lichessoauthclientsecret"`
	EncryptionKey            string `json:"encryptionkey"`
}

func (c *Configuration) setDefaults() (bool, error) {
	changed := false

	if c.LichessOAuthClientID == "" {
		secret, err := generateSecret(*base64.RawStdEncoding, 32)
		if err != nil {
			return false, err
		}
		c.LichessOAuthClientID = secret
		changed = true
	}

	if c.LichessOAuthClientSecret == "" {
		secret, err := generateSecret(*base64.RawStdEncoding, 32)
		if err != nil {
			return false, err
		}
		c.LichessOAuthClientSecret = secret
		changed = true
	}

	if c.EncryptionKey == "" {
		secret, err := generateSecret(*base64.RawStdEncoding, 32)
		if err != nil {
			return false, err
		}

		c.EncryptionKey = secret
		changed = true
	}

	return changed, nil
}

func (c *Configuration) getBaseURL() string {
	return "https://lichess.org/"
}

func (c *Configuration) sanitize() {
	c.LichessOAuthClientID = strings.TrimSpace(c.LichessOAuthClientID)
	c.LichessOAuthClientSecret = strings.TrimSpace(c.LichessOAuthClientSecret)
}

func (c *Configuration) IsOAuthConfigured() bool {
	return (c.LichessOAuthClientID != "" && c.LichessOAuthClientSecret != "")
}

func (c *Configuration) ToMap() (map[string]interface{}, error) {
	var out map[string]interface{}
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *Configuration) Clone() *Configuration {
	var clone = *c
	return &clone
}

func (c *Configuration) IsValid() error {
	if c.LichessOAuthClientID == "" {
		return errors.New("must have an oauth client id")
	}
	if c.LichessOAuthClientSecret == "" {
		return errors.New("must have an oauth client secret")
	}
	if c.EncryptionKey == "" {
		return errors.New("must have an encryption key")
	}
	return nil
}

func (p *Plugin) getConfiguration() *Configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &Configuration{}
	}
	return p.configuration
}

func (p *Plugin) setConfiguration(c *Configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if c != nil && p.configuration == c {
		if reflect.ValueOf(*c).NumField() == 0 {
			return
		}

		panic("setConfiguration called with the existing configuration")
	}

	p.configuration = c
}

func (p *Plugin) OnConfigurationChange() error {
	var configuration = new(Configuration)

	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	configuration.sanitize()

	p.setConfiguration(configuration)

	return nil
}
