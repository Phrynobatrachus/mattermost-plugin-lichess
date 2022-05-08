package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-server/v6/plugin"
	"golang.org/x/oauth2"
)

type LichessPlugin struct {
	plugin.MattermostPlugin
}

const (
	authServerURL = "https://lichess.org"
)

var (
	config = oauth2.Config{
		ClientID:     "abcdef",
		ClientSecret: "123456789",
		Scopes:       []string{"preference:read"},
		RedirectURL:  "http://localhost:8065/plugins/com.mattermost.lichess-plugin/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  authServerURL + "/oauth",
			TokenURL: authServerURL + "/api/token",
		},
	}
	globalToken    *oauth2.Token
	globalVerifier string
	globalState    string
)

func genVerifier() (string, error) {
	seed, err := rand.Prime(rand.Reader, 256)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(seed.Bytes()), nil
}

func genCodeChallengeS256(s string) string {
	s256 := sha256.Sum256([]byte(s))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(s256[:])
}

func (p *LichessPlugin) handleLogin(w http.ResponseWriter, r *http.Request) {
	verifier, err := genVerifier()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	globalVerifier = verifier

	state, err := genVerifier()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	globalState = state

	u := config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge", genCodeChallengeS256(verifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	http.Redirect(w, r, u, http.StatusFound)
}

func (p *LichessPlugin) handleCallback(w http.ResponseWriter, r *http.Request) {
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

	token, err := config.Exchange(context.Background(), c, oauth2.SetAuthURLParam("code_verifier", globalVerifier))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	globalToken = token

	e := json.NewEncoder(w)
	e.SetIndent("", " ")
	e.Encode(token)
}

func (p *LichessPlugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("MMUSERID")

	if err != nil {
		http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
	}

	switch r.URL.Path {
	case "/login":
		p.handleLogin(w, r)
	case "/callback":
		p.handleCallback(w, r)
	default:
		http.NotFound(w, r)
	}

}

func main() {
	plugin.ClientMain(&LichessPlugin{})
}
