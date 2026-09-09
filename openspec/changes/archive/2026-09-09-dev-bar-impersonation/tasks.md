# Tasks: Dev Bar Impersonation

- [ ] 1. API `api/internal/auth/dev/handler.go` — status, list employees (real DB), impersonate with env guard + session cookie
- [ ] 2. API `api/internal/auth/dev/routes.go` — `RegisterRoutes` returning chi.Router, guard inside handlers (404 when not dev)
- [ ] 3. API mount — conditional mount in `api/cmd/server/main.go` (isolated block, removable)
- [ ] 4. Frontend `web/src/lib/components/DevBar.svelte` — status probe, employee selector, impersonate → ensureSession + reload
- [ ] 5. Frontend mount — import DevBar in `web/src/routes/+layout.svelte` conditionally after status probe (or top-level conditional)
- [ ] 6. Verify — `ENV=development` shows bar + impersonation works; prod returns 404

> Ponytrail: all dev logic in `dev/` folder; delete folder + 5-line mount to remove.
