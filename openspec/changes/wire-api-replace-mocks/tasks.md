# Tasks: wire-api-replace-mocks

Estrategia: **5 PRs encadenados**. Cada PR depende del anterior. Total: ~18 tareas, ~2h c/u.

---

## PR1 — Setup: cliente HTTP + tipos generados + stores de sesión/ciclo

### T1.1 ~~Instalar dependencias y configurar script de generación~~ ✅

- [x] Add `openapi-typescript ^7.x` y `openapi-fetch ^0.x` a `web/package.json` (devDependencies y dependencies respectivamente)
- [x] Agregar script `"gen:api"` que ejecute `openapi-typescript` para cada spec
- [x] Configurar `.env` con `VITE_API_URL=http://localhost:8080/api/v1` y `VITE_USE_API=false`
- **AC:** `pnpm install` succeede, `pnpm run gen:api` genera los 6 archivos `.d.ts` en `src/lib/api/schemas/`

### T1.2 ~~Crear client.ts con interceptors~~ ✅

- [x] Crear `web/src/lib/api/client.ts` con `createClient` + `baseURL` desde `VITE_API_URL` o fallback `/api/v1`
- [x] Agregar `onRequest` interceptor que setea `credentials: 'include'`
- [x] Agregar `onResponse` interceptor que redirige a `/login` en 401
- **AC:** `client.ts` compila sin errores, interceptor 401 se puede testear unitariamente

### T1.3 ~~Crear session.svelte.ts~~ ✅

- [x] Crear `web/src/lib/api/session.svelte.ts` con `$state` para `user`, `loading`, `error`
- [x] Implementar `ensureSession()` que llama `GET /auth/me` con fallback fixture en DEV
- [x] Exponer `getSession()` para componentes
- **AC:** `ensureSession()` retorna `AuthUser` mock en DEV; en producción fetch real; error state captura fallos

### T1.4 ~~Crear cycle.svelte.ts~~ ✅

- [x] Crear `web/src/lib/api/cycle.svelte.ts` con `$state` para `activePhase`, `loading`, `error`
- [x] Implementar `loadCycle()` que llama `GET /cycle/current` con fallback fixture en DEV
- [x] Exponer `getActivePhase()` y `getCycleState()`
- **AC:** `loadCycle()` settea `activePhase` correctamente; coincide con valores de `CyclePhase`

### T1.5 Integrar session en layout raíz ✅

- [x] En `web/src/routes/+layout.svelte`, llamar `ensureSession()` en `onMount`
- [x] Pasar sesión como state global vía `getSession()`
- **AC:** Al cargar la app, se invoca `ensureSession()`; si falla, componentes pueden leer `error`

### T1.6 Crear página `/login` (UI SSO limpia) ✅

- [x] Crear `web/src/routes/login/+page.svelte` — UI sin formularios, solo redirect/status
- [x] En PROD: redirect automático a SSO (futuro OIDC/SAML), mostrar spinner "Iniciando sesión con SSO..."
- [x] En DEV: botón "Acceso demo" que llama `devLogin()` a `session.svelte.ts`
- [x] `$effect` reactivo: si ya hay sesión activa, redirigir a `/`
- **AC:** `/login` renderiza sin 404; en DEV muestra botón demo; en PROD redirige a SSO; post-login redirige a home

### T1.7 Agregar logout en Sidebar ✅

- [x] Modificar `web/src/lib/components/Sidebar.svelte` — agregar botón de logout
- [x] Implementar `logout()` en `session.svelte.ts` que llama `POST /auth/logout`
- [x] Limpiar estado local (user = null) y redirigir a `/login`
- **AC:** Botón logout visible en sidebar; al hacer click, limpia sesión y redirige a `/login`

---

## PR2 — Migración de stores (5 stores async)

### T2.1 ~~Migrar goalsStore.svelte.ts~~ ✅

- [x] Reemplazar `$state` de fixtures por triplete `data`/`loading`/`error` tipado con `paths` desde `goals.d.ts`
- [x] Implementar `load()`: en DEV usa `structuredClone(fixtureGoalsData)`, en producción fetch real
- [x] Implementar `reload()` como alias de `load()`
- [x] Migrar getters (getCategories, getGoals, getKpis, etc.) a derivaciones de `data`
- [x] Migrar mutaciones (addCategory, updateGoal, etc.) a POST/PUT/DELETE + `await reload()`
- [x] Reemplazar `getCyclePhase()` de devContext por `getActivePhase()` de cycle store
- [x] **AC:** `load()` popular data desde fixture en DEV; `addCategory()` hace POST y reload; getters mantienen misma API pública

### T2.2 ~~Migrar competencyStore.svelte.ts~~ ✅

- [x] Triplete `data`/`loading`/`error` con `StoreData` interface
- [x] `load()` con guard DEV + 3 requests paralelos (pillars + levels + acceptance-levels), con fase 2 de competencias por pilar
- [x] Migrar getters a derivaciones null-safe desde `data`
- [x] Migrar mutaciones a async: pillars (POST/PUT/DELETE), competencies (POST/PUT/DELETE), acceptance levels (POST), otras local-only
- [x] Agregar `CompetencyPaths` a `client.ts` para tipado de endpoints
- **AC:** Misma API pública; `load()` executa requests en fases; mutations via API + reload en modo producción

### T2.3 ~~Migrar evaluationStore.svelte.ts~~ ✅

- [x] Triplete `data`/`loading`/`error` con tipo `StoreData` desde `evaluations.d.ts`
- [x] `load()` con guard DEV que carga fixtures + merge RH; en PROD llama `GET /evaluations/{id}`
- [x] `reload()` como alias de `load()`
- [x] Getters migrados a derivaciones de `data` (getCompetencyRatings, getGoalClosures, getEvaluationStatus)
- [x] Mutaciones migradas a `async`: `rateCompetency`, `closeGoal`, `rhRateCompetency`, `rhAssessGoal`, `addManagerComment`
- [x] Nuevas funciones `submitSelfEvaluation()`, `submitRHEvaluation()`, `finalizeEvaluation()` con llamadas POST + reload
- [x] Merge de RH evaluations en competencyRatings se mantiene como lógica local post-fetch
- [x] Referencia a `getPhase()` de devContext reemplazada por `getActivePhase()` de cycle store
- **AC:** Carga fixture en DEV; submit via POST; reload tras mutación

### T2.4 ~~Migrar nineBoxStore.svelte.ts~~ ✅

- [x] Triplete `data`/`loading`/`error` con tipo desde `evaluations.d.ts`
- [x] `load()` hace `GET /nine-box/matrices`, `GET /nine-box/quadrants`
- [x] `computeQuadrant()` se mantiene como función pura (sin cambios)
- [x] Getters migrados a derivaciones de `data?.entries ?? []`, `data?.quadrantDefs ?? []`
- [x] Mutaciones migradas a async con DEV fallback local y API call + reload
- **AC:** computeQuadrant funciona igual; load obtiene matrices reales en prod

### T2.5 ~~Migrar orgHierarchyStore.svelte.ts~~ ✅

- [x] Triplete `data`/`loading`/`error`
- [x] `load()`: DEV fixture → structuredClone; API → descubre árbol vía `/org-trees`, luego fetch `/org-trees/{treeId}/nodes?format=nested&depth=-1`
- [x] Traversal helpers (findNode, getDescendants, getLeafIds) se mantienen como funciones puras sobre `data`
- [x] Getters migrados a `data?.x ?? fallback`
- [x] `reload()` alias, `isLoading()`, `getError()` para componentes (T3.5)
- **AC:** Árbol se carga desde API en prod; traversal helpers funcionan idéntico

---

## PR3 — Componentes: loading/error states

### T3.1 ~~Actualizar componentes de goals~~ ✅

- [x] Exportar `loading` y `error` desde goalsStore
- [x] Página de asignación: envolver contenido con `{#if loading}`/`{:else if error}`/`{:else}`
- [x] Biblioteca de KPI: envolver contenido con loading/error states
- [x] Reemplazar `alert alert-error` inline por `notifications.error()` en componentes que consumen goalsStore
- **AC:** Cada componente muestra skeleton durante carga, error state en fallo, contenido normal en éxito

### T3.2 Actualizar componentes de evaluación ✅

- [x] En `web/src/lib/components/evaluation/*.svelte`, agregar loading/error wrappers
- [x] `EmployeeEvaluationTable.svelte`, `CompetencyRatingCard.svelte`, `EmployeeEvaluationDetail.svelte`, `GoalClosureCard.svelte`
- [x] Reemplazar inline error alerts con notificaciones en `EmployeeEvaluationDetail`
- [x] Agregar `isLoading()` y `getError()` a `evaluationStore.svelte.ts`
- **AC:** Mismo comportamiento que T3.1 para evaluación

### T3.3 Actualizar componentes de competency ✅

- [x] Exportar `isLoading()` y `getError()` desde competencyStore
- [x] Pages de pillars/competencias usan loading/error del store con PageSkeleton/ErrorState
- [x] Reemplazar success alerts por `notifications.success()`
- **AC:** Mismo comportamiento

### T3.4 Actualizar componentes de nine-box ✅

- [x] Exportar `isLoading()` y `getError()` desde nineBoxStore
- [x] NineBoxMatrix envuelto con PageSkeleton / ErrorState (reload callback)
- [x] NineBoxSliders usa `isLoading()` para auto-disabled
- [x] Page 9x9 maneja loading/error antes de EmptyState
- **AC:** Mismo comportamiento

### T3.5 ~~Actualizar componentes de org-hierarchy~~ ✅

- [x] En `web/src/lib/components/org-hierarchy/*.svelte`, agregar loading/error wrappers
- [x] `OrgHierarchyTree.svelte`: props `loading`, `error`, `onretry` con `PageSkeleton`/`ErrorState`
- [x] `jerarquia/+page.svelte`: usa `isLoading()`, `getError()`, `reload()` del store
- **AC:** Mismo comportamiento

### T3.6 Reemplazar getPhase() de devContext por cycle store en todos los componentes ✅

- [x] Buscar todos los usos de `getPhase()` de `$lib/stores/devContext.svelte` en componentes
- [x] Reemplazar por `getActivePhase()` de `$lib/api/cycle.svelte`
- [x] En DEV, `getActivePhase()` retorna el valor de `devContext` (cycle store hace fallback a devContext cuando `import.meta.env.DEV && !VITE_USE_API`)
- [x] **AC:** Cero imports de devContext en componentes de producción; ciclo funcional en ambos modos

---

## PR4 — Backend: RequireAuth wiring

### T4.1 ~~Reemplazar AuthPlaceholder en goal/routes.go~~ ✅

- [x] Cambiar `r.Use(middleware.AuthPlaceholder)` por `r.Use(middleware.RequireAuth(authSvc))`
- [x] Agregar `RequirePermission` en grupos de endpoints según permiso
- [x] Actualizar firma de `NewRouter` para recibir `*svc.AuthService`
- **AC:** `NewRouter(handler, authSvc)` compila; endpoints de lectura requieren `PermissionGoalRead`; escritura require `PermissionGoalWrite`

### T4.2 ~~Reemplazar AuthPlaceholder en evaluation/routes.go~~ ✅

- [x] Mismo patrón que T4.1
- [x] Endpoints de RH evaluation requieren `RequirePermission(auth.PermEvalRH)`; nine-box `RequirePermission(auth.PermEval9x9)`
- **AC:** Compila; permisos diferenciados por endpoint

### T4.3 ~~Reemplazar AuthPlaceholder en competency/routes.go~~ ✅

- [x] Mismo patrón
- [x] Endpoints de solo lectura (GET) sin permiso extra (basta `RequireAuth`); escritura con `RequirePermission(auth.PermCompetencyWrite)`
- **AC:** Compila

### T4.4 ~~Reemplazar AuthPlaceholder en cycle/routes.go~~ ✅

- [x] Mismo patrón
- **AC:** Compila

### T4.5 ~~Reemplazar AuthPlaceholder en org/routes.go~~ ✅

- [x] Mismo patrón
- **AC:** Compila

### T4.6 ~~Actualizar main.go con inyección de dependencias~~ ✅

- [x] En `api/cmd/server/main.go`, pasar `authSvc` a cada `NewRouter(...)` y `RegisterRoutes(...)`
- **AC:** `go build ./...` compila sin errores; `AuthPlaceholder` ya no se referencia en routers

### T4.7 ~~Crear dev auth service temporal~~ ✅

- [x] Crear `api/internal/auth/dev/service.go` — servicio temporal que autentica sin SSO real
- [x] Usuarios preset: `dev-rh@empresa.com` (RH), `dev-jefe@empresa.com` (Jefe), `dev-colaborador@empresa.com` (Colaborador)
- [x] Cada usuario tiene roles y permisos predefinidos (misma estructura que AuthUser)
- [x] Función `CreateDevSession` crea sesión real en DB para preset users
- [x] Solo activo cuando `ENV=development` (guard por handler)
- **AC:** `go build ./...` compila

### T4.8 ~~Interfaz SSOAdapter (preparación para futuro)~~ ✅

- [x] Crear `api/internal/auth/sso/adapter.go` — interfaz pluggable
- [x] Métodos: `ValidateToken`, `GetEndSessionURL`, `GetUserFromToken`
- [x] Incluye `noopAdapter` como placeholder que redirige a `/login`
- [x] Documentar en código cómo conectar OIDC/SAML en futuro
- **AC:** Interfaz compilable; documentación clara en comments

---

## PR5 — Tests y verificación

### T5.1 ~~Tests de client.ts~~ ✅

- [x] Crear `web/src/lib/api/client.test.ts`
- [x] Testear que interceptor 401 invoca `window.location.href = '/login'`
- [x] Testear que `baseURL` usa `VITE_API_URL` cuando está seteado
- [x] Testear que `credentials: 'include'` se aplica en cada request
- **AC:** Tests pasan con Vitest ✅

### T5.2 Tests de goalsStore (dev fallback + API load)

- Crear `web/src/lib/stores/__tests__/goalsStore.test.ts`
- Testear que en `import.meta.env.DEV` sin `VITE_USE_API`, `load()` carga desde fixture
- Testear que en modo API, `load()` llama al endpoint correcto
- Testear que `addCategory()` llama POST y luego reload
- **AC:** Tests unitarios pasan

### T5.3 Test de integración backend: 401 sin token ✅

- [x] Crear `api/integration/auth_test.go` con test `TestUnauthenticatedRequestsReturn401`
- [x] Testea GET/POST a cycles, evaluations y org-trees sin credenciales → `401 Unauthorized`
- [x] Verifica response body es JSON con campo `error.message`
- [x] Repara `setupTestServer` envolviendo `RegisterRoutes` en `Group` (Chi no permite `Use` tras rutas)
- [x] Actualiza `routes_test.go` para incluir `401` en `AllowStatus` de rutas protegidas
- **AC:** `go test ./...` pasa sin errores

---

## Resumen de PRs

| PR | Archivos | Tareas | Depende de |
|---|---|---|---|
| PR1 — Setup | ~10 archivos | T1.1–T1.7 | Ninguna |
| PR2 — Stores | 5 archivos | T2.1–T2.5 | PR1 |
| PR3 — Componentes | ~10 archivos | T3.1–T3.6 | PR2 |
| PR4 — Backend Auth | ~8 archivos | T4.1–T4.8 | Ninguna (independiente) |
| PR5 — Tests | 3 archivos | T5.1–T5.3 | PR1 + PR2 + PR4 |

PR4 puede hacerse en paralelo con PR2/PR3 porque no comparte archivos con el frontend.

Orden recomendado de merge: PR1 → PR2 → PR3 (serie), PR4 en paralelo, PR5 al final.
