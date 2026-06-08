# Spec: wire-api-replace-mocks

## Propósito

Reemplazar los datos sincrónicos de fixture en los stores de Svelte 5 con llamadas asíncronas a las APIs reales del backend usando `openapi-fetch` + tipos generados desde `api/openapi/*.yaml`. Los fixtures se conservan como fallback exclusivo en desarrollo (`import.meta.env.DEV`) para trabajo offline.

## Alcance

### Requerimientos funcionales

| ID | Requerimiento | Prioridad |
|---|---|---|
| F1 | Cliente HTTP tipado que conoce la estructura de cada endpoint desde OpenAPI | Alta |
| F2 | Los 5 stores existentes (goals, nineBox, competency, evaluation, orgHierarchy) migran a async con triplete `data`/`loading`/`error` | Alta |
| F3 | En `import.meta.env.DEV` sin `VITE_USE_API=true`, stores cargan desde fixtures (mismo comportamiento actual) | Alta |
| F4 | En producción (o `VITE_USE_API=true`), stores fetch desde backend real | Alta |
| F5 | Store `session.svelte.ts` expone usuario autenticado desde `GET /auth/me` con `ensureSession()` | Alta |
| F6 | Store `cycle.svelte.ts` expone fase activa desde `GET /cycle/current` | Alta |
| F7 | Interceptor 401 redirige a `/login` automáticamente | Alta |
| F8 | Componentes existentes muestran `<PageSkeleton />` cuando `loading === true` | Media |
| F9 | Componentes existentes muestran `<ErrorState />` cuando `error !== null`, con botón reintentar que llama `reload()` | Media |
| F10 | Backend: reemplazar `AuthPlaceholder` por `RequireAuth`/`RequirePermission` en los 5 routers handler | Alta |
| F11 | Configuración de `pnpm run gen:api` que genera tipos desde los OpenAPI specs | Alta |

### No alcance

- UI de login (change futuro)
- Caching avanzado (SWR, React Query style) — primera iteración es fetch directo
- Tests E2E (change separado)
- Modificar handlers del backend ni sus specs OpenAPI
- Eliminar fixtures ni `devContext.svelte.ts`

## Escenarios

### E1: Fetch exitoso (200)

```
Given el backend responde 200 con body válido
When goalsStore.load() se ejecuta
Then data contiene el array de categorías con goals anidados
And loading = false
And error = null
```

### E2: Error de red / servidor caído

```
Given el backend no responde (network error / 5xx)
When goalsStore.load() se ejecuta
Then error contiene el mensaje del error
And data = null
And loading = false
And el componente muestra <ErrorState /> con botón reintentar
When usuario hace clic en reintentar
Then goalsStore.load() se ejecuta de nuevo
```

### E3: Sesión expirada (401)

```
Given el token de sesión expiró
When cualquier store hace una llamada API
Then client interceptor captura 401
And window.location.href = '/login'
```

### E4: Dev mode fixture fallback

```
Given import.meta.env.DEV = true
And VITE_USE_API no está seteado o es false
When goalsStore.load() se ejecuta
Then data = structuredClone(fixtureData)
And loading = false
And error = null
And NO se hace fetch al backend
```

### E5: Dev mode con VITE_USE_API=true

```
Given import.meta.env.DEV = true
And VITE_USE_API = true
When goalsStore.load() se ejecuta
Then se hace fetch al backend real (como en producción)
And se respetan loading/error states
```

### E6: Mutación con reload posterior

```
Given goalsStore.load() completó exitosamente
When goalsStore.addCategory(data) se ejecuta
Then POST /employees/{empId}/categories se llama con el body correcto
And goalsStore.load() se ejecuta tras éxito para refrescar
```

### E7: Auth middleware reemplazado en backend

```
Given un request sin token a PATCH /goals/{goalId}/progress
When el request llega al router
Then RequireAuth retorna 401
And NO se ejecuta el handler
```

## Contratos de datos

### session.svelte.ts

```ts
interface SessionState {
  user: AuthUser | null;     // GET /auth/me response
  loading: boolean;          // true mientras se resuelve la sesión
  error: string | null;      // mensaje de error si falló
}

interface AuthUser {
  employeeId: string;        // uuid
  email: string;
  name: string;
  profileId: string;         // EvaluationProfile
  organizationId: string;    // uuid
}
```

### cycle.svelte.ts

```ts
interface CycleState {
  activePhase: CyclePhase | null;  // 'inicio-anio' | 'medio-anio' | 'fin-anio'
  loading: boolean;
  error: string | null;
}
```

### Store pattern (aplica a los 5 stores)

```ts
interface AsyncStoreState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
}

interface AsyncStore<T> extends AsyncStoreState<T> {
  load(): Promise<void>;
  reload(): Promise<void>;    // alias que resetea error + loading antes de fetch
}
```

### goalsStore.svelte.ts — API surface

| Función | Reemplazo API | Fixture |
|---|---|---|
| `load()` | `GET /api/v1/employees/{empId}/categories` | `fixtures/goals/goal-categories.json` |
| `getCategories()` | `data` (state interno) | — |
| `getGoals()` | `data.categories[].goals` (anidado) | `fixtures/goals/goals.json` |
| `getKpis()` | `GET /api/v1/kpis` | `fixtures/goals/kpis.json` |
| `addCategory()` | `POST /api/v1/employees/{empId}/categories` → `reload()` | mutación local |
| `updateCategory()` | `PUT /api/v1/employees/{empId}/categories/{catId}` → `reload()` | mutación local |
| `deleteCategory()` | `DELETE /api/v1/employees/{empId}/categories/{catId}` → `reload()` | mutación local |
| `addGoal()` | `POST /api/v1/employees/{empId}/categories/{catId}/goals` → `reload()` | mutación local |
| `updateGoal()` | `PUT /api/v1/goals/{goalId}` → `reload()` | mutación local |
| `deleteGoal()` | `DELETE /api/v1/goals/{goalId}` → `reload()` | mutación local |
| `updateGoalProgress()` | `PATCH /api/v1/goals/{goalId}/progress` → `reload()` | mutación local |
| `linkKpiToGoal()` | `POST /api/v1/goals/{goalId}/kpis` → `reload()` | mutación local |
| `unlinkKpiFromGoal()` | `DELETE /api/v1/goals/{goalId}/kpis/{kpiId}` → `reload()` | mutación local |
| `getAssignment()` | `GET /api/v1/employees/{empId}/assignments` | fixture |
| `addAssignment()` | `POST /api/v1/employees/{empId}/assignments` → `reload()` | mutación local |

### competencyStore.svelte.ts — API surface

| Función | Reemplazo API | Fixture |
|---|---|---|
| `load()` | `GET /api/v1/pillars`, `GET /api/v1/competencies?include=scale-criteria`, `GET /api/v1/levels`, `GET /api/v1/acceptance-levels` | `fixtures/competency/*` |
| `getPillars()` | `data.pillars` | — |
| `getCompetencies()` | `data.competencies` | — |
| `getLevelDefinitions()` | `data.levelDefinitions` | — |
| `getCompetencyAcceptanceLevels()` | `data.acceptanceLevels` | — |
| Mutaciones (addPillar, etc.) | POST/PUT/DELETE → `reload()` | — |

### evaluationStore.svelte.ts — API surface

| Función | Reemplazo API | Fixture |
|---|---|---|
| `load()` | `GET /api/v1/evaluations`, `GET /api/v1/evaluations/summary` | `fixtures/evaluations/*` |
| `getCompetencyRatings()` | `data.competencyRatings` | — |
| `submitSelfEvaluation()` | `POST /api/v1/evaluations/{id}/self-evaluation` → `reload()` | — |
| `submitRHEvaluation()` | `POST /api/v1/evaluations/{id}/rh-evaluation` → `reload()` | — |
| `finalizeEvaluation()` | `POST /api/v1/evaluations/{id}/finalize` → `reload()` | — |

### nineBoxStore.svelte.ts — API surface

| Función | Reemplazo API | Fixture |
|---|---|---|
| `load()` | `GET /api/v1/nine-box/matrices`, `GET /api/v1/nine-box/scales`, `GET /api/v1/nine-box/quadrants` | `fixtures/nine-box/*` |
| `getEntries()` | `data.entries` | — |
| `getScales()` | `data.scales` | — |
| `createMatrix()` | `POST /api/v1/nine-box/matrices` → `reload()` | — |
| `upsertEntry()` | `POST /api/v1/nine-box/matrices/{id}/entries` → `reload()` | — |

### orgHierarchyStore.svelte.ts — API surface

| Función | Reemplazo API | Fixture |
|---|---|---|
| `load()` | `GET /api/v1/org-tree` | `fixtures/org-hierarchy/org-tree.json` |
| `getTree()` | `data` | — |
| `getNodeById()` | `data` (traversal local) | — |
| `getDescendants()` | `data` (traversal local) | — |

### client.ts — contratos HTTP

```ts
interface ClientConfig {
  baseURL: string;                           // VITE_API_URL ?? '/api/v1'
  credentials: 'include';                    // envía cookie httpOnly
  headers: { 'Content-Type': 'application/json' };
}

// Interceptor onResponse:
//   401 → window.location.href = '/login'
//   other error → propaga para que stores capturen
```

## Criterios de éxito

- [ ] `pnpm run gen:api` genera tipos TypeScript sin errores desde los 6 OpenAPI specs
- [ ] `client.ts` configurado con `baseURL`, `credentials: 'include'`, interceptor 401
- [ ] `pnpm run check` pasa sin errores de tipo
- [ ] `go build ./...` en `api/` compila sin errores tras conectar auth middleware
- [ ] Los 5 stores expone `data`/`loading`/`error` como `$state`
- [ ] En DEV sin `VITE_USE_API`, todos los stores cargan desde fixtures
- [ ] En producción, todos los stores fetch desde backend real
- [ ] Componentes consumidores renderizan `<PageSkeleton />` durante carga
- [ ] Componentes consumidores renderizan `<ErrorState />` en errores con reintentar
- [ ] 401 redirect funciona en todos los requests del client
- [ ] `AuthPlaceholder` eliminado de los 5 routers; reemplazado por `RequireAuth` + `RequirePermission`
