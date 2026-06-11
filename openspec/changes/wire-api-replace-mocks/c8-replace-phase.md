# C8 — Replace Phase: Auth, SSO, Fallback Logout y Riesgos

> Este documento actualiza la fase de reemplazo de C8 (`wire-api-replace-mocks`) con la arquitectura de autenticación, integración SSO, ruta fallback de logout, y análisis de riesgos al migrar de fixtures a API real.

---

## 1. Estado actual del frontend

| Aspecto | Estado | Archivo clave |
|---------|--------|---------------|
| Auth real | ❌ No existe — `FIXTURE_USER` hardcodeado | `api/session.svelte.ts` |
| Login page | ❌ No existe `/login` | — |
| Logout | ❌ Schema definido, nunca llamado | `api/schemas/auth.d.ts` |
| Route guards | ❌ Sin protección — cualquier URL accesible | `routes/+layout.ts` |
| Mock data | 19 archivos JSON, 100% de la app | `fixtures/` |
| Stores con fixtures | 5 stores con CRUD síncrono sobre arrays | `stores/*Store.svelte.ts` |
| DevToolbar | ✅ Funcional — perfil + fase, `Ctrl+Shift+D` | `components/DevToolbar.svelte` |
| RBAC/permisos | Solo filtrado visual en sidebar | `nav/menuConfig.ts` |
| ForbiddenState | Existe pero nadie lo importa | `ui/ForbiddenState.svelte` |

### Archivos de fixtures

```
fixtures/
├── activity/activity-logs.json
├── competency/acceptance-levels.json, competencies.json, competency-acceptance-levels.json,
│              pillars.json, scale-criteria.json
├── evaluations/goal-closures.json, rh-evaluations.json, self-evaluations.json
├── goals/assignments.json, cycle.json, goal-categories.json, goal-kpi-links.json,
│         goals.json, kpis.json
├── nine-box/matrix-entries.json, quadrant-definitions.json, scale-definitions.json
└── org-hierarchy/org-tree.json
```

---

## 2. Arquitectura de Auth propuesta

### 2.1 Flujo general

```
┌─────────────────────────────────────────────────────────────┐
│                    FRONTEND (Svelte 5)                       │
│                                                             │
│  ┌────────────┐   ┌───────────────┐   ┌─────────────────┐  │
│  │ DevToolbar  │   │  AuthGuard    │   │  Sidebar         │  │
│  │ (DEV only)  │   │  (+layout)    │   │  (perfil-filter) │  │
│  └──────┬─────┘   └──────┬────────┘   └─────────────────┘  │
│         │                │                                  │
│         ▼                ▼                                  │
│  ┌──────────────────────────────────┐                       │
│  │       session.svelte.ts          │                       │
│  │  - ensureSession()               │                       │
│  │  - user: $state<AuthUser>        │                       │
│  │  - logout() → POST /auth/logout  │                       │
│  │  - refreshToken()                │                       │
│  └──────────┬───────────────────────┘                       │
│             │                                               │
│      ┌──────┴──────┐                                        │
│      ▼             ▼                                        │
│  DEV: fixture   PROD: API call                              │
│  (devContext)   (session/cookie)                             │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│                    BACKEND (Go + Chi)                        │
│                                                             │
│  ┌───────────┐  ┌─────────────┐  ┌──────────────────────┐  │
│  │ POST      │  │ POST        │  │ GET                   │  │
│  │ /auth/    │  │ /auth/      │  │ /auth/me              │  │
│  │ login     │  │ logout      │  │ (validate session)    │  │
│  └─────┬─────┘  └─────────────┘  └──────────────────────┘  │
│        │                                                     │
│        ▼                                                     │
│  ┌───────────────────────────────────┐                       │
│  │  Session Store (DB)               │                       │
│  │  - session_id (httpOnly cookie)   │                       │
│  │  - employee_id                    │                       │
│  │  - expires_at                     │                       │
│  │  - refresh_token                  │                       │
│  └───────────────────────────────────┘                       │
│                                                             │
│  ┌───────────────────────────────────┐  ← FUTURO SSO        │
│  │  SSO Adapter (pluggable)          │                       │
│  │  - OIDC / SAML / LDAP             │                       │
│  │  - Token validation               │                       │
│  │  - User info extraction           │                       │
│  └───────────────────────────────────┘                       │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Definición de `AuthUser`

```typescript
// web/src/lib/api/session.svelte.ts

interface AuthUser {
  employeeId: string;
  email: string;
  name: string;
  profile: string;       // 'colaborador' | 'jefe' | 'rh' | ...
  organizationId: string;
  roles: string[];       // ['evaluator', 'manager'] — para RBAC local
  permissions: string[]; // ['goal:read', 'evaluation:write'] — granular
}
```

### 2.3 Validación de token/sesión desde el servicio

```
Request → Backend
  │
  ├─ 1. Extraer cookie httpOnly (session_id)
  │
  ├─ 2. Buscar sesión en DB:
  │     SELECT * FROM sessions
  │     WHERE id = $1 AND expires_at > NOW()
  │
  ├─ 3. Si existe → inyectar employee_id + roles en context
  │     ctx = { employeeId, roles, permissions }
  │
  ├─ 4. Si no existe → 401 Unauthorized
  │
  └─ 5. RequirePermission verifica ctx.permissions vs endpoint requerido
```

**El frontend NO valida tokens** — solo adjunta la cookie (`credentials: 'include'`). Toda la validación es server-side.

---

## 3. Ruta fallback de logout con SSO

### 3.1 Flujo

```
Frontend                     Backend                      SSO
   │                            │                           │
   │  POST /auth/logout         │                           │
   │  { withSSO: true }         │                           │
   │  ──────────────────────►   │                           │
   │                            │  1. Eliminar sesión DB    │
   │                            │  2. Invalidar cookie      │
   │                            │                           │
   │                            │  ¿SSO configurado?        │
   │                            │  ├─ Sí: OIDC end_session  │
   │                            │  │  ──────────────────►   │
   │                            │  │  ◄──────────────────   │
   │                            │  │  { end_session_url }   │
   │                            │  └─ No: { redirect: /login│
   │                            │                           │
   │  { redirect: "..." }       │                           │
   │  ◄──────────────────────   │                           │
   │                            │                           │
   │  window.location = redirect│                           │
   │  (SSO page or /login)      │                           │
```

### 3.2 Backend: handler de logout

```go
// api/internal/handler/auth/handler.go

type LogoutRequest struct {
    WithSSO bool `json:"withSSO,omitempty"`
}

type LogoutResponse struct {
    Redirect string `json:"redirect"` // URL a donde redirigir
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    sessionID := extractSessionID(r) // from httpOnly cookie

    // 1. Eliminar sesión de la DB
    if err := h.sessionStore.Delete(r.Context(), sessionID); err != nil {
        // Log but don't fail — session might already be expired
    }

    // 2. Limpiar cookie
    clearSessionCookie(w)

    // 3. Si SSO está configurado, retornar end_session_url
    if h.ssoAdapter != nil && extractSSOFlag(r) {
        endSessionURL, err := h.ssoAdapter.GetEndSessionURL(r.Context(), sessionID)
        if err == nil {
            respondJSON(w, http.StatusOK, LogoutResponse{Redirect: endSessionURL})
            return
        }
        // Fallback: if SSO fails, redirect to local login
    }

    // 4. Sin SSO o fallback
    respondJSON(w, http.StatusOK, LogoutResponse{Redirect: "/login"})
}
```

### 3.3 Frontend: función de logout

```typescript
// web/src/lib/api/session.svelte.ts

export async function logout(): Promise<void> {
  try {
    const { data } = await authClient.POST('/auth/logout', {
      body: { withSSO: true },
    });

    // Limpiar estado local
    user = null;
    loading = false;
    error = null;

    // Redirigir (SSO page o /login)
    if (data?.redirect) {
      window.location.href = data.redirect;
    } else {
      window.location.href = '/login';
    }
  } catch {
    // Si el backend no responde, limpiar localmente
    user = null;
    window.location.href = '/login';
  }
}
```

### 3.4 SSO Adapter (futuro, pluggable)

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

**Por ahora (C8):** No se implementa SSO. El adapter es `nil` y el logout siempre redirige a `/login`. La interfaz queda preparada para futuro.

---

## 4. DevToolbar compatible con API real

### 4.1 Problema

En producción, el perfil viene de `GET /auth/me` (asignado por el backend). En DEV, el DevToolbar permite cambiar el perfil manualmente. Si activamos `VITE_USE_API=true`, el DevToolbar deja de funcionar porque `session.user.profile` viene del backend.

### 4.2 Solución: header de override en DEV

```
DevToolbar (DEV)          Backend
      │                      │
      │  Request with        │
      │  X-Dev-Profile: rh   │
      │  ─────────────────►  │
      │                      │  ¿DEV mode?
      │                      │  ├─ Sí: usar X-Dev-Profile
      │                      │  └─ No: ignorar header
      │                      │
      │  Response con        │
      │  perfil "rh"         │
      │  ◄─────────────────  │
```

### 4.3 Frontend: inyectar header en DEV

```typescript
// web/src/lib/api/client.ts

apiClient.use({
  onRequest({ request }) {
    request.credentials = 'include';

    // DEV only: inyectar perfil override desde DevToolbar
    if (import.meta.env.DEV) {
      const devProfile = sessionStorage.getItem('dev-profile');
      if (devProfile) {
        request.headers.set('X-Dev-Profile', devProfile);
      }
    }

    return request;
  },
});
```

### 4.4 Backend: leer header en DEV

```go
// api/internal/middleware/dev_override.go

func DevProfileOverride(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Solo en modo desarrollo
        if os.Getenv("ENV") != "development" {
            next.ServeHTTP(w, r)
            return
        }

        if profile := r.Header.Get("X-Dev-Profile"); profile != "" {
            // Override el perfil en el context
            ctx := context.WithValue(r.Context(), "dev_profile_override", profile)
            next.ServeHTTP(w, r.WithContext(ctx))
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

### 4.5 DevToolbar se mantiene igual

El componente `DevToolbar.svelte` no cambia. Sigue guardando `{ profile, phase }` en `sessionStorage`. La diferencia es que ahora el perfil se inyecta como header en cada request, y el backend lo usa solo en modo DEV.

---

## 5. Permisos, roles y validaciones secundarias

### 5.1 Carga local de roles/permisos

```
GET /auth/me → { employeeId, profile, roles, permissions }
                    │
                    ▼
            session.svelte.ts
                    │
                    ├─ user.roles = ['manager', 'evaluator']
                    ├─ user.permissions = ['goal:read', 'goal:write', 'evaluation:read']
                    │
                    ▼
            Componentes usan:
            {#if hasPermission('goal:write')}
              <EditButton />
            {/if}
```

### 5.2 Función de verificación de permisos

```typescript
// web/src/lib/api/session.svelte.ts

export function hasPermission(permission: string): boolean {
  return user?.permissions.includes(permission) ?? false;
}

export function hasRole(role: string): boolean {
  return user?.roles.includes(role) ?? false;
}

// Uso en componentes:
import { hasPermission } from '$lib/api/session.svelte';

{#if hasPermission('evaluation:write')}
  <button on:click={finalize}>Finalizar evaluación</button>
{/if}
```

### 5.3 Route guards via +layout.server.ts

```typescript
// web/src/routes/rh/+layout.server.ts

import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ parent }) => {
  const { user } = await parent();

  if (!user) {
    throw redirect(302, '/login');
  }

  if (!user.roles.includes('rh') && !user.roles.includes('admin')) {
    throw redirect(302, '/'); // o a página de "no autorizado"
  }

  return { user };
};
```

### 5.4 Validaciones secundarias (offline/local)

| Validación | Dónde | Cuándo |
|------------|-------|--------|
| Fase del ciclo permite edición | `getGoalPermissions()` en goalsStore | Antes de mutar |
| Owner del objetivo | `isOwner` check en componente | Antes de mostrar botones |
| Límite de peso por categoría | `weightValidationService` | En formulario de asignación |
| Estado de evaluación (draft/submitted) | `evaluationStore` | Antes de submit |
| Perfil mínimo para acción | `hasRole()` en componente | Renderizado condicional |

Estas validaciones **no requieren backend** — se ejecutan localmente con los datos ya cargados. Son la segunda línea de defensa después de los permisos del backend.

---

## 6. Riesgos del reemplazo mock → API real

### 6.1 Riesgos ALTOS

| # | Riesgo | Impacto | Probabilidad | Mitigación |
|---|--------|---------|--------------|------------|
| **H1** | **No hay login page** — `VITE_USE_API=true` activa el interceptor 401 → redirect a `/login` que da 404 | App inutilizable | Alta | Crear `routes/login/+page.svelte` ANTES de activar API |
| **H2** | **Stores mutan fixtures directamente** — cada `addGoal()` hace `goals.push()` sobre arrays importados | Al cambiar a API, cada store necesita reescritura completa del CRUD | Alta | Refactor progresivo: 1 store a la vez, mantener fixture como fallback |
| **H3** | **Sin route guards** — cualquier usuario escribe `/rh/evaluaciones` | Datos expuestos, comportamiento inesperado | Alta | Crear `+layout.server.ts` con load function que valide角色 |
| **H4** | **Sesión no se valida entre navegaciones** — `ensureSession()` se llama una vez, queda en memoria | Cookie expira → usuario ve datos stale o errores 500 | Media | Agregar validación en `+layout.ts` o refresh automático |

### 6.2 Riesgos MEDIOS

| # | Riesgo | Impacto | Probabilidad | Mitigación |
|---|--------|---------|--------------|------------|
| **M1** | **DevToolbar depende de `FIXTURE_USER`** — al activar API, el perfil viene del backend | DevToolbar dejaría de funcionar | Alta | Header `X-Dev-Profile` (sección 4) |
| **M2** | **18 páginas sin validación de roles** — cada una asume acceso | Errores 500 o datos vacíos | Media | Crear `AuthGuard` o `+layout` con load function |
| **M3** | **`ForbiddenState.svelte` existe pero nadie lo usa** | Componente muerto, confusión将来 | Baja | Integrarlo en guards o eliminarlo |
| **M4** | **500+ líneas de lógica en stores** — cada store tiene CRUD completo | Migración costosa, riesgo de regression | Alta | Crear servicios intermedios (`goalService.ts`) que abstraigan API vs fixture |
| **M5** | **Backend retorna estructura diferente al fixture** | Componentes reciben `undefined` en campos esperados | Media | Tipos OpenAPI generados ya validan la estructura; testear contra backend real |

### 6.3 Riesgos BAJOS

| # | Riesgo | Impacto | Probabilidad | Mitigación |
|---|--------|---------|--------------|------------|
| **L1** | **OpenAPI types ya generados** — los 6 schemas `.d.ts` están listos | Ya hay contrato tipado | — | Solo consumir |
| **L2** | **`client.ts` ya maneja 401** — redirige a `/login` | El mecanismo base existe | — | Completar con login page |
| **L3** | **Fixtures se conservan como fallback** | No hay pérdida de funcionalidad DEV | — | Solo documentar propósito |
| **L4** | **PR4 (backend auth) es independiente** — puede hacerse en paralelo | No bloquea frontend | — | Merge order: PR1→PR2→PR3, PR4 en paralelo |

### 6.4 Matriz de riesgo vs esfuerzo

```
                    ESFUERZO
              Bajo        Medio       Alto
         ┌───────────┬───────────┬───────────┐
    Alta │           │  H2, M4   │  H1       │
R        │           │           │           │
I   Med  │  M3, L2   │  H3, H4   │  M1       │
E        │           │  M2, M5   │           │
S   Baja │  L1, L3   │           │           │
G        │           │           │           │
O        └───────────┴───────────┴───────────┘
```

---

## 7. Orden de implementación (actualizado)

### Fase 1: Auth básica (sin romper nada)

| Tarea | Archivos | Dependencia |
|-------|----------|-------------|
| 1.1 Crear `/login` page | `routes/login/+page.svelte`, `routes/login/+page.server.ts` | Ninguna |
| 1.2 Implementar `POST /auth/login` backend | `handler/auth/handler.go` | C7 (ya existe) |
| 1.3 Agregar `+layout.server.ts` raíz | `routes/+layout.server.ts` | 1.1 |
| 1.4 Crear componente `AuthGuard` | `components/AuthGuard.svelte` | 1.3 |
| 1.5 Implementar logout | `session.svelte.ts` (logout fn), `Sidebar.svelte` (botón) | 1.1 |

### Fase 2: DevToolbar compatible con API

| Tarea | Archivos | Dependencia |
|-------|----------|-------------|
| 2.1 Inyectar header `X-Dev-Profile` en client.ts | `api/client.ts` | — |
| 2.2 Crear middleware `DevProfileOverride` en backend | `middleware/dev_override.go` | — |
| 2.3 Actualizar DevToolbar para guardar perfil en sessionStorage | `components/DevToolbar.svelte` | — |

### Fase 3: Migración progresiva de mocks a API

| Tarea | Archivos | Dependencia |
|-------|----------|-------------|
| 3.1 Migrar `orgHierarchyStore` (más simple, solo lectura) | `stores/orgHierarchyStore.svelte.ts` | Fase 1 |
| 3.2 Migrar `competencyStore` | `stores/competencyStore.svelte.ts` | Fase 1 |
| 3.3 Migrar `goalsStore` (CRUD completo) | `stores/goalsStore.svelte.ts` | Fase 1 |
| 3.4 Migrar `evaluationStore` | `stores/evaluationStore.svelte.ts` | Fase 1 |
| 3.5 Migrar `nineBoxStore` | `stores/nineBoxStore.svelte.ts` | Fase 1 |

### Fase 4: Route guards + RBAC

| Tarea | Archivos | Dependencia |
|-------|----------|-------------|
| 4.1 Crear `+layout.server.ts` por módulo (rh, evaluacion, etc.) | `routes/rh/+layout.server.ts`, etc. | Fase 1 |
| 4.2 Integrar `ForbiddenState.svelte` en guards | `ui/ForbiddenState.svelte` | 4.1 |
| 4.3 Agregar `hasPermission()` checks en componentes | Componentes afectados | Fase 3 |

### Fase 5: SSO (futuro)

| Tarea | Archivos | Dependencia |
|-------|----------|-------------|
| 5.1 Definir interfaz `SSOAdapter` | `auth/sso/adapter.go` | — |
| 5.2 Implementar `OIDCAdapter` | `auth/sso/oidc.go` | 5.1 |
| 5.3 Conectar adapter en `main.go` | `cmd/server/main.go` | 5.2 |
| 5.4 Actualizar logout para SSO | `handler/auth/handler.go` | 5.2 |

---

## 8. Criterios de éxito

- [ ] Login page funcional (`/login` renderiza sin 404)
- [ ] `POST /auth/login` retorna cookie httpOnly válida
- [ ] `POST /auth/logout` limpia sesión y retorna redirect
- [ ] DevToolbar funciona con `VITE_USE_API=true` (header override)
- [ ] Al menos 1 store migrado a API real con fixture como fallback
- [ ] Route guards protegen al menos `/rh/*` y `/evaluacion/*`
- [ ] `ForbiddenState` se muestra al acceder sin permiso
- [ ] `hasPermission()` funciona en componentes
- [ ] `go build ./...` compila sin errores
- [ ] `pnpm run check` pasa sin errores de tipo
- [ ] Todos los tests unitarios existentes siguen pasando
- [ ] Integración test: 78/78 rutas pasan contra Docker

---

## 9. Archivos afectados (resumen)

### Frontend (nuevos)
- `web/src/routes/login/+page.svelte`
- `web/src/routes/login/+page.server.ts`
- `web/src/routes/+layout.server.ts`
- `web/src/lib/components/AuthGuard.svelte`
- `web/src/lib/api/authService.ts` (wrapper de logout/refresh)

### Frontend (modificados)
- `web/src/lib/api/client.ts` — header DEV override
- `web/src/lib/api/session.svelte.ts` — logout(), hasPermission(), hasRole()
- `web/src/lib/components/Sidebar.svelte` — botón de logout
- `web/src/lib/components/DevToolbar.svelte` — guardar perfil en sessionStorage
- 5 stores — migración a async (ver sección 7)
- `web/src/routes/rh/+layout.server.ts` (nuevo guard)
- `web/src/routes/evaluacion/+layout.server.ts` (nuevo guard)

### Backend (nuevos)
- `api/internal/middleware/dev_override.go`

### Backend (modificados)
- `api/internal/handler/auth/handler.go` — logout con SSO fallback

### Backend (intactos, solo referencia)
- `api/internal/middleware/auth.go` — `RequireAuth`, `RequirePermission` (ya implementados en C7)
