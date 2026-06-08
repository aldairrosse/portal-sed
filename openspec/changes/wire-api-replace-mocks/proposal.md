# Propuesta: C8 — wire-api-replace-mocks

## 1. Intent

Reemplazar los datos sincrónicos de fixture (`$lib/fixtures/`) en los stores de Svelte 5 con llamadas asíncronas a las APIs reales del backend (`api/internal/handler/`), usando `openapi-fetch` + tipos generados desde `api/openapi/*.yaml`. Los fixtures se conservan como fallback exclusivo en desarrollo (`import.meta.env.DEV`) para permitir el trabajo offline y la iteración rápida de UI sin backend corriendo.

**Especificaciones base:** `api/openapi/*.yaml` (goals-api, competency-framework-api, evaluations-and-9x9, org-hierarchy, auth, cycle)

**Decisiones de arquitectura referenciadas:**
- C2–C6 handlers ya implementados y archivados — las APIs existen y responden.
- C7 (identity-access) implementó `RequireAuth` middleware — C8 lo conecta en el frontend.
- Stores actuales son sincrónicos (modifican arrays en memoria) — C8 los migra a async con fetch real.
- `devContext.svelte.ts` hoy maneja profile/phase con valores mockeados — en producción esos valores vienen de `GET /auth/me` y `GET /cycle/current`.

## 2. Scope

### In Scope

- **Instalar dependencias:** `openapi-fetch` (cliente HTTP tipado) + `openapi-typescript` (generador de tipos desde OpenAPI).
- **Generar tipos TypeScript** desde cada spec YAML en `api/openapi/` usando `openapi-typescript`.
- **Crear `web/src/lib/api/client.ts`:** Cliente HTTP tipado con base URL configurable, interceptors para adjuntar token de sesión (desde cookie/httpOnly), y manejo de errores estándar.
- **Crear `web/src/lib/api/schemas/`:** Tipos generados por `openapi-typescript` para cada dominio.
- **Crear `web/src/lib/session.svelte.ts`:** Store runa (`$state`) que expone el usuario autenticado, expiración de sesión y método `ensureSession()` que llama a `GET /auth/me` (con redirect a login si 401).
- **Crear `web/src/lib/cycle.svelte.ts`:** Store runa que expone la fase activa del ciclo desde `GET /cycle/current`, reemplazando la lógica de `devContext` para fase.
- **Reescribir cada store a async** con el siguiente patrón:

  ```ts
  // Antes (sincrónico con fixture):
  let items = $state(structuredClone(fixtureData));
  export function getItems() { return items; }

  // Después (async con fetch + fixture de desarrollo como fallback):
  let data = $state<Data | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

  export async function load(): Promise<void> {
    if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
      data = structuredClone(fixtureData);
      loading = false;
      return;
    }
    const { data: result, error: apiError } = await client.GET('/path');
    if (apiError) { error = apiError.error.message; loading = false; return; }
    data = result;
    loading = false;
  }
  ```

- **Stores a reescribir:**
  | Store | Reemplazo API | Fixture |
  |---|---|---|
  | `goalsStore.svelte.ts` | `GET /api/v1/employees/:empId/categories`, `POST/PUT/DELETE /goals/*` | `fixtures/goals/*` |
  | `nineBoxStore.svelte.ts` | `GET /api/v1/evaluations/:evaluatedId/competency-ratings`, matrix endpoints | `fixtures/nine-box/*` |
  | `competencyStore.svelte.ts` | `GET /api/v1/pillars`, `GET /api/v1/competencies`, etc. | `fixtures/competency/*` |
  | `evaluationStore.svelte.ts` | `GET /api/v1/evaluations/*`, `POST /evaluations/:id/ratings`, etc. | `fixtures/evaluations/*` |
  | `orgHierarchyStore.svelte.ts` | `GET /api/v1/org-tree` | `fixtures/org-hierarchy/*` |

- **Añadir estados de carga y error** en componentes que consumen stores: `<Loader />` para `loading === true`, `<ErrorState />` para `error !== null`.
- **Actualizar `svelte.config.js` o `vite.config.ts`:** Configurar `$api` alias que apunte a `$lib/api/`.
- **Conectar el middleware `RequireAuth` del backend** en las rutas de C2–C6 handlers (quitar `TODO(auth:C7)` placeholders).

### Out of Scope

- **No se modifican los handlers del backend** (`api/internal/handler/`) — las APIs ya entregan los datos que el frontend necesita.
- **No se eliminan los fixtures** — se conservan como fallback en `import.meta.env.DEV` y se documenta su propósito.
- **No se implementa UI de login** — eso es un change futuro. C8 asume que el usuario ya está autenticado (sesión existente) y solo conecta el token a los requests.
- **No se implementa caching avanzado** (SWR, React Query style) — la primera iteración es fetch directo + loading state. El caching y optimistic updates se abordan en change posterior si se justifica.
- **No se implementan tests E2E** — se prioriza que el refactor compile y funcione correctamente; los tests E2E se agregan como change separado.
- **No se toca `devContext.svelte.ts`** — sigue existiendo para desarrollo local (selector de perfil/fase mockeada). En producción, los stores toman la fase desde `cycle.svelte.ts` y el perfil desde `session.svelte.ts`.

## 3. Approach

### Capa de API client

```
web/src/lib/api/
├── client.ts          # openapi-fetch instance con baseURL, headers, error handling
├── schemas/           # tipos generados por openapi-typescript
│   ├── goals.d.ts
│   ├── competency.d.ts
│   ├── evaluations.d.ts
│   ├── org-hierarchy.d.ts
│   ├── auth.d.ts
│   └── cycle.d.ts
├── session.svelte.ts  # store runa: usuario autenticado, token management
└── cycle.svelte.ts    # store runa: fase activa del ciclo
```

**`client.ts` contrato:**

```ts
import createClient from 'openapi-fetch';
import type { paths } from './schemas/goals';

export const apiClient = createClient<paths>({
  baseURL: import.meta.env.VITE_API_URL ?? '/api/v1',
  headers: { 'Content-Type': 'application/json' },
});

// Interceptor: adjunta token de sesión desde cookie httpOnly (automático con credentials: 'include')
apiClient.use({
  onRequest: ({ request }) => {
    request.credentials = 'include';
    return request;
  },
  onResponse: ({ response }) => {
    if (response.status === 401) {
      // Token expirado o inválido → redirect a login
      window.location.href = '/login';
    }
    return response;
  },
});
```

### Estrategia de migración por store

Cada store se transforma siguiendo estos pasos:

1. **Añadir triplete de estado:** `data`, `loading`, `error` como `$state`.
2. **Extraer lógica de sincronización a función async** `load()` que decide entre fixture (dev) o API (prod).
3. **Añadir `reload()`** para refetch manual cuando se requiere (ej: después de una mutación).
4. **Mutaciones locales se convierten en llamadas API** (`POST`, `PUT`, `DELETE`) seguidas de `reload()` optimista (o await en primera iteración).
5. **Componentes consumidores se actualizan** para mostrar `<Loader />` y `<ErrorState />` según corresponda.

### Manejo de sesión y ciclo

| Contexto | Dev (fixture) | Producción |
|---|---|---|
| Perfil (rol) | `devContext.getProfile()` — selector manual en UI | `session.user.profile` desde `GET /auth/me` |
| Fase del ciclo | `devContext.getPhase()` — selector manual en UI | `cycle.activePhase` desde `GET /cycle/current` |
| Empleado ID | Mockeado en fixture o desde selector | `session.user.employeeId` desde `GET /auth/me` |

### Transiciones y UX

- Cada store expone `loading` (boolean) que los componentes usan para mostrar skeletons con DaisyUI.
- Cada store expone `error` (string | null) para mostrar `<ErrorState />` con opción de reintentar.
- Las mutaciones (crear/editar/eliminar) usan `fetch` directo con `await` y actualizan el estado local optimistamente cuando es seguro, o llaman `reload()` después de éxito.
- No se implementa `use:enhance` en formularios — los handlers de submit llaman a la mutación del store directamente.

### Conexión de `RequireAuth` en handlers existentes

C7 documentó los placeholders `TODO(auth:C7)` en C2–C6. C8:
1. Busca todos los `TODO(auth:C7)` en `api/internal/handler/`.
2. Reemplaza `AuthPlaceholder` con `RequireAuth` + `RequirePermission` según el endpoint.
3. Verifica que el middleware de sesión se ejecute antes de llegar a los handlers.

## 4. Dependencies

- **C2–C6:** Handlers del backend ya implementados, probados y archivados. Las rutas existen en `api/internal/handler/`.
- **C7 (identity-access):** Middleware `RequireAuth` y `RequirePermission` implementados. Sesión manejada vía cookie httpOnly + token SHA-256.
- **C1 (data-model-core):** Esquema de base de datos en Ent, migraciones aplicadas.
- **`api/openapi/*.yaml`:** Especificaciones OpenAPI 3.1 de cada dominio — fuente de verdad para generación de tipos.

## 5. Risk & Mitigation

### Riesgo: Regression silenciosa en componentes que consumen stores

- **Mitigación:** Cada store migrado mantiene la misma API pública (nombres de funciones exportadas, tipos de retorno). Los componentes existentes requieren cambios solo para agregar loading/error states, no para cambiar lógica de negocio.

### Riesgo: Backend no responde exactamente como el fixture

- **Mitigación:** El cambio se prueba contra el backend real en staging antes de merge. Los tipos generados desde OpenAPI garantizan que el frontend espera la estructura correcta. Cualquier discrepancia se documenta y corrige del lado de la spec.

### Riesgo: Migración grande y difícil de revisar

- **Mitagación:** Se planea como una serie de PRs pequeños: (1) setup de herramientas + client, (2) migración de un store piloto (goalsStore), (3–5) stores restantes en paralelo, (6) conexión de auth middleware.

### Riesgo: Sesión expirada durante uso

- **Mitigación:** El interceptor de `client.ts` detecta `401` y redirige a login. El store `session.svelte.ts` refresca la sesión automáticamente antes de expirar usando el endpoint `POST /auth/refresh`.

## 6. Success Criteria

- [ ] `openapi-fetch` + `openapi-typescript` instalados y generan tipos sin errores desde todos los `api/openapi/*.yaml`.
- [ ] `client.ts` configurado con base URL desde `VITE_API_URL`, `credentials: 'include'`, y manejo de `401`.
- [ ] `session.svelte.ts` expone `user`, `loading`, `error` y llama `GET /auth/me` al inicializar.
- [ ] `cycle.svelte.ts` expone `activePhase`, `loading`, `error` desde `GET /cycle/current`.
- [ ] Los 5 stores migrados (goals, nineBox, competency, evaluation, orgHierarchy) son asíncronos con triplete `data`/`loading`/`error`.
- [ ] En `import.meta.env.DEV` sin `VITE_USE_API=true`, los stores cargan desde fixtures (mismo comportamiento actual).
- [ ] En producción (o con `VITE_USE_API=true`), los stores fetch desde el backend real.
- [ ] Los componentes consumidores muestran `<Loader />` mientras `loading === true` y `<ErrorState />` cuando `error !== null`.
- [ ] Los placeholders `TODO(auth:C7)` en `api/internal/handler/` están reemplazados con `RequireAuth`/`RequirePermission`.
- [ ] `pnpm run check` (svelte-check) pasa sin errores de tipo.
- [ ] `go build` en `api/` compila sin errores tras conectar auth middleware.
