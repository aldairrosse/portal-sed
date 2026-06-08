# Tasks: C7 — identity-access

## Summary

| Phase | Tasks | Estimated Time |
|-------|-------|----------------|
| Phase 1: Database | 2 | ~1h |
| Phase 2: Core Auth Package | 3 | ~3h |
| Phase 3: Auth Service | 1 | ~2h |
| Phase 4: Auth Handler & Routes | 2 | ~2h |
| Phase 5: Auth Middleware | 1 | ~2h |
| Phase 6: OpenAPI Spec | 1 | ~1h |
| Phase 7: OpenSpec Artifacts | 3 | ~1h |
| **Total** | **13** | **~12h** |

---

## Phase 1: Database (~1h)

### Task 1: Session migration (up)
- **ID:** 1
- **Title:** Create 000003_add_session_tables.up.sql
- **Description:** Write the up migration creating the `sessions` table with all columns, foreign key to `employees`, and three indexes (token_hash unique, employee_id for listing, partial expires for active sessions).
- **Acceptance Criteria:**
  - [ ] `sessions` table exists with all specified columns.
  - [ ] `UNIQUE INDEX` on `token_hash`.
  - [ ] `INDEX` on `employee_id`.
  - [ ] Partial `INDEX` on `expires_at WHERE NOT is_revoked`.
  - [ ] Foreign key `fk_sessions_employee` → `employees(id)` with `ON DELETE CASCADE`.
- **Estimated Time:** 0.5h
- **Dependencies:** C1 (employees table must exist)
- **Layer:** Database

### Task 2: Session migration (down)
- **ID:** 2
- **Title:** Create 000003_add_session_tables.down.sql
- **Description:** Write the down migration dropping the sessions table and its indexes in reverse order.
- **Acceptance Criteria:**
  - [ ] `DROP INDEX` for all three indexes.
  - [ ] `DROP TABLE IF EXISTS sessions CASCADE`.
- **Estimated Time:** 0.5h
- **Dependencies:** Task 1
- **Layer:** Database

---

## Phase 2: Core Auth Package (~3h)

### Task 3: RBAC definitions
- **ID:** 3
- **Title:** Implement `internal/auth/rbac.go`
- **Description:** Define `Role` type with 9 constants, `Permission` type with granular permissions, `RolePermissions` map, and helper functions `HasPermission`, `HasAnyPermission`, `ProfileNameToRole`.
- **Acceptance Criteria:**
  - [ ] 9 Role constants defined.
  - [ ] Permission constants defined for all 6 domains (goal, competency, evaluation, cycle, org, admin).
  - [ ] `RolePermissions` maps each role to its allowed permissions per the design matrix.
  - [ ] `HasPermission(role, perm)` returns correct boolean.
  - [ ] `HasAnyPermission(role, perms...)` returns true if any match.
  - [ ] `ProfileNameToRole(name)` maps profile names correctly, falls back to `RoleColaborador`.
  - [ ] `go build` compiles without errors.
- **Estimated Time:** 1h
- **Dependencies:** None
- **Layer:** Core

### Task 4: Session store
- **ID:** 4
- **Title:** Implement `internal/auth/session.go`
- **Description:** Implement `Session` struct, `SessionStore` with `database/sql` for `Create`, `GetByToken`, `Refresh`, `Revoke`, `RevokeAllEmployeeSessions`, plus `GenerateToken` and `HashToken` helpers.
- **Acceptance Criteria:**
  - [ ] `GenerateToken()` returns 64-char hex string (32 bytes).
  - [ ] `HashToken(token)` returns SHA-256 hex digest.
  - [ ] `Create()` inserts session, returns raw token (shown once).
  - [ ] `GetByToken()` hashes token, queries DB, checks expiry + revocation.
  - [ ] `Refresh()` extends expiry + updates `last_active_at`.
  - [ ] `Revoke()` sets `is_revoked = true`.
  - [ ] `RevokeAllEmployeeSessions()` revokes all active sessions for an employee.
  - [ ] Expired or revoked sessions return `nil, nil` from `GetByToken`.
  - [ ] `go build` compiles without errors.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 1 (migration must be run)
- **Layer:** Core

### Task 5: Context helpers
- **ID:** 5
- **Title:** Implement `internal/auth/context.go`
- **Description:** Define context key constants (`SessionKey`, `EmployeeIDKey`, `RoleKey`, `ProfileIDKey`) and functions `WithSession`, `GetSession`, `GetEmployeeID`, `GetRole`, `GetProfileID`.
- **Acceptance Criteria:**
  - [ ] `WithSession` stores all four values in context.
  - [ ] `GetSession` retrieves `*Session`, returns `false` if not set.
  - [ ] `GetEmployeeID` retrieves `uuid.UUID`, returns `false` if not set.
  - [ ] `GetRole` retrieves `Role`, returns `false` if not set.
  - [ ] `GetProfileID` retrieves `uuid.UUID`, returns `false` if not set.
  - [ ] `go build` compiles without errors.
- **Estimated Time:** 0.5h
- **Dependencies:** None
- **Layer:** Core

---

## Phase 3: Auth Service (~2h)

### Task 6: Auth service
- **ID:** 6
- **Title:** Implement `internal/service/auth/auth_service.go`
- **Description:** Implement `EmployeeRow`, `EmployeeReader` interface, production `employeeReader` (SQL-backed), `AuthService` struct with `Login`, `ValidateSession`, `Logout`, `Refresh`, `Employee` methods. `EmployeeReader` also exported for testability.
- **Acceptance Criteria:**
  - [ ] `Login(email, ip, ua)` finds employee, creates session, returns `LoginResult` with role and profile info.
  - [ ] `Login` rejects inactive employees.
  - [ ] `ValidateSession(token)` validates session, looks up employee and profile, returns enriched result with role.
  - [ ] `Logout(sessionID)` revokes session.
  - [ ] `Refresh(sessionID)` extends expiry.
  - [ ] `Employee(id)` retrieves employee details.
  - [ ] `EmployeeReader` interface with `GetByID` and `GetByEmail`.
  - [ ] Production `employeeReader` uses `database/sql` queries.
  - [ ] `go build` compiles without errors.
- **Estimated Time:** 2h
- **Dependencies:** Tasks 3, 4, 5
- **Layer:** Service

---

## Phase 4: Auth Handler & Routes (~2h)

### Task 7: Auth handler
- **ID:** 7
- **Title:** Implement `internal/handler/auth/auth_handler.go`
- **Description:** Implement HTTP handlers: `Login` (POST), `Logout` (POST), `Refresh` (POST), `Me` (GET). Follow existing handler patterns (`writeJSON`, `writeError`, `generateTraceID`). Set httpOnly cookie on login, clear on logout.
- **Acceptance Criteria:**
  - [ ] `POST /login` accepts `{email}`, returns token + session + role, sets cookie.
  - [ ] `POST /login` returns 400 for empty email.
  - [ ] `POST /logout` revokes session, clears cookie, returns 200.
  - [ ] `POST /refresh` extends expiry, returns new `expires_at`.
  - [ ] `GET /me` returns employee details + role + profile.
  - [ ] All errors returned in standard `APIError` format.
  - [ ] `go build` compiles without errors.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 6
- **Layer:** Handler

### Task 8: Auth routes
- **ID:** 8
- **Title:** Implement `internal/handler/auth/routes.go`
- **Description:** Define `AuthRoutes(handler)` returning `chi.Router` with four endpoints: POST /login, POST /logout, POST /refresh, GET /me.
- **Acceptance Criteria:**
  - [ ] Chi router mounts all four endpoints.
  - [ ] No auth middleware on /login (public).
  - [ ] Returns `chi.Router` for mounting at `/api/v1/auth`.
  - [ ] `go build` compiles without errors.
- **Estimated Time:** 0.5h
- **Dependencies:** Task 7
- **Layer:** Handler

---

## Phase 5: Auth Middleware (~2h)

### Task 9: Replace auth middleware
- **ID:** 9
- **Title:** Replace `internal/middleware/auth.go`
- **Description:** Replace `AuthPlaceholder` with real middleware: `RequireAuth(authSvc)`, `RequirePermission(perm)`, `RequireAnyPermission(perms...)`. Keep backward-compatible `EmployeeIDFromContext`, `OrgIDFromContext`, `RolesFromContext` functions. Extract Bearer token from Authorization header with cookie fallback.
- **Acceptance Criteria:**
  - [ ] `RequireAuth` extracts Bearer token, validates session, injects auth context.
  - [ ] `RequireAuth` falls back to `session_token` cookie if no Authorization header.
  - [ ] `RequireAuth` returns 401 JSON for missing/invalid/expired token.
  - [ ] `RequirePermission` returns 403 JSON if role lacks permission.
  - [ ] `RequireAnyPermission` returns 403 if role has none of the permissions.
  - [ ] Backward-compatible `EmployeeIDFromContext` returns string UUID.
  - [ ] `OrgIDFromContext` and `RolesFromContext` still work.
  - [ ] `go build` compiles without errors.
- **Estimated Time:** 2h
- **Dependencies:** Tasks 5, 6
- **Layer:** Middleware

---

## Phase 6: OpenAPI Spec (~1h)

### Task 10: Auth OpenAPI spec
- **ID:** 10
- **Title:** Create `api/openapi/auth.yaml`
- **Description:** Write OpenAPI 3.1 spec with four endpoints, request/response schemas, security scheme (bearerAuth), error responses.
- **Acceptance Criteria:**
  - [ ] `POST /auth/login` defined with request body and response.
  - [ ] `POST /auth/logout` defined with security and response.
  - [ ] `POST /auth/refresh` defined with security and response.
  - [ ] `GET /auth/me` defined with security and response.
  - [ ] `LoginResponse` schema with session, token, employee, role.
  - [ ] `MeResponse` schema with employee, role, profile.
  - [ ] `bearerAuth` security scheme defined.
  - [ ] Standard `Error` schema for error responses.
  - [ ] Valid YAML syntax.
- **Estimated Time:** 1h
- **Dependencies:** None
- **Layer:** Documentation

---

## Phase 7: OpenSpec Artifacts (~1h)

### Task 11: Proposal
- **ID:** 11
- **Title:** Create proposal.md
- **Description:** Write change proposal with intent, scope, dependencies, success criteria.
- **Acceptance Criteria:**
  - [ ] Intent clearly describes auth module purpose.
  - [ ] In Scope and Out of Scope sections complete.
  - [ ] Dependencies listed.
  - [ ] Success criteria checkbox list.
- **Estimated Time:** 0.5h
- **Dependencies:** None
- **Layer:** Documentation

### Task 12: Design
- **ID:** 12
- **Title:** Create design.md
- **Description:** Write technical design with architecture overview, package structure, session flow diagrams, RBAC matrix, schema, API endpoints, wiring plan for C8.
- **Acceptance Criteria:**
  - [ ] Package structure diagram.
  - [ ] Layer dependency graph.
  - [ ] Session flow (login, auth request, logout).
  - [ ] Permission matrix (roles × permissions).
  - [ ] Table schema with indexes.
  - [ ] API endpoint table.
  - [ ] C8 wiring plan with files to update.
  - [ ] Error codes table.
- **Estimated Time:** 1h
- **Dependencies:** Tasks 1–10
- **Layer:** Documentation

### Task 13: Tasks
- **ID:** 13
- **Title:** Create tasks.md
- **Description:** Write implementation task breakdown with phases, dependencies, acceptance criteria.
- **Acceptance Criteria:**
  - [ ] All 13 tasks defined with IDs, titles, descriptions.
  - [ ] Each task has acceptance criteria checklist.
  - [ ] Dependencies between tasks clear.
  - [ ] Estimated times provided.
- **Estimated Time:** 0.5h
- **Dependencies:** Task 12
- **Layer:** Documentation

---

## Notes

- **Implementation order:** Follow phase order. Tasks within a phase can be parallelized where dependencies allow.
- **Compile gating:** Each task's Go code must compile (`go build`) before moving to the next.
- **Backward compatibility:** `EmployeeIDFromContext`, `OrgIDFromContext`, `RolesFromContext` must remain functional after middleware replacement.
- **C8 dependency:** The middleware wiring into C2–C6 route files is documented but NOT done in this change. It happens in C8 (wire-api-replace-mocks).
