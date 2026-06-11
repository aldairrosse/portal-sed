// Package sso defines the pluggable SSO adapter interface for future
// integration with OIDC, SAML, or LDAP providers.
//
// Currently, no implementation is active (adapter is nil in main.go).
// Logout always redirects to /login until SSO is configured.
//
// # Future integration guide
//
// To connect a real SSO provider:
//
//  1. Implement the SSOAdapter interface for your provider:
//     - OIDCAdapter for Keycloak, Auth0, Azure AD (see oidc.go)
//     - SAMLAdapter for ADFS, Okta (see saml.go)
//     - LDAPAdapter for corporate directory (see ldap.go)
//
//  2. Inject the adapter into AuthHandler:
//     authH := authhandler.NewAuthHandler(authSvc, ssoAdapter)
//
//  3. Update POST /auth/login to redirect to SSO provider:
//     - Call ValidateToken with the OIDC/SAML response
//     - Extract user info via GetUserFromToken
//     - Create local session and set httpOnly cookie
//
//  4. Update POST /auth/logout to call GetEndSessionURL:
//     - Return the provider's end-session URL as the redirect target
//
// See c8-replace-phase.md section 3 for the full logout flow with SSO.
package sso

import "context"

// SSOAdapter defines the interface for external SSO providers.
// Implement this interface to integrate with OIDC, SAML, or LDAP.
type SSOAdapter interface {
	// ValidateToken verifies an external token from the SSO provider
	// (OIDC id_token, SAML assertion, etc.). Returns the parsed user
	// info on success, or an error if the token is invalid/expired.
	ValidateToken(ctx context.Context, token string) (*SSOUser, error)

	// GetEndSessionURL returns the provider's end-session (logout) URL
	// for the given session. The frontend redirects the user to this URL
	// to log out of the SSO provider as well.
	//
	// Returns the local /login URL as fallback if SSO logout is not
	// available or if the provider does not support SLO (single logout).
	GetEndSessionURL(ctx context.Context, sessionID string) (string, error)

	// GetUserFromToken extracts user information from an already-validated
	// SSO token. This includes identity details (name, email) and
	// authorization claims (roles, groups, organization).
	GetUserFromToken(ctx context.Context, token string) (*SSOUser, error)
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

// Compile-time check: ensure future implementations satisfy the interface.
var _ SSOAdapter = (*noopAdapter)(nil)

// noopAdapter is a placeholder implementation that always falls back
// to local login. Used when no SSO provider is configured.
type noopAdapter struct{}

// NewNoopAdapter creates a no-op SSO adapter that returns local URLs.
// This is the default adapter when no SSO provider is configured.
func NewNoopAdapter() SSOAdapter {
	return &noopAdapter{}
}

// ValidateToken always returns an error — no SSO provider configured.
func (a *noopAdapter) ValidateToken(_ context.Context, _ string) (*SSOUser, error) {
	return nil, errNoSSO
}

// GetEndSessionURL returns the local /login URL as fallback.
func (a *noopAdapter) GetEndSessionURL(_ context.Context, _ string) (string, error) {
	return "/login", nil
}

// GetUserFromToken always returns an error — no SSO provider configured.
func (a *noopAdapter) GetUserFromToken(_ context.Context, _ string) (*SSOUser, error) {
	return nil, errNoSSO
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
