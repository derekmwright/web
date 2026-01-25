package auth0

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const StateKey = "state"

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), err
}

func HandleLogin(deps *deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := generateRandomState()
		if err != nil {
			deps.log.Error("unable to generate random state", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		deps.log.Info("generated state", "state", state)

		deps.sessions.Put(r.Context(), StateKey, state)

		http.Redirect(w, r, deps.auth.AuthCodeURL(state), http.StatusFound)
	}
}

func HandleLogout(deps *deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deps.sessions.Put(r.Context(), "user", nil)
		deps.sessions.Put(r.Context(), StateKey, nil)

		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}

		returnTo := scheme + "://" + r.Host

		logout := deps.logoutBase.ResolveReference(&url.URL{})
		q := logout.Query()
		q.Add("returnTo", returnTo)
		q.Add("client_id", deps.auth.ClientID)
		logout.RawQuery = q.Encode()

		http.Redirect(w, r, logout.String(), http.StatusFound)
	}
}

type Authenticator interface {
	Exchange(context.Context, string) (*oauth2.Token, error)
	VerifyIDToken(context.Context, *oauth2.Token) (*oidc.IDToken, error)
}

func HandleCallback(deps *deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user SessionUser

		_, ok := deps.sessions.Get(r.Context(), StateKey).(string)
		if !ok {
			deps.log.Error("no state found in session")
			http.Error(w, "no state found in session", http.StatusInternalServerError)
			return
		}

		deps.sessions.Put(r.Context(), StateKey, nil)

		token, err := deps.auth.Exchange(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			deps.log.Error("unable to exchange auth code for token", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		idToken, err := deps.auth.VerifyIDToken(r.Context(), token)
		if err != nil {
			deps.log.Error("unable to verify ID token", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var rawClaims map[string]json.RawMessage
		if err = idToken.Claims(&rawClaims); err != nil {
			deps.log.Error("unable to decode ID token claims", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if sub, ok := rawClaims["sub"]; ok {
			json.Unmarshal(sub, &user.Sub)
		}
		if name, ok := rawClaims["name"]; ok {
			json.Unmarshal(name, &user.Name)
		}
		if email, ok := rawClaims["email"]; ok {
			json.Unmarshal(email, &user.Email)
		}
		if emailVerified, ok := rawClaims["email_verified"]; ok {
			json.Unmarshal(emailVerified, &user.EmailVerified)
		}
		if picture, ok := rawClaims["picture"]; ok {
			json.Unmarshal(picture, &user.Picture)
		}

		customMap := make(map[string]json.RawMessage)
		for k, v := range rawClaims {
			if k != "sub" && k != "name" && k != "email" && k != "picture" {
				customMap[k] = v
			}
		}

		if len(customMap) > 0 {
			user.Custom, _ = json.Marshal(customMap)
		}

		deps.sessions.Put(r.Context(), "user", user)
		deps.sessions.Put(r.Context(), "access_token", token.AccessToken)

		for _, hook := range deps.postLoginHooks {
			if err = hook(r.Context(), &user, r); err != nil {
				deps.log.Error("error running post login hook", "error", err)
			}
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}
