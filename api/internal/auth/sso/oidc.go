package sso

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
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

type OIDCAdapter struct {
	provider      *oidc.Provider
	verifier      *oidc.IDTokenVerifier
	oauth2Cfg     *oauth2.Config
	clientID      string
	clientSecret  string
	endSessionURL string
	introspectURL string
	revokeURL     string
	postLogoutURI string
}

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

func GeneratePKCE() (verifier, challenge string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("oidc: pkce: %w", err)
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])
	return verifier, challenge, nil
}

func (a *OIDCAdapter) AuthorizationURL(state, acr, codeVerifier string) string {
	challenge := ""
	if codeVerifier != "" {
		h := sha256.Sum256([]byte(codeVerifier))
		challenge = base64.RawURLEncoding.EncodeToString(h[:])
	}

	var params []oauth2.AuthCodeOption
	params = append(params, oauth2.SetAuthURLParam("response_type", "code"))
	if acr != "" {
		params = append(params, oauth2.SetAuthURLParam("acr_values", acr))
	}
	if challenge != "" {
		params = append(params, oauth2.SetAuthURLParam("code_challenge", challenge))
		params = append(params, oauth2.SetAuthURLParam("code_challenge_method", "S256"))
	}

	return a.oauth2Cfg.AuthCodeURL(state, params...)
}

func (a *OIDCAdapter) ExchangeCode(ctx context.Context, code, codeVerifier string) (string, string, string, error) {
	var opts []oauth2.AuthCodeOption
	if codeVerifier != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	}

	oauth2Token, err := a.oauth2Cfg.Exchange(ctx, code, opts...)
	if err != nil {
		return "", "", "", fmt.Errorf("oidc: token exchange failed: %w", err)
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return "", "", "", fmt.Errorf("oidc: no id_token in token response")
	}

	return oauth2Token.AccessToken, rawIDToken, oauth2Token.RefreshToken, nil
}

type idTokenClaims struct {
	Sub               string                    `json:"sub"`
	Email             string                    `json:"email"`
	Name              string                    `json:"name"`
	PreferredUsername string                    `json:"preferred_username"`
	ACR               string                    `json:"acr"`
	ResourceAccess    map[string]clientRolesRaw `json:"resource_access"`
}

type clientRolesRaw struct {
	Roles []string `json:"roles"`
}

type AccessTokenClaims struct {
	Sub               string                    `json:"sub"`
	Email             string                    `json:"email"`
	Name              string                    `json:"name"`
	PreferredUsername string                    `json:"preferred_username"`
	ACR               string                    `json:"acr"`
	ResourceAccess    map[string]clientRolesRaw `json:"resource_access"`
}

func HasClientRole(claims *AccessTokenClaims, clientID, role string) bool {
	access, ok := claims.ResourceAccess[clientID]
	if !ok {
		return false
	}
	for _, r := range access.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func Requires2FA(claims *AccessTokenClaims, clientID string) bool {
	return HasClientRole(claims, clientID, "otp_required")
}

func HasLoA2(claims *AccessTokenClaims) bool {
	return claims.ACR == "mobo-2fa"
}

func (a *OIDCAdapter) ValidateToken(ctx context.Context, token string) (*SSOUser, error) {
	idToken, err := a.verifier.Verify(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("oidc: invalid id_token: %w", err)
	}

	var claims idTokenClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("oidc: failed to parse id_token: %w", err)
	}

	user := &SSOUser{
		ExternalID: claims.Sub,
		Email:      claims.Email,
		Name:       claims.Name,
		ACR:        claims.ACR,
	}
	if claims.PreferredUsername != "" {
		user.ExternalID = claims.PreferredUsername
	}
	return user, nil
}

func (a *OIDCAdapter) GetUserFromToken(ctx context.Context, accessToken string) (*SSOUser, error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("oidc: invalid access_token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to decode access_token: %w", err)
	}

	var claims AccessTokenClaims
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
		ExternalID:  claims.Sub,
		Email:       claims.Email,
		Name:        claims.Name,
		Roles:       roles,
		ACR:         claims.ACR,
		Requires2FA: Requires2FA(&claims, a.clientID),
	}
	if claims.PreferredUsername != "" {
		user.ExternalID = claims.PreferredUsername
	}
	return user, nil
}

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

func (a *OIDCAdapter) ClientID() string {
	return a.clientID
}
