package main

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/v6/plugin"
)

type LichessPlugin struct {
	plugin.MattermostPlugin

	botID string
}

func (p *LichessPlugin) handleLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "login data?")
}

func (p *LichessPlugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("MMUSERID")

	if err != nil {
		http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
	}

	switch r.URL.Path {
	case "/login":
		p.handleLogin(w, r)
	default:
		http.NotFound(w, r)
	}

}

func main() {
	plugin.ClientMain(&LichessPlugin{})
}
