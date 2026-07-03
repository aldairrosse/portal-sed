# Design: wire-api-replace-mocks

## 1. Arquitectura general

```
web/src/lib/api/
├── client.ts                  # openapi-fetch instance singleton
├── schemas/                   # tipos generados por openapi-typescript
│   ├── goals.d.ts
│   ├── competency.d.ts
│   ├── evaluations.d.ts
│   ├── org-hierarchy.d.ts
│   ├── auth.d.ts
│   └── cycle.d.ts
├── session.svelte.ts          # store runa: sesión del usuario
└── cycle.svelte.ts            # store runa: fase activa del ciclo

web/src/lib/stores/
├── goalsStore.svelte.ts       # migrado a async
├── competencyStore.svelte.ts  # migrado a async
├── evaluationStore.svelte.ts  # migrado a async
├── nineBoxStore.svelte.ts     # migrado a async
├── orgHierarchyStore.svelte.ts # migrado a async
└── devContext.svelte.ts       # intacto (solo dev)
```

### Flujo de datos

```
Component (Svelte) → store.load() → [DEV? fixture : client.GET()]
                                         ↓
                                    openapi-fetch
                                         ↓
                                    Backend API
                                         ↓
                              store.data = response
                              store.loading = false
```

## 2. OpenAPI type generation

### Dependencias

```json
// web/package.json
{
  "devDependencies": {
    "openapi-typescript": "^7.x",
    "openapi-fetch": "^0.x"
  },
  "scripts": {
    "gen:api": "openapi-typescript"
  }
}
```

### Script de generación

```bash
# pnpm run gen:api — genera tipos desde todos los specs
openapi-typescript ../../api/openapi/goals-api.yaml -o src/lib/api/schemas/goals.d.ts
openapi-typescript ../../api/openapi/competency-framework-api.yaml -o src/lib/api/schemas/competency.d.ts
openapi-typescript ../../api/openapi/evaluations-and-9x9.yaml -o src/lib/api/schemas/evaluations.d.ts
openapi-typescript ../../api/openapi/org-hierarchy.yaml -o src/lib/api/schemas/org-hierarchy.d.ts
openapi-typescript ../../api/openapi/auth.yaml -o src/lib/api/schemas/auth.d.ts
openapi-typescript ../../api/openapi/cycle.yaml -o src/lib/api/schemas/cycle.d.ts
```

Cada spec genera un archivo `.d.ts` con un type `paths` que contiene todas las rutas, métodos, parámetros y schemas de respuesta. Los stores importan `paths` de su spec correspondiente y tipan las llamadas a `client.GET/POST/PUT/DELETE`.

### Integración con SvelteKit

El `tsconfig.json` de `web/` ya resuelve `$lib` → `src/lib/`. Los tipos generados se importan como:

```ts
import type { paths } from '$lib/api/schemas/goals';
```

No se requiere configuración adicional de aliases.

## 3. Client HTTP (`client.ts`)

```ts
import createClient from 'openapi-fetch';
import type { paths } from '$lib/api/schemas/goals';

export const apiClient = createClient<paths>({
  baseURL: import.meta.env.VITE_API_URL ?? '/api/v1',
  headers: { 'Content-Type': 'application/json' },
});

apiClient.use({
  onRequest({ request }) {
    request.credentials = 'include';
    return request;
  },
  onResponse({ response }) {
    if (response.status === 401) {
      window.location.href = '/login';
    }
    return response;
  },
});
```

### Cliente multi-spec

`openapi-fetch` se instancia con un type de paths a la vez. Para llamar a distintos dominios, se crean wrappers por spec o se usa un client genérico:

```ts
import type { paths as goals } from '$lib/api/schemas/goals';
import type { paths as competency } from '$lib/api/schemas/competency';
// ...

// Opción A: client instance única + type assertion en cada llamada
// Opción B: wrapper por dominio
export const goalsClient = createClient<goals>({ ... });
export const competencyClient = createClient<competency>({ ... });
```

Opción B recomendada: cada store importa su client específico.

## 4. Store migration pattern

### Before (sync)

```ts
let items = $state(structuredClone(fixtureData));

export function getItems() { return items; }
export function addItem(item: Item) { items = [...items, item]; }
```

### After (async)

```ts
let data = $state<Data | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);

export async function load(): Promise<void> {
  loading = true;
  error = null;

  if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
    data = structuredClone(fixtureData);
    loading = false;
    return;
  }

  const { data: result, error: apiError } = await client.GET('/path');
  if (apiError) {
    error = apiError.error.message ?? 'Error desconocido';
    loading = false;
    return;
  }
  data = result;
  loading = false;
}

export function reload(): Promise<void> {
  return load();
}
```

### Reglas de migración

1. **Getters puros** (derivaciones síncronas de `data`) se mantienen igual — solo cambia la fuente de `data`.
   ```ts
   // Antes: getGoals() devolvía $state con clone de fixture
   // Después: getGoals() lee de data (poblado por load())
   export function getGoals(): Goal[] {
     return data?.categories.flatMap(c => c.goals) ?? [];
   }
   ```

2. **Mutaciones locales** se convierten en llamadas API seguidas de `reload()`:
   ```ts
   export async function addCategory(cat: CreateCategoryRequest): Promise<void> {
     if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
       data = { ...data, categories: [...(data?.categories ?? []), { id: crypto.randomUUID(), ...cat, goals: [] }] };
       return;
     }
     const { error: apiError } = await client.POST('/employees/{empId}/categories', {
       params: { path: { empId: getEmployeeId() } },
       body: cat,
     });
     if (apiError) throw new Error(apiError.error.message);
     await reload();
   }
   ```

3. **Validaciones** (como `isAssignmentValid()`) se mantienen como funciones puras que operan sobre `data`.

4. **`getCyclePhase()`** en goalsStore se reemplaza: en DEV usa `devContext.getPhase()`, en producción usa `cycle.activePhase`.

### Manejo de estados en componentes

```svelte
<script lang="ts">
  import { goals, loading, error, load } from '$lib/stores/goalsStore.svelte';
  import PageSkeleton from '$lib/components/ui/PageSkeleton.svelte';
  import ErrorState from '$lib/components/ui/ErrorState.svelte';

  // load() se llama en el root layout o en el componente raíz del módulo
</script>

{#if loading}
  <PageSkeleton rows={4} />
{:else if error}
  <ErrorState {message} on:retry={load} />
{:else if goals}
  {#each goals as goal}
    ...
  {/each}
{/if}
```

## 5. Session y Cycle stores

### session.svelte.ts

```ts
import type { paths } from '$lib/api/schemas/auth';
import { createClient } from 'openapi-fetch';

const authClient = createClient<paths>({
  baseURL: import.meta.env.VITE_API_URL ?? '/api/v1',
  headers: { 'Content-Type': 'application/json' },
});
authClient.use({ onRequest({ request }) { request.credentials = 'include'; return request; } });

let user = $state<AuthUser | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);

export async function ensureSession(): Promise<AuthUser | null> {
  if (user) return user;
  loading = true;
  error = null;

  if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
    user = { employeeId: 'dev-user', email: 'dev@empresa.com', name: 'Usuario Dev', profileId: 'rh', organizationId: 'org-1' };
    loading = false;
    return user;
  }

  const { data: result, error: apiError } = await authClient.GET('/auth/me');
  if (apiError) { error = 'Sesión no disponible'; loading = false; return null; }
  user = result;
  loading = false;
  return user;
}

export function getSession() { return { user, loading, error }; }
```

Se llama `ensureSession()` en el layout raíz (+layout.svelte) antes de renderizar cualquier ruta protegida.

### cycle.svelte.ts

```ts
import type { paths } from '$lib/api/schemas/cycle';

let activePhase = $state<CyclePhase | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);

export async function loadCycle(): Promise<void> {
  loading = true;
  error = null;

  if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
    activePhase = 'inicio-anio'; // valor por defecto
    loading = false;
    return;
  }

  const { data: result, error: apiError } = await cycleClient.GET('/cycles/current');
  if (apiError) { error = apiError.error.message; loading = false; return; }
  activePhase = result.currentPhase;
  loading = false;
}

export function getActivePhase() { return activePhase; }
export function getCycleState() { return { activePhase, loading, error }; }
```

## 6. Login, Logout y Dev Auth Service

### 6.1 Página `/login` — UI SSO limpia

**Flujo PROD (SSO real):**
```
Browser → GET /login → Renderiza spinner "Iniciando sesión con SSO..."
         → Redirige a SSO provider (OIDC/SAML)
         → SSO valida credenciales
         → Callback a /auth/callback con token
         → Backend valida token, crea sesión, retorna cookie httpOnly
         → Frontend redirige a /
```

**Flujo DEV (sin SSO):**
```
Browser → GET /login → Renderiza botón "Acceso demo"
         → POST /auth/dev-login { email: "dev-rh@empresa.com" }
         → Backend crea sesión temporal con roles/permisos preset
         → Retorna cookie httpOnly + AuthUser
         → Frontend redirige a /
```

**Componente `+page.svelte`:**
```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { goto } from '$app/navigation';
  import { session } from '$lib/api/session.svelte';

  let loading = $state(true);
  let error = $state<string | null>(null);

  onMount(async () => {
    // Si ya hay sesión, redirigir a home
    if (session.user) {
      goto('/');
      return;
    }

    // En PROD: redirigir a SSO provider
    if (!import.meta.env.DEV) {
      // Futuro: window.location.href = ssoRedirectURL;
      error = 'SSO no configurado aún';
      loading = false;
      return;
    }

    // En DEV: mostrar botón de demo
    loading = false;
  });

  async function devLogin(email: string) {
    loading = true;
    error = null;
    try {
      await session.devLogin(email);
      goto('/');
    } catch (e) {
      error = e instanceof Error ? e.message : 'Error al iniciar sesión';
      loading = false;
    }
  }
</script>

<div class="login-container">
  {#if loading}
    <div class="spinner">● ● ●</div>
    <p>Iniciando sesión con SSO...</p>
  {:else if error}
    <div class="error">{error}</div>
    <button onclick={() => devLogin('dev-rh@empresa.com')}>Reintentar</button>
  {:else}
    <h1>Portal SED</h1>
    <p>Inicia sesión para continuar</p>

    {#if import.meta.env.DEV}
      <div class="dev-section">
        <p>Acceso demo (solo desarrollo)</p>
        <button onclick={() => devLogin('dev-rh@empresa.com')}>RH</button>
        <button onclick={() => devLogin('dev-jefe@empresa.com')}>Jefe</button>
        <button onclick={() => devLogin('dev-colaborador@empresa.com')}>Colaborador</button>
      </div>
    {/if}
  {/if}
</div>
```

### 6.2 Logout

**Frontend — `session.svelte.ts`:**
```typescript
export async function logout(): Promise<void> {
  try {
    await authClient.POST('/auth/logout');
  } catch {
    // Si el backend no responde, limpiar localmente
  }
  user = null;
  window.location.href = '/login';
}
```

**Frontend — `Sidebar.svelte`:**
```svelte
<button onclick={logout} class="btn btn-ghost btn-sm">
  Cerrar sesión
</button>
```

**Backend — handler de logout:**
```go
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    sessionID := extractSessionID(r)

    // 1. Eliminar sesión de la DB
    h.sessionStore.Delete(r.Context(), sessionID)

    // 2. Limpiar cookie
    clearSessionCookie(w)

    // 3. Si SSO configurado, retornar end_session_url (futuro)
    // Por ahora: siempre redirigir a /login
    respondJSON(w, http.StatusOK, LogoutResponse{Redirect: "/login"})
}
```

### 6.3 Dev Auth Service

**Propósito:** Permitir desarrollo completo del frontend sin SSO real. Usuarios preset con roles y permisos que pasan todas las validaciones.

**Usuarios preset:**

| Email | Nombre | Perfil | Roles | Permisos |
|-------|--------|--------|-------|----------|
| `dev-rh@empresa.com` | Frankil Perez | rh | admin, evaluator | goal:read, goal:write, evaluation:read, evaluation:write, competency:read |
| `dev-jefe@empresa.com` | Juan Carlos | jefe | manager, evaluator | goal:read, goal:write, evaluation:read, evaluation:write |
| `dev-colaborador@empresa.com` | Maria Lopez | colaborador | collaborator | goal:read, evaluation:read |

**Backend — `api/internal/auth/dev/service.go`:**
```go
package dev

import (
    "context"
    "time"
)

type DevUser struct {
    EmployeeID    string
    Email         string
    Name          string
    Profile       string
    OrganizationID string
    Roles         []string
    Permissions   []string
}

var presetUsers = map[string]*DevUser{
    "dev-rh@empresa.com": {
        EmployeeID: "emp-001",
        Email:      "dev-rh@empresa.com",
        Name:       "Frankil Perez",
        Profile:    "rh",
        OrganizationID: "org-1",
        Roles:      []string{"admin", "evaluator"},
        Permissions: []string{"goal:read", "goal:write", "evaluation:read", "evaluation:write", "competency:read"},
    },
    "dev-jefe@empresa.com": {
        EmployeeID: "emp-002",
        Email:      "dev-jefe@empresa.com",
        Name:       "Juan Carlos",
        Profile:    "jefe",
        OrganizationID: "org-1",
        Roles:      []string{"manager", "evaluator"},
        Permissions: []string{"goal:read", "goal:write", "evaluation:read", "evaluation:write"},
    },
    "dev-colaborador@empresa.com": {
        EmployeeID: "emp-003",
        Email:      "dev-colaborador@empresa.com",
        Name:       "Maria Lopez",
        Profile:    "colaborador",
        OrganizationID: "org-1",
        Roles:      []string{"collaborator"},
        Permissions: []string{"goal:read", "evaluation:read"},
    },
}

func GetUserByEmail(email string) *DevUser {
    return presetUsers[email]
}

func CreateDevSession(ctx context.Context, user *DevUser) (*Session, error) {
    // Crear sesión temporal con expiración de 8 horas
    session := &Session{
        ID:        generateSessionID(),
        UserID:    user.EmployeeID,
        ExpiresAt: time.Now().Add(8 * time.Hour),
    }
    // Guardar en DB o en memoria para DEV
    return session, nil
}
```

**Endpoint — `POST /auth/dev-login`:**
```go
func (h *AuthHandler) DevLogin(w http.ResponseWriter, r *http.Request) {
    // Solo en modo desarrollo
    if os.Getenv("ENV") != "development" {
        respondError(w, http.StatusForbidden, "Dev login not available in production")
        return
    }

    var req struct {
        Email string `json:"email"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request")
        return
    }

    user := dev.GetUserByEmail(req.Email)
    if user == nil {
        respondError(w, http.StatusUnauthorized, "Unknown dev user")
        return
    }

    session, err := dev.CreateDevSession(r.Context(), user)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to create session")
        return
    }

    // Setear cookie httpOnly
    setSessionCookie(w, session.ID)

    respondJSON(w, http.StatusOK, AuthResponse{
        User: user,
        SessionID: session.ID,
    })
}
```

### 6.4 Interfaz SSOAdapter (preparación para futuro)

```go
// api/internal/auth/sso/adapter.go

type SSOAdapter interface {
    // ValidateToken verifica un token externo (OIDC/SAML)
    ValidateToken(ctx context.Context, token string) (*SSOUser, error)

    // GetEndSessionURL retorna la URL de logout del proveedor SSO
    GetEndSessionURL(ctx context.Context, sessionID string) (string, error)

    // GetUserFromToken extrae info del usuario desde el token SSO
    GetUserFromToken(ctx context.Context, token string) (*SSOUser, error)
}

type SSOUser struct {
    ExternalID    string
    Email         string
    Name          string
    Roles         []string
    OrganizationID string
}

// Implementaciones futuras:
// - OIDCAdapter (Keycloak, Auth0, Azure AD)
// - SAMLAdapter
// - LDAPAdapter
```

**Estado actual:** `SSOAdapter` es `nil` en main.go. Cuando se conecte SSO real:
1. Implementar `OIDCAdapter` o `SAMLAdapter`
2. Inyectar en `AuthService` o `AuthHandler`
3. Actualizar login para redirigir a SSO
4. Actualizar logout para llamar `GetEndSessionURL`

## 7. Conexión RequireAuth en backend

### Estado actual

| Router | Middleware actual |
|---|---|
| `api/internal/handler/goal/routes.go` | `AuthPlaceholder` |
| `api/internal/handler/evaluation/routes.go` | `AuthPlaceholder` |
| `api/internal/handler/competency/routes.go` | `AuthPlaceholder` |
| `api/internal/handler/cycle/routes.go` | `AuthPlaceholder` |
| `api/internal/handler/org/routes.go` | `AuthPlaceholder` |

### Cambio

Cada router recibe `*svc.AuthService` como dependencia y reemplaza:

```go
// Antes
r.Use(middleware.AuthPlaceholder)

// Después
r.Use(middleware.RequireAuth(authSvc))
```

Y en handlers individuales se agrega `RequirePermission` donde corresponde:

```go
r.Group(func(r chi.Router) {
  r.Use(middleware.RequirePermission(auth.PermissionEvaluationWrite))
  r.Post("/evaluations/{id}/finalize", handler.FinalizeEvaluation)
})
```

Los permisos se definen según la documentación de C7 (identity-access):

| Handler | Endpoints de solo lectura | Endpoints de escritura |
|---|---|---|
| goal | `PermissionGoalRead` | `PermissionGoalWrite` |
| evaluation | `PermissionEvaluationRead` | `PermissionEvaluationWrite` |
| competency | `PermissionCompetencyRead` | `PermissionCompetencyWrite` |
| cycle | `PermissionCycleRead` | `PermissionCycleWrite` |
| org | `PermissionOrgRead` | `PermissionOrgWrite` |

### Impacto en constructores

Los `NewRouter` en cada handler necesitan `authSvc` como parámetro. Ejemplo:

```go
// Antes
func NewRouter(handler *GoalHandler) chi.Router

// Después
func NewRouter(handler *GoalHandler, authSvc *svc.AuthService) chi.Router
```

Esto propaga hacia `main.go` donde se inyectan las dependencias.

## 8. Vite config

No se requieren cambios en `vite.config.ts`. El alias `$lib` ya existe por defecto en SvelteKit. Solo se necesita agregar `VITE_API_URL` al `.env`:

```
# web/.env
VITE_API_URL=http://localhost:8080/api/v1
VITE_USE_API=false
```

## 9. Plan de cambios por archivo

### PR1 — Setup: client + types + session/cycle + login/logout (aprox. 10 archivos)

| Archivo | Acción |
|---|---|
| `web/package.json` | Add `openapi-typescript`, `openapi-fetch` deps |
| `web/package.json` | Add `"gen:api"` script |
| `web/src/lib/api/client.ts` | Create — `openapi-fetch` instance con interceptors |
| `web/src/lib/api/schemas/` | Create — types generados (6 .d.ts files) |
| `web/src/lib/api/session.svelte.ts` | Create — `ensureSession()`, `getSession()` |
| `web/src/lib/api/cycle.svelte.ts` | Create — `loadCycle()`, `getActivePhase()` |
| `web/.env` | Create — `VITE_API_URL`, `VITE_USE_API` |
| `web/src/routes/+layout.svelte` | Update — call `ensureSession()` en load |
| `web/src/routes/login/+page.svelte` | Create — UI SSO limpia con botón demo en DEV |
| `web/src/lib/components/Sidebar.svelte` | Update — agregar botón logout |

### PR2 — Stores migration (5 stores × ~15 funciones = 5 archivos)

| Archivo | Acción |
|---|---|
| `web/src/lib/stores/goalsStore.svelte.ts` | Rewrite — async load/reload, mutations via API |
| `web/src/lib/stores/competencyStore.svelte.ts` | Rewrite — async load/reload, mutations via API |
| `web/src/lib/stores/evaluationStore.svelte.ts` | Rewrite — async load/reload, mutations via API |
| `web/src/lib/stores/nineBoxStore.svelte.ts` | Rewrite — async load/reload, mutations via API |
| `web/src/lib/stores/orgHierarchyStore.svelte.ts` | Rewrite — async load/reload (read-only tree) |

### PR3 — Componentes: loading/error states (~10 archivos)

| Archivo | Acción |
|---|---|
| `web/src/lib/components/evaluation/EmployeeEvaluationTable.svelte` | Update — wrap con loading/error |
| `web/src/lib/components/evaluation/CompetencyRatingCard.svelte` | Update — wrap con loading/error |
| `web/src/lib/components/evaluation/EmployeeEvaluationDetail.svelte` | Update — wrap con loading/error |
| `web/src/lib/components/evaluation/GoalClosureCard.svelte` | Update — wrap con loading/error |
| `web/src/lib/components/competency/*.svelte` | Update — wrap con loading/error |
| `web/src/lib/components/goals/*.svelte` | Update — wrap con loading/error |
| `web/src/lib/components/nine-box/*.svelte` | Update — wrap con loading/error |
| `web/src/lib/components/org-hierarchy/*.svelte` | Update — wrap con loading/error |
| Web deps prop across all components | Update — replace `getPhase()` de devContext por `getActivePhase()` de cycle |

### PR4 — Backend: RequireAuth wiring + dev auth service (~8 archivos)

| Archivo | Acción |
|---|---|
| `api/internal/handler/goal/routes.go` | Update — `AuthPlaceholder` → `RequireAuth(authSvc)` + perms |
| `api/internal/handler/evaluation/routes.go` | Update — `AuthPlaceholder` → `RequireAuth(authSvc)` + perms |
| `api/internal/handler/competency/routes.go` | Update — `AuthPlaceholder` → `RequireAuth(authSvc)` + perms |
| `api/internal/handler/cycle/routes.go` | Update — `AuthPlaceholder` → `RequireAuth(authSvc)` + perms |
| `api/internal/handler/org/routes.go` | Update — `AuthPlaceholder` → `RequireAuth(authSvc)` + perms |
| `api/cmd/main.go` | Update — pasar `authSvc` a cada `NewRouter(...)` |
| `api/internal/auth/dev/service.go` | Create — dev auth service con usuarios preset |
| `api/internal/auth/sso/adapter.go` | Create — interfaz SSOAdapter para futuro |
| `api/internal/middleware/auth.go` | Cleanup — eliminar `AuthPlaceholder` (opcional, mantener si hay otros usos) |

### PR5 — Tests y verificación (~3 archivos)

| Archivo | Acción |
|---|---|
| `web/src/lib/api/client.test.ts` | Create — test interceptor 401, base URL config |
| `web/src/lib/stores/__tests__/goalsStore.test.ts` | Create — test dev fallback, API load, error state |
| `api/internal/handler/goal/routes_test.go` | Update — test 401 sin token |
