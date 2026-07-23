// Package sso defines the SSO adapter interface for OIDC providers (Keycloak).
//
// Active implementation: OIDCAdapter (see oidc.go).
// Logout redirects to the provider's end-session endpoint.
//
// To add a new provider type (SAML, LDAP):
//  1. Implement SSOAdapter interface
//  2. Wire it in main.go via NewAuthHandler(authSvc, ssoAdapter)
package sso

import "context"

// SSOAdapter defines the interface for external SSO providers.
// Implement this interface to integrate with OIDC, SAML, or LDAP.
type SSOAdapter interface {
	// AuthorizationURL returns the provider's login URL with the given
	// CSRF state parameter. The handler redirects the user here.
	AuthorizationURL(state string) string

	// ExchangeCode exchanges an OIDC authorization code for tokens.
	// Returns the access_token, id_token, and refresh_token strings.
	ExchangeCode(ctx context.Context, code string) (accessToken, idToken, refreshToken string, err error)

	// ValidateToken verifies an external token from the SSO provider
	// (OIDC id_token, SAML assertion, etc.). Returns the parsed user
	// info on success, or an error if the token is invalid/expired.
	ValidateToken(ctx context.Context, token string) (*SSOUser, error)

	// GetUserFromToken extracts user information from the access_token.
	// This includes identity details (name, email), authorization claims
	// (roles), and checks for the "access" role via HasAccess.
	GetUserFromToken(ctx context.Context, token string) (*SSOUser, error)

	// HasAccess checks whether the access_token grants access to this
	// client (the "access" role in resource_access).
	HasAccess(ctx context.Context, accessToken string) bool

	// ValidateAccessToken calls the provider's UserInfo endpoint to check
	// whether the access token is still valid (not expired, not revoked).
	// Returns nil if valid, error otherwise.
	ValidateAccessToken(ctx context.Context, accessToken string) error

	// GetEndSessionURL returns the provider's end-session (logout) URL
	// for the given id_token_hint. The frontend redirects the user to this
	// URL to log out of the SSO provider as well.
	//
	// Returns the local /login URL as fallback if SSO logout is not
	// available or if the provider does not support SLO (single logout).
	GetEndSessionURL(ctx context.Context, idTokenHint string) (string, error)

	// RevokeToken sends a token (typically a refresh_token) to the
	// provider's revocation endpoint so it can no longer be used to
	// obtain new access tokens. Returns nil on success.
	RevokeToken(ctx context.Context, token string) error
}

// SSOUser represents a user authenticated via an external SSO provider.
type SSOUser struct {
	// ExternalID is the unique identifier from the SSO provider
	// (e.g., the "sub" claim in OIDC, or the NameID in SAML).
	ExternalID string

	// Email is the verified email address from the provider.
	Email string

	// Name is the display name from the provider.
	Name string

	// Roles are the authorization roles/groups mapped from the
	// provider's claims (e.g., OIDC "groups" claim, SAML attribute).
	Roles []string

	// OrganizationID identifies the organization in the local system.
	OrganizationID string
}

// Compile-time checks: ensure implementations satisfy the interface.
var _ SSOAdapter = (*noopAdapter)(nil)
var _ SSOAdapter = (*OIDCAdapter)(nil)

// noopAdapter is a placeholder implementation that always falls back
// to local login. Used when no SSO provider is configured.
type noopAdapter struct{}

// NewNoopAdapter creates a no-op SSO adapter that returns local URLs.
// This is the default adapter when no SSO provider is configured.
func NewNoopAdapter() SSOAdapter {
	return &noopAdapter{}
}

// AuthorizationURL returns an empty string — no SSO provider configured.
func (a *noopAdapter) AuthorizationURL(_ string) string { return "" }

// ExchangeCode always returns an error — no SSO provider configured.
func (a *noopAdapter) ExchangeCode(_ context.Context, _ string) (string, string, string, error) {
	return "", "", "", errNoSSO
}

// ValidateToken always returns an error — no SSO provider configured.
func (a *noopAdapter) ValidateToken(_ context.Context, _ string) (*SSOUser, error) {
	return nil, errNoSSO
}

// GetUserFromToken always returns an error — no SSO provider configured.
func (a *noopAdapter) GetUserFromToken(_ context.Context, _ string) (*SSOUser, error) {
	return nil, errNoSSO
}

// HasAccess always returns false — no SSO provider configured.
func (a *noopAdapter) HasAccess(_ context.Context, _ string) bool { return false }

// ValidateAccessToken always returns an error — no SSO provider configured.
func (a *noopAdapter) ValidateAccessToken(_ context.Context, _ string) error {
	return errNoSSO
}

// GetEndSessionURL returns the local /login URL as fallback.
func (a *noopAdapter) GetEndSessionURL(_ context.Context, _ string) (string, error) {
	return "/login", nil
}

// RevokeToken always returns an error — no SSO provider configured.
func (a *noopAdapter) RevokeToken(_ context.Context, _ string) error {
	return errNoSSO
}

// errNoSSO is returned when no SSO provider is configured.
var errNoSSO = &ErrNoSSOConfigured{}

// ErrNoSSOConfigured indicates that no SSO provider has been configured.
type ErrNoSSOConfigured struct{}

func (e *ErrNoSSOConfigured) Error() string {
	return "no SSO provider configured"
}

// Is provides compatibility with errors.Is.
func (e *ErrNoSSOConfigured) Is(target error) bool {
	_, ok := target.(*ErrNoSSOConfigured)
	return ok
}
