package authenticator

import (
	"context"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Config defines required configuration values for Auth0.
//
// * Values are read from the environment.
// They cannot be overridden or set from code.
type Config struct {
	Domain       string
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type Authenticator struct {
	*oidc.Provider
	oauth2.Config
	LogoutURL string
}

func New() (*Authenticator, error) {
	cfg := Config{
		Domain:       os.Getenv("AUTH0_DOMAIN"),
		ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
		ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("AUTH0_REDIRECT_URI"),
	}

	if cfg.Domain == "" {
		return nil, ErrEmptyDomain
	}

	if cfg.ClientID == "" {
		return nil, ErrEmptyClientID
	}

	if cfg.ClientSecret == "" {
		return nil, ErrEmptyClientSecret
	}

	if cfg.RedirectURI == "" {
		return nil, ErrEmptyRedirectURI
	}

	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+cfg.Domain+"/",
	)
	if err != nil {
		return nil, err
	}

	return &Authenticator{
		Provider: provider,
		Config: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURI,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
			Endpoint:     provider.Endpoint(),
		},
		LogoutURL: "https://" + cfg.Domain + "/v2/logout",
	}, nil
}

func (a *Authenticator) VerifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, ErrNoIDToken
	}

	oidcConfig := &oidc.Config{
		ClientID: a.ClientID,
	}

	return a.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}
