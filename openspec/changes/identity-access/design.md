# Design: C7 — identity-access

## 1. Overview

C7 delivers the authentication and role-based access control (RBAC) module
for the SED evaluation platform. It replaces the `AuthPlaceholder` middleware
used in C2–C6 with real session validation and permission checks.

### Key Decisions

- **Session tokens:** Random 32-byte hex tokens, SHA-256 hashed before storage.
  Only the hash is persisted; the raw token is shown once at login.
- **Token transport:** `Authorization: Bearer <token>` header, with fallback to
  `session_token` httpOnly cookie.
- **RBAC model:** Roles map to evaluation profiles in the database. Each role
  has a static set of permissions defined in code.
- **Dev mode:** Login accepts email only, no password. Marked `TODO(auth:prod)`.
- **Stateless middleware:** All session state lives in PostgreSQL; middleware
  queries the DB on every request (cachable via Redis in production).

---

## 2. Package Structure

```
api/
├── migrations/
│   ├── 000003_add_session_tables.up.sql
│   └── 000003_add_session_tables.down.sql
├── internal/
│   ├── auth/
│   │   ├── session.go          # SessionStore (DB operations)
│   │   ├── rbac.go             # Roles, permissions, RolePermissions
│   │   └── context.go          # Context helpers (WithSession, GetEmployeeID, etc.)
│   ├── service/auth/
│   │   └── auth_service.go     # AuthService (Login, ValidateSession, Logout, Refresh)
│   ├── handler/auth/
│   │   ├── auth_handler.go     # HTTP handlers (Login, Logout, Refresh, Me)
│   │   └── routes.go           # Chi router wiring
│   └── middleware/
│       └── auth.go             # RequireAuth, RequirePermission, RequireAnyPermission
├── openapi/
│   └── auth.yaml               # OpenAPI 3.1 spec
```

### Layer Dependencies

```
handler/auth → service/auth → auth (session, rbac, context)
                               ↕
                           database/sql

middleware/auth → service/auth → auth
                → pkg/errors
```

No circular dependencies. The `auth` package has zero dependencies on handler,
service, or middleware packages.

---

## 3. Session Flow

### Login
```
Client                     Server
  │    POST /auth/login      │
  │    { email }             │
  │                          ├─ Find employee by email
  │                          ├─ Lookup evaluation profile → map to Role
  │                          ├─ Generate 32-byte random token
  │                          ├─ SHA-256 hash → store in sessions table
  │    ← 200 { token,       │
  │      session, role }     │
  │    ← Set-Cookie:         │
  │      session_token=...   │
```

### Authenticated Request
```
Client                     Server
  │    GET /resource         │
  │    Authorization:        │
  │      Bearer <token>      │
  │                          ├─ RequireAuth middleware:
  │                          │   1. Extract token from header
  │                          │   2. SHA-256 hash
  │                          │   3. Query sessions table
  │                          │   4. Check not expired, not revoked
  │                          │   5. Lookup employee + profile → Role
  │                          │   6. Inject Session, EmployeeID, Role, ProfileID into context
  │                          ├─ RequirePermission middleware:
  │                          │   1. Get Role from context
  │                          │   2. Check RolePermissions map
  │                          │   3. 403 if permission missing
  │                          ├─ Handler executes
  │    ← 200 { ... }         │
```

### Logout
```
Client                     Server
  │    POST /auth/logout     │
  │    Authorization: ...    │
  │                          ├─ RequireAuth validates session
  │                          ├─ Revoke: UPDATE sessions SET is_revoked = true
  │                          ├─ Clear session cookie
  │    ← 200 { message }    │
```

---

## 4. RBAC Model

### Role ↔ Profile Mapping

| Database `evaluation_profiles.name` | Role constant | Description |
|---|---|---|
| `colaborador` | `RoleColaborador` | Individual contributor |
| `jefe` | `RoleJefe` | Manager / supervisor |
| `vendedor` | `RoleVendedor` | Sales role |
| `gerente-tienda` | `RoleGerenteTienda` | Store manager |
| `divisional` | `RoleDivisional` | Division head |
| `regional` | `RoleRegional` | Regional manager |
| `director` | `RoleDirector` | Director |
| `director-general` | `RoleDirectorGeneral` | CEO / General director |
| `rh` | `RoleRH` | HR administrator |

### Permission Matrix

| Permission | colaborador | jefe | vendedor | gerente | divisional | regional | director | dir-general | rh |
|---|---|---|---|---|---|---|---|---|---|
| goal:create | ✓ |   | ✓ |   |   |   |   |   | ✓ |
| goal:read | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| goal:update | ✓ |   | ✓ |   |   |   |   |   | ✓ |
| goal:delete | ✓ |   | ✓ |   |   |   |   |   | ✓ |
| goal:progress | ✓ |   | ✓ |   |   |   |   |   | ✓ |
| competency:read | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| competency:write |   |   |   |   |   |   |   |   | ✓ |
| competency:delete |   |   |   |   |   |   |   |   | ✓ |
| eval:self | ✓ |   | ✓ |   |   |   |   |   | ✓ |
| eval:rh |   |   |   |   |   |   |   |   | ✓ |
| eval:9x9 |   | ✓ |   | ✓ | ✓ | ✓ | ✓ |   |   |
| eval:read | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| cycle:read | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| cycle:transition |   |   |   |   |   |   | ✓ | ✓ | ✓ |
| org:read | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| org:write |   |   |   |   |   |   | ✓ | ✓ |   |
| admin:all |   |   |   |   |   |   |   | ✓ |   |

---

## 5. Session Table Schema

```sql
CREATE TABLE sessions (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID        NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    token_hash      TEXT        NOT NULL,
    ip_address      INET        NULL,
    user_agent      TEXT        NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_active_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_revoked      BOOLEAN     NOT NULL DEFAULT false
);

CREATE UNIQUE INDEX idx_sessions_token_hash ON sessions (token_hash);
CREATE INDEX idx_sessions_employee ON sessions (employee_id);
CREATE INDEX idx_sessions_expires ON sessions (expires_at) WHERE NOT is_revoked;
```

- **token_hash:** SHA-256 hex digest (64 chars). Unique index for fast lookup.
- **expires_at:** Default 24 hours from creation; extended on refresh.
- **is_revoked:** Soft delete for logout. Queries filter by `NOT is_revoked`.
- **expires index:** Partial index on active sessions for cleanup jobs.

---

## 6. API Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/auth/login` | None | Dev login (email only) |
| POST | `/api/v1/auth/logout` | RequireAuth | Revoke session |
| POST | `/api/v1/auth/refresh` | RequireAuth | Extend session expiry |
| GET | `/api/v1/auth/me` | RequireAuth | Current user info |

---

## 7. Files to Update in C8 (wire-api-replace-mocks)

The following route files currently use `AuthPlaceholder` and need to switch to `RequireAuth`:

| File | Current | Replace with |
|---|---|---|
| `api/internal/handler/cycle/routes.go` | `r.Use(middleware.AuthPlaceholder)` | `r.Use(middleware.RequireAuth(authSvc))` |
| `api/internal/handler/goal/routes.go` | `r.Use(middleware.AuthPlaceholder)` | `r.Use(middleware.RequireAuth(authSvc))` |
| `api/internal/handler/evaluation/routes.go` | `TODO(auth:C7)` | `r.Use(middleware.RequireAuth(authSvc))` |
| `api/internal/handler/competency/routes.go` | `TODO(auth:C7)` | `r.Use(middleware.RequireAuth(authSvc))` |
| `api/internal/handler/org/routes.go` | `TODO(auth:C7)` | `r.Use(middleware.RequireAuth(authSvc))` |

Additionally, `cmd/server/main.go` needs to:
1. Initialize `*sql.DB` connection
2. Create `auth.NewSessionStore(db)`
3. Create `svcauth.NewEmployeeReader(db)`
4. Create `svcauth.NewAuthService(sessionStore, employeeReader, db)`
5. Create `authhandler.NewAuthHandler(authSvc)`
6. Register auth routes: `r.Mount("/api/v1/auth", authhandler.AuthRoutes(...))`
7. Pass `authSvc` to each `NewRouter` call so routes can use `RequireAuth(authSvc)`

---

## 8. Error Codes

| Code | HTTP Status | Description |
|---|---|---|
| `INVALID_REQUEST` | 400 | Missing/malformed email, invalid JSON |
| `UNAUTHORIZED` | 401 | Missing token, expired session, revoked session |
| `FORBIDDEN` | 403 | Insufficient permissions for the resource |
| `EMPLOYEE_NOT_FOUND` | 404 | Email not found (login) |
