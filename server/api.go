package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-plugin-api/experimental/bot/logger"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type Context struct {
	Ctx    context.Context
	UserID string
	Log    logger.Logger
}

type UserContext struct {
	Context,
	LCInfo *LichessUserInfo
}

type HTTPHandlerFuncWithContext func(c *Context, w http.ResponseWriter, r *http.Request)

type HTTPHandlerFuncWithUserContext func(c *UserContext, w http.ResponseWriter, r *http.Request)

type ResponseType string

type APIErrorResponse struct {
	ID         string `json:"id"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (e *APIErrorResponse) Error() string {
	return e.Message
}

const (
	ResponseTypeJson  ResponseType = "JSON_RESPONSE"
	ResponseTypePlain ResponseType = "TEXT_RESPONSE"
)

func (p *Plugin) writeJSON(w http.ResponseWriter, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		p.API.LogWarn("failed to marshal JSON response", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		p.API.LogWarn("failed to write JSON response", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (p *Plugin) writeAPIError(w http.ResponseWriter, apiErr *APIErrorResponse) {
	b, err := json.Marshal(apiErr)
	if err != nil {
		p.API.LogWarn("failed to marshal API error", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(apiErr.StatusCode)

	_, err = w.Write(b)
	if err != nil {
		p.API.LogWarn("failed to write JSON response", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (p *Plugin) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if x := recover(); x != nil {
				p.API.LogError("recovered from a panic",
					"url", r.URL.String(), "error", x, "stack", string(debug.Stack()))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (p *Plugin) initializeAPI() {
	p.router = mux.NewRouter()
	p.router.Use(p.withRecovery)

	oauthRouter := p.router.PathPrefix("/oauth").Subrouter()

	oauthRouter.HandleFunc("/connect", p.checkAuth(p.attachContext(p.handleLogin), ResponseTypePlain)).Methods(http.MethodGet)
	oauthRouter.HandleFunc("/complete", p.checkAuth(p.attachContext(p.handleCallback), ResponseTypePlain)).Methods(http.MethodGet)
}

func (p *Plugin) checkAuth(handler http.HandlerFunc, responseType ResponseType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := r.Cookie("MMUSERID")
		if err != nil || userID.String() == "" {
			switch responseType {
			case ResponseTypeJson:
				p.writeAPIError(w, &APIErrorResponse{ID: "", Message: "not authorized", StatusCode: http.StatusUnauthorized})
			case ResponseTypePlain:
				http.Error(w, "not authorized", http.StatusUnauthorized)
			default:
				p.API.LogError("unknown response type")
			}
			return
		}

		handler(w, r)
	}
}

func (p *Plugin) createContext(w http.ResponseWriter, r *http.Request) (*Context, context.CancelFunc) {
	userID, err := r.Cookie("MMUSERID")
	if err != nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return nil, nil
	}

	logger := logger.New(p.API).With(logger.LogContext{
		"userid": userID.String(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	context := &Context{
		Ctx:    ctx,
		UserID: userID.String(),
		Log:    logger,
	}

	return context, cancel
}

func (p *Plugin) attachContext(handler HTTPHandlerFuncWithContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		context, cancel := p.createContext(w, r)
		defer cancel()

		handler(context, w, r)
	}
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	p.router.ServeHTTP(w, r)
}

func (p *Plugin) handleLogin(c *Context, w http.ResponseWriter, r *http.Request) {
	state, err := generateSecret(*base64.RawURLEncoding, 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	verifier, err := generateSecret(*base64.RawURLEncoding, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	oauthState := OAuthState{
		UserID:   c.UserID,
		State:    state,
		Verifier: verifier,
	}

	stateBytes, err := json.Marshal(oauthState)
	if err != nil {
		http.Error(w, "json marshal failed", http.StatusInternalServerError)
		return
	}

	appErr := p.API.KVSetWithExpiry(lichessOauthKey+oauthState.State, stateBytes, 10*60)
	if appErr != nil {
		http.Error(w, "error setting oauth request state", http.StatusInternalServerError)
		return
	}

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

	ch := p.oauthBroker.SubscribeOAuthComplete(c.UserID)

	go func() {
		_, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		// var errorMsg string
		// select {
		// case err := <-ch:
		// 	if err != nil {
		// 		errorMsg = err.Error()
		// 	}
		// case <-ctx.Done():
		// 	errorMsg = "Timed out waiting for OAuth"
		// }

		p.oauthBroker.UnsubscribeOAuthComplete(c.UserID, ch)
	}()

	http.Redirect(w, r, u, http.StatusFound)
}

func (p *Plugin) handleCallback(c *Context, w http.ResponseWriter, r *http.Request) {
	var rErr error
	defer func() {
		p.oauthBroker.publishOAuthComplete(c.UserID, rErr, false)
	}()

	qs := r.URL.Query()
	s := qs.Get("state")
	code := qs.Get("code")

	storedState, appErr := p.API.KVGet(lichessOauthKey + s)
	if appErr != nil {
		c.Log.Warnf("Failed to get state token", "error", appErr.Error())

		rErr = errors.Wrap(appErr, "missing stored state")
		http.Error(w, rErr.Error(), http.StatusBadRequest)
		return
	}

	var oauthState OAuthState
	if err := json.Unmarshal(storedState, &oauthState); err != nil {
		rErr = errors.Wrap(err, "json unmarshal failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	appErr = p.API.KVDelete(lichessOauthKey + s)
	if appErr != nil {
		c.Log.WithError(appErr).Warnf("Failed to delete state token")

		rErr = errors.Wrap(appErr, "error deleting stored state")
		http.Error(w, rErr.Error(), http.StatusBadRequest)
		return
	}

	if s != oauthState.State {
		http.Error(w, "State invalid", http.StatusBadRequest)
		return
	}

	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	if c.UserID != oauthState.UserID {
		rErr = errors.New("not authorized, incorrect user")
		http.Error(w, rErr.Error(), http.StatusUnauthorized)
		return
	}

	config, err := p.getOAuthConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tok, err := config.Exchange(c.Ctx, code, oauth2.SetAuthURLParam("code_verifier", oauthState.Verifier))
	if err != nil {
		c.Log.WithError(err).Warnf("Failed to exchange oauth code into token")

		rErr = errors.Wrap(err, "Failed to exchange oauth code into token")
		http.Error(w, rErr.Error(), http.StatusInternalServerError)
		return
	}

	ts := oauth2.StaticTokenSource(tok)
	tc := oauth2.NewClient(c.Ctx, ts)

	res, err := tc.Get("https://lichess.org/api/account/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	var prefs UserPrefs
	err = json.Unmarshal(body, prefs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo := &LichessUserInfo{
		UserID: oauthState.UserID,
		Token:  tok,
	}

	if err = p.storeLichessUserInfo(userInfo); err != nil {
		c.Log.WithError(err).Warnf("failed to store Lichess user info")
		rErr = errors.Wrap(err, "unable to connect user to Lichess")
		http.Error(w, rErr.Error(), http.StatusInternalServerError)
		return
	}

	// fetchedInfo, err := p.getLichessUserInfo(oauthState.UserID)
	// if err != nil {
	// 	c.Log.WithError(err).Warnf("failed to get Lichess user info")
	// 	rErr = errors.Wrap(err, "unable to connect user to Lichess")
	// 	http.Error(w, rErr.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// p.writeJSON(w, fetchedInfo)

	html := `
	<!DOCTYPE html>
	<html>
	<head>
	<script>
	//window.close();
	</script>
	<body>
	<p>Completed connecting to Lichess. Please close this window.</p>
	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write([]byte(html))
	if err != nil {
		c.Log.WithError(err).Warnf("failed to write html response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
