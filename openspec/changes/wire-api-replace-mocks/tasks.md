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

### T1.5 Integrar session en layout raíz

- En `web/src/routes/+layout.svelte` (o +layout.ts), llamar `ensureSession()` antes de renderizar
- Pasar sesión vía context o directamente como state global
- **AC:** Al cargar la app, se invoca `ensureSession()`; si falla, componentes pueden leer `error`

---

## PR2 — Migración de stores (5 stores async)

### T2.1 Migrar goalsStore.svelte.ts

- Reemplazar `$state` de fixtures por triplete `data`/`loading`/`error` tipado con `paths` desde `goals.d.ts`
- Implementar `load()`: en DEV usa `structuredClone(fixtureGoalsData)`, en producción fetch real
- Implementar `reload()` como alias de `load()`
- Migrar getters (getCategories, getGoals, getKpis, etc.) a derivaciones de `data`
- Migrar mutaciones (addCategory, updateGoal, etc.) a POST/PUT/DELETE + `await reload()`
- Reemplazar `getCyclePhase()` de devContext por `getActivePhase()` de cycle store
- **AC:** `load()` populat data desde fixture en DEV; `addCategory()` hace POST y reload; getters mantienen misma API pública

### T2.2 Migrar competencyStore.svelte.ts

- Triplete `data`/`loading`/`error` con tipo desde `competency.d.ts`
- `load()` hace `GET /pillars`, `GET /competencies`, `GET /levels`, `GET /acceptance-levels` (o endpoints batch si existen)
- Migrar getters y mutaciones al patrón async
- **AC:** Misma API pública; `load()` executa 2-4 requests paralelos; mutations via API + reload

### T2.3 Migrar evaluationStore.svelte.ts

- Triplete `data`/`loading`/`error` con tipo desde `evaluations.d.ts`
- `load()` hace `GET /evaluations` + `GET /evaluations/summary`
- Migrar `submitSelfEvaluation()`, `submitRHEvaluation()`, `finalizeEvaluation()` a llamadas API
- Merge de RH evaluations en competencyRatings se mantiene como lógica local post-fetch
- **AC:** Carga fixture en DEV; submit via POST; reload tras mutación

### T2.4 Migrar nineBoxStore.svelte.ts

- Triplete `data`/`loading`/`error` con tipo desde `evaluations.d.ts` (o spec propio nine-box)
- `load()` hace `GET /nine-box/matrices`, `GET /nine-box/scales`, `GET /nine-box/quadrants`
- Migrar `computeQuadrant()` como función pura (sin cambios)
- **AC:** computeQuadrant funciona igual; load obtiene matrices reales en prod

### T2.5 Migrar orgHierarchyStore.svelte.ts

- Triplete `data`/`loading`/`error`
- `load()` hace `GET /org-tree`
- Traversal helpers (findNode, getDescendants, getLeafIds) se mantienen como funciones puras sobre `data`
- **AC:** Árbol se carga desde API en prod; traversal helpers funcionan idéntico

---

## PR3 — Componentes: loading/error states

### T3.1 Actualizar componentes de goals

- En componentes de `web/src/lib/components/goals/`, agregar `{#if loading}` con `<PageSkeleton />`
- Agregar `{:else if error}` con `<ErrorState message={error} onretry={load} />`
- Reemplazar imports de `getPhase()` de devContext por `getActivePhase()` de cycle store
- **AC:** Cada componente muestra skeleton durante carga, error state en fallo, contenido normal en éxito

### T3.2 Actualizar componentes de evaluación

- En `web/src/lib/components/evaluation/*.svelte`, agregar loading/error wrappers
- `EmployeeEvaluationTable.svelte`, `CompetencyRatingCard.svelte`, `EmployeeEvaluationDetail.svelte`, `GoalClosureCard.svelte`
- **AC:** Mismo comportamiento que T3.1 para evaluación

### T3.3 Actualizar componentes de competency

- En `web/src/lib/components/competency/*.svelte`, agregar loading/error wrappers
- **AC:** Mismo comportamiento

### T3.4 Actualizar componentes de nine-box

- En `web/src/lib/components/nine-box/*.svelte`, agregar loading/error wrappers
- **AC:** Mismo comportamiento

### T3.5 Actualizar componentes de org-hierarchy

- En `web/src/lib/components/org-hierarchy/*.svelte`, agregar loading/error wrappers
- **AC:** Mismo comportamiento

### T3.6 Reemplazar getPhase() de devContext por cycle store en todos los componentes

- Buscar todos los usos de `getPhase()` de `$lib/stores/devContext.svelte` en componentes
- Reemplazar por `getActivePhase()` de `$lib/api/cycle.svelte`
- En DEV, `getActivePhase()` retorna el valor de `devContext` (cycle store hace fallback a devContext cuando `import.meta.env.DEV && !VITE_USE_API`)
- **AC:** Cero imports de devContext en componentes de producción; ciclo funcional en ambos modos

---

## PR4 — Backend: RequireAuth wiring

### T4.1 Reemplazar AuthPlaceholder en goal/routes.go

- Cambiar `r.Use(middleware.AuthPlaceholder)` por `r.Use(middleware.RequireAuth(authSvc))`
- Agregar `RequirePermission` en grupos de endpoints según permiso
- Actualizar firma de `NewRouter` para recibir `*svc.AuthService`
- **AC:** `NewRouter(handler, authSvc)` compila; endpoints de lectura requieren `PermissionGoalRead`; escritura require `PermissionGoalWrite`

### T4.2 Reemplazar AuthPlaceholder en evaluation/routes.go

- Mismo patrón que T4.1
- Endpoints de RH evaluation pueden requerir `RequirePermission(auth.PermissionEvaluationWrite)`
- **AC:** Compila; permisos diferenciados por endpoint

### T4.3 Reemplazar AuthPlaceholder en competency/routes.go

- Mismo patrón
- Endpoints de solo lectura (GET) sin permiso extra (basta `RequireAuth`); escritura con `RequirePermission`
- **AC:** Compila

### T4.4 Reemplazar AuthPlaceholder en cycle/routes.go

- Mismo patrón
- **AC:** Compila

### T4.5 Reemplazar AuthPlaceholder en org/routes.go

- Mismo patrón
- **AC:** Compila

### T4.6 Actualizar main.go con inyección de dependencias

- En `api/cmd/main.go` (o el entry point), pasar `authSvc` a cada `NewRouter(...)` 
- **AC:** `go build ./...` compila sin errores; `AuthPlaceholder` ya no se referencia en routers

---

## PR5 — Tests y verificación

### T5.1 Tests de client.ts

- Crear `web/src/lib/api/client.test.ts`
- Testear que interceptor 401 invoca `window.location.href = '/login'`
- Testear que `baseURL` usa `VITE_API_URL` cuando está seteado
- Testear que `credentials: 'include'` se aplica en cada request
- **AC:** Tests pasan con Vitest

### T5.2 Tests de goalsStore (dev fallback + API load)

- Crear `web/src/lib/stores/__tests__/goalsStore.test.ts`
- Testear que en `import.meta.env.DEV` sin `VITE_USE_API`, `load()` carga desde fixture
- Testear que en modo API, `load()` llama al endpoint correcto
- Testear que `addCategory()` llama POST y luego reload
- **AC:** Tests unitarios pasan

### T5.3 Test de integración backend: 401 sin token

- En `api/internal/handler/goal/routes_test.go` (o archivo similar existente)
- Crear test que envía request a endpoint protegido sin token
- Verificar response `401 Unauthorized` con body error JSON
- **AC:** Tests Go pasan con `go test ./...`

---

## Resumen de PRs

| PR | Archivos | Tareas | Depende de |
|---|---|---|---|
| PR1 — Setup | ~8 archivos | T1.1–T1.5 | Ninguna |
| PR2 — Stores | 5 archivos | T2.1–T2.5 | PR1 |
| PR3 — Componentes | ~10 archivos | T3.1–T3.6 | PR2 |
| PR4 — Backend Auth | 6 archivos | T4.1–T4.6 | Ninguna (independiente) |
| PR5 — Tests | 3 archivos | T5.1–T5.3 | PR1 + PR2 + PR4 |

PR4 puede hacerse en paralelo con PR2/PR3 porque no comparte archivos con el frontend.

Orden recomendado de merge: PR1 → PR2 → PR3 (serie), PR4 en paralelo, PR5 al final.
