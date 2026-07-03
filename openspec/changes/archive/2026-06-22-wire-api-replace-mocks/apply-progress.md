# Apply Progress: wire-api-replace-mocks — PR1 (T1.5–T1.7)

## Status: Complete ✅

## Completed Tasks

### T1.5 ✅ — Integrate session in root layout
- Called `ensureSession()` in `onMount` of `web/src/routes/+layout.svelte`
- Session state is shared globally via `getSession()` (Svelte 5 $state runes)
- Components can read `user`, `loading`, `error` from `getSession()`

### T1.6 ✅ — Create `/login` page
- Created `web/src/routes/login/+page.svelte`
- DEV mode: shows 3 demo user buttons (RH, Jefe, Colaborador) that call `devLogin(email)`
- PROD mode: shows spinner "Iniciando sesión con SSO..." + error message (SSO not yet configured)
- `$effect` reactively redirects to `/` when session becomes available
- Added `devLogin()` to `session.svelte.ts` with fixture fallback in DEV, API call in PROD

### T1.7 ✅ — Add logout button to Sidebar
- Added `logout()` to `session.svelte.ts`: calls `POST /auth/logout`, clears user, redirects to `/login`
- Added logout button with LogOut icon to `web/src/lib/components/Sidebar.svelte`
- Button styled consistently with nav items, turns red on hover

## Files Changed
| File | Action |
|------|--------|
| `web/src/routes/+layout.svelte` | Modified — added `onMount` calling `ensureSession()` |
| `web/src/routes/login/+page.svelte` | Created — SSO login page with DEV demo buttons |
| `web/src/lib/api/session.svelte.ts` | Modified — added `devLogin()` and `logout()` functions |
| `web/src/lib/components/Sidebar.svelte` | Modified — added logout button |
| `openspec/changes/wire-api-replace-mocks/tasks.md` | Modified — marked T1.5, T1.6, T1.7 as complete |

## Deviations from Design
- Used `$effect` for reactive redirect instead of `afterNavigate` (more reliable when session loads async)
- `devLogin()` uses fixture user map in DEV instead of trying `POST /auth/dev-login` (backend not ready in PR1)

## Issues
- No `openapi-fetch` POST typing: used `as never` cast, consistent with existing `client.GET('/auth/me' as never)` pattern

## Remaining Tasks in PR1
- PR1 fully complete (T1.1–T1.7 all done)
