# Exploration: OAuth Authentication for SED

## Current State

The platform **already has a fully functional session-based auth system**:

- **Session tokens**: 32-byte random hex tokens, SHA-256 hashed before storage (`api/internal/auth/session.go`)
- **Session table**: `sessions` with token_hash, employee_id, expires_at, is_revoked, ip_address, user_agent (migration 000003)
- **httpOnly cookie transport**: `session_token` cookie with SameSite=Strict, Secure flag when TLS
- **Bearer token fallback**: `Authorization: Bearer <token>` header also supported
- **RBAC**: 9 roles (colaborador through rh) with granular permissions (goal:*, eval:*, cycle:*, org:*, admin:*) in `api/internal/auth/rbac.go`
- **Middleware**: `RequireAuth`, `RequirePermission`, `RequireAnyPermission` in `api/internal/middleware/auth.go`
- **Auth endpoints**: POST /login, /logout, /refresh, GET /me — all wired and functional
- **Dev mode**: Email-only login (no password), 3 preset dev users, `POST /auth/dev-login`
- **Frontend**: Login page at `/login` with dev user buttons; non-dev shows "SSO no configurado aún"

### What's Missing (the gap OAuth fills)

- **No real identity provider**: Login is email-only in dev; no password, no SSO
- **SSO adapter interface exists but is unimplemented**: `api/internal/auth/sso/adapter.go` defines `SSOAdapter` with `ValidateToken`, `GetEndSessionURL`, `GetUserFromToken` — only `noopAdapter` implemented
- **No OAuth library**: go.mod has no `golang.org/x/oauth2` or `coreos/go-oidc`
- **No login UI for production**: Frontend login page has no SSO redirect flow
- **No user provisioning from OAuth**: Employees must be seeded manually; no auto-create from OAuth claims

## Affected Areas

| File/Path | Why Affected |
|-----------|-------------|
| `api/internal/auth/sso/adapter.go` | Need to implement OIDCAdapter (replace noopAdapter) |
| `api/internal/service/auth/auth_service.go` | Login() must support OAuth token validation alongside dev mode |
| `api/internal/handler/auth/auth_handler.go` | Login handler needs OAuth redirect/callback endpoints |
| `api/internal/handler/auth/routes.go` | New routes: GET /auth/sso (redirect), GET /auth/callback (OIDC callback) |
| `api/openapi/auth.yaml` | New endpoints for OAuth flow |
| `web/src/routes/login/+page.svelte` | Replace dev-only UI with SSO button + redirect |
| `web/src/lib/api/session.svelte.ts` | Handle OAuth callback, token from cookie |
| `api/cmd/server/main.go` | Initialize OAuth provider config, inject SSOAdapter |
| `api/go.mod` | Add golang.org/x/oauth2 + coreos/go-oidc |
| `.env.example` | New vars: OAUTH_CLIENT_ID, OAUTH_CLIENT_SECRET, OAUTH_ISSUER, OAUTH_REDIRECT_URL |

## Approaches

### 1. OIDC with `coreos/go-oidc` (Recommended)

Implement the existing `SSOAdapter` interface using the `coreos/go-oidc` library against a standard OIDC provider (Keycloak, Auth0, Azure AD, Google Workspace).

**Flow**: Authorization Code with PKCE
```
User → GET /auth/sso → Redirect to OIDC provider
     → User authenticates at provider
     → Provider redirects to GET /auth/callback?code=...&state=...
     → Backend exchanges code for tokens, validates id_token
     → Backend creates local session, sets httpOnly cookie
     → Redirect to frontend with session active
```

**Pros**:
- SSOAdapter interface already designed for this — minimal architecture change
- PKCE is the modern standard for SPAs (no client secret exposed)
- Works with any OIDC provider (Keycloak, Azure AD, Google, Auth0)
- Keeps session-based auth (AGENTS.md compliance: "sesión httpOnly + rotación")
- No JWT in frontend — backend validates OIDC token, issues local session
- State parameter prevents CSRF on callback

**Cons**:
- Requires OIDC provider setup (external dependency)
- Need to map OIDC claims → employee records (user provisioning strategy)
- More env vars to configure

**Effort**: Medium (3-5 days)

### 2. SAML with existing adapter pattern

Implement `SAMLAdapter` for corporate ADFS/Okta environments.

**Pros**:
- SSOAdapter interface already mentions SAMLAdapter
- Common in enterprise/corporate environments

**Cons**:
- More complex protocol, heavier libraries
- SAML is older tech; OIDC is the modern standard
- XML parsing, certificate management
- Overkill if the org can use OIDC

**Effort**: High (5-8 days)

### 3. Social OAuth (Google/Microsoft direct)

Use Google or Microsoft OAuth directly without an intermediary IdP.

**Pros**:
- Simpler setup (no Keycloak/Auth0 needed)
- Users already have Google/Microsoft accounts

**Cons**:
- No centralized user management
- Harder to enforce org-level policies
- Can't easily support multiple providers later
- Less enterprise-grade

**Effort**: Low-Medium (2-3 days)

## Recommendation

**Approach 1: OIDC with `coreos/go-oidc`** — implement the existing SSOAdapter interface.

Rationale:
1. The architecture is already prepared for this (SSOAdapter interface, noopAdapter placeholder)
2. AGENTS.md says "sesión httpOnly + rotación" — OIDC + local session satisfies this
3. OIDC is provider-agnostic: works with Keycloak (self-hosted), Azure AD (corporate), Google Workspace, or Auth0
4. PKCE flow is the standard for SPAs — no client secret in frontend
5. The local session model stays unchanged — OAuth is just the identity verification step

### Implementation Sketch

```
New files:
  api/internal/auth/sso/oidc.go          — OIDCAdapter implementation
  api/internal/auth/sso/oidc_test.go     — tests

Modified files:
  api/internal/handler/auth/auth_handler.go  — Add SSO(), Callback() handlers
  api/internal/handler/auth/routes.go        — Add /auth/sso, /auth/callback routes
  api/internal/service/auth/auth_service.go  — Add LoginWithSSO() method
  api/openapi/auth.yaml                      — New endpoints
  api/cmd/server/main.go                     — Init OIDC provider, wire adapter
  api/go.mod                                 — +golang.org/x/oauth2, +coreos/go-oidc
  web/src/routes/login/+page.svelte          — SSO button
  .env.example                               — OAUTH_* vars
```

### User Provisioning Strategy

Two options (needs decision in proposal):
- **Option A**: Employee must exist in DB first (matched by email). OAuth just verifies identity. Simpler, preserves existing RBAC model.
- **Option B**: Auto-create employee on first OAuth login. Needs defaults for org_node, profile, etc. More complex.

Recommendation: **Option A** — match by email, require employee to exist. This keeps the RBAC model clean and avoids orphan accounts.

## Key Questions / Trade-offs

1. **Which OIDC provider?** Keycloak (self-hosted, free), Azure AD (corporate), Google Workspace, or Auth0? This affects env vars and claim mapping. Decision needed in proposal.
2. **User provisioning**: Match existing employee by email (simple) vs auto-create (complex)?
3. **Dev mode preservation**: Keep dev-login alongside OAuth for local development? (Yes, recommended — gate with ENV=development)
4. **Multi-provider support**: Design for one provider now, or multi-provider from start? (One now, YAGNI)
5. **Session rotation**: Current session is 24h fixed. Should OAuth sessions have shorter TTL with refresh? (Keep current 24h, add refresh endpoint — already exists)
6. **Logout scope**: Local logout only (revoke local session) or also redirect to OIDC end_session_endpoint? (Both — use GetEndSessionURL from SSOAdapter)

## Risks

- **OIDC provider availability**: If the provider is down, no one can log in. Mitigation: dev-login remains for ENV=development.
- **Claim mapping**: OIDC claims may not map cleanly to the 9 evaluation profiles. Mitigation: email-based lookup keeps mapping in our DB.
- **Email verification**: Must trust OIDC provider's email claim. Mitigation: only accept verified email claims (`email_verified: true`).
- **CORS/callback URL config**: Misconfigured redirect URIs break the flow. Mitigation: clear env var documentation.
- **Breaking change to login flow**: Frontend login page changes. Mitigation: dev-login still works in development.

## Ready for Proposal

**Yes** — the codebase is well-prepared for OAuth. The SSOAdapter interface, session infrastructure, and RBAC model are all in place. The main work is:
1. Implement `OIDCAdapter` (~200 lines)
2. Add `/auth/sso` and `/auth/callback` handlers (~150 lines)
3. Update login page with SSO redirect (~50 lines)
4. Wire everything in main.go (~20 lines)

Total estimated effort: **3-5 days** for a competent developer.

The orchestrator should ask the user which OIDC provider they want to use (Keycloak, Azure AD, Google, Auth0) before creating the proposal.
