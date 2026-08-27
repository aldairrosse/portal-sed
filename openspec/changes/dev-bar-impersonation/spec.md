# Spec — Dev Bar Impersonation

## Endpoints

### GET /api/v1/dev/status
- Guard: if `ENV`/`APP_ENV` != `development` → 404.
- 200 body: `{ "enabled": true }`

### GET /api/v1/dev/employees
- Guard: 404 when not dev.
- Query: `SELECT e.id, e.first_name, e.last_name, e.email, COALESCE(ep.name,'') AS profile_name FROM employees e LEFT JOIN evaluation_profiles ep ON e.profile_id = ep.id WHERE e.is_active = true ORDER BY e.last_name, e.first_name LIMIT 100`
- 200 body: `{ "employees": [{ "id","first_name","last_name","profile_name","email" }] }`

### POST /api/v1/dev/impersonate
- Guard: 404 when not dev.
- Body: `{ "employee_id": "uuid" }` — 400 if missing/invalid.
- Validates employee exists + is_active; 404 if not found.
- Creates real session via `auth.SessionStore.Create(ctx, employeeID, ip, ua, "", "", "", "", false, time.Time{}, time.Time{})`.
- Sets `session_token` cookie: `HttpOnly, SameSite=Lax, Path=/, Expires=session.ExpiresAt, Secure=(TLS||X-Forwarded-Proto==https)`.
- 200 body: `{ "employee_id","profile_name","role" }` where role = `ProfileNameToRole(profile_name)`.

## Frontend

- `DevBar.svelte`:
  - On mount: `fetch("/api/v1/dev/status", {credentials:"include"})` — if not 200, render nothing.
  - If 200: fetch employees `GET /api/v1/dev/employees`, render `<select>` + Impersonate/Clear buttons.
  - On impersonate: `POST /api/v1/dev/impersonate` with selected id; on success call `ensureSession()` then `location.reload()`.
  - Clear = `POST /api/v1/auth/logout` or reload (session cleared).

## Security

- No dev routes mounted or all return 404 when `ENV`/`APP_ENV` != `development`.
- No bypass of RBAC: session role derived from employee's real `evaluation_profiles.name`.
