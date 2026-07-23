package sso

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCAdapter implements SSOAdapter for an OpenID Connect provider (Keycloak).
type OIDCAdapter struct {
	provider       *oidc.Provider
	verifier       *oidc.IDTokenVerifier
	oauth2Cfg      *oauth2.Config
	clientID       string
	clientSecret   string
	endSessionURL  string
	introspectURL  string
	revokeURL      string
	postLogoutURI  string
}

// NewOIDCAdapter creates an OIDCAdapter by discovering the provider's config.
func NewOIDCAdapter(ctx context.Context, issuer, clientID, clientSecret, redirectURI, postLogoutURI string) (*OIDCAdapter, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to discover provider %s: %w", issuer, err)
	}

	oauth2Cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURI,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	base := strings.TrimSuffix(issuer, "/") + "/protocol/openid-connect"

	return &OIDCAdapter{
		provider:      provider,
		verifier:      provider.Verifier(&oidc.Config{ClientID: clientID}),
		oauth2Cfg:     oauth2Cfg,
		clientID:      clientID,
		clientSecret:  clientSecret,
		endSessionURL: base + "/logout",
		introspectURL: base + "/token/introspect",
		revokeURL:     base + "/revoke",
		postLogoutURI: postLogoutURI,
	}, nil
}

// AuthorizationURL returns the Keycloak authorization URL with the CSRF state.
func (a *OIDCAdapter) AuthorizationURL(state string) string {
	return a.oauth2Cfg.AuthCodeURL(state,
		oauth2.SetAuthURLParam("response_type", "code"),
	)
}

// ExchangeCode exchanges an authorization code for tokens via backchannel.
func (a *OIDCAdapter) ExchangeCode(ctx context.Context, code string) (string, string, string, error) {
	oauth2Token, err := a.oauth2Cfg.Exchange(ctx, code)
	if err != nil {
		return "", "", "", fmt.Errorf("oidc: token exchange failed: %w", err)
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return "", "", "", fmt.Errorf("oidc: no id_token in token response")
	}

	return oauth2Token.AccessToken, rawIDToken, oauth2Token.RefreshToken, nil
}

// ValidateToken verifies an id_token and returns the user info.
func (a *OIDCAdapter) ValidateToken(ctx context.Context, token string) (*SSOUser, error) {
	idToken, err := a.verifier.Verify(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("oidc: invalid id_token: %w", err)
	}

	var claims struct {
		Sub               string `json:"sub"`
		Email             string `json:"email"`
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("oidc: failed to parse id_token: %w", err)
	}

	user := &SSOUser{
		ExternalID: claims.Sub,
		Email:      claims.Email,
		Name:       claims.Name,
	}
	if claims.PreferredUsername != "" {
		user.ExternalID = claims.PreferredUsername
	}
	return user, nil
}

// GetUserFromToken extracts user info from an access_token JWT (without
// signature verification — the token came from a backchannel code exchange
// authenticated with the client_secret).
func (a *OIDCAdapter) GetUserFromToken(ctx context.Context, accessToken string) (*SSOUser, error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("oidc: invalid access_token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to decode access_token: %w", err)
	}

	var claims struct {
		Sub               string `json:"sub"`
		Email             string `json:"email"`
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
		ResourceAccess    map[string]struct {
			Roles []string `json:"roles"`
		} `json:"resource_access"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("oidc: failed to parse access_token claims: %w", err)
	}

	var roles []string
	if ra, ok := claims.ResourceAccess[a.clientID]; ok {
		for _, r := range ra.Roles {
			if r != "access" {
				roles = append(roles, r)
			}
		}
	}

	user := &SSOUser{
		ExternalID: claims.Sub,
		Email:      claims.Email,
		Name:       claims.Name,
		Roles:      roles,
	}
	if claims.PreferredUsername != "" {
		user.ExternalID = claims.PreferredUsername
	}
	return user, nil
}

// HasAccess checks whether the access_token's resource_access contains the
// "access" role for this adapter's client — required for app entry.
func (a *OIDCAdapter) HasAccess(ctx context.Context, accessToken string) bool {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return false
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	var claims struct {
		ResourceAccess map[string]struct {
			Roles []string `json:"roles"`
		} `json:"resource_access"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return false
	}

	ra, ok := claims.ResourceAccess[a.clientID]
	if !ok {
		return false
	}
	for _, r := range ra.Roles {
		if r == "access" {
			return true
		}
	}
	return false
}

// ValidateAccessToken calls the provider's Introspect endpoint to verify the
// access token is still active (not expired, not revoked).
// Returns nil if active, error otherwise.
func (a *OIDCAdapter) ValidateAccessToken(ctx context.Context, accessToken string) error {
	data := url.Values{
		"client_id":     {a.clientID},
		"client_secret": {a.clientSecret},
		"token":         {accessToken},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.introspectURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("oidc: failed to create introspect request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("oidc: introspect request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("oidc: failed to read introspect response: %w", err)
	}

	var result struct {
		Active bool `json:"active"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("oidc: failed to decode introspect response: %w", err)
	}

	if !result.Active {
		return fmt.Errorf("oidc: token is not active")
	}
	return nil
}

// RevokeToken sends a token (typically refresh_token) to the provider's
// revocation endpoint. After revocation the token can no longer be used
// to obtain new access or refresh tokens. Returns nil on success.
func (a *OIDCAdapter) RevokeToken(ctx context.Context, token string) error {
	data := url.Values{
		"client_id":       {a.clientID},
		"client_secret":   {a.clientSecret},
		"token":           {token},
		"token_type_hint": {"refresh_token"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("oidc: failed to create revoke request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("oidc: revoke request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("oidc: revoke returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// GetEndSessionURL returns the Keycloak end-session URL with the
// post_logout_redirect_uri and id_token_hint parameters so the user
// returns to the app after SSO logout completes.
func (a *OIDCAdapter) GetEndSessionURL(_ context.Context, idTokenHint string) (string, error) {
	u, err := url.Parse(a.endSessionURL)
	if err != nil {
		return "", fmt.Errorf("oidc: invalid end-session URL: %w", err)
	}
	q := u.Query()
	q.Set("post_logout_redirect_uri", a.postLogoutURI)
	q.Set("client_id", a.clientID)
	if idTokenHint != "" {
		q.Set("id_token_hint", idTokenHint)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
