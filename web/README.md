# Web (Vite + SvelteKit 5 + DaisyUI 5)

Portal SED — Evaluación de desempeño. **Implementado** con 22+ rutas.

## Comandos

```bash
cd web

pnpm install          # Instalar dependencias
pnpm dev              # Desarrollo en http://localhost:5173 (Vite dev server)
pnpm build            # Build de producción (output en build/)
pnpm preview          # Previsualizar build de producción
pnpm check            # Type-check + svelte-check
pnpm lint             # ESLint
pnpm format           # Prettier --write
pnpm gen:api          # Generar tipos TS desde api/openapi/*.yaml → src/lib/api/schemas/
pnpm test             # Vitest run
pnpm test:watch       # Vitest watch mode
```

## Stack

- **Framework:** SvelteKit 2 (SPA, `ssr = false`, adapter-static)
- **UI:** DaisyUI 5 + Tailwind CSS 4
- **Lenguaje:** TypeScript 5
- **Paquete:** pnpm (lockfile comprometido)
- **Testing:** Vitest + Testing Library + jsdom
- **Cliente API:** `openapi-fetch` + tipos generados con `openapi-typescript`

## Rutas implementadas (`src/routes/`)

| Ruta | Descripción |
|------|-------------|
| `/` | Dashboard / landing |
| `/login` | Autenticación (SSO/OIDC) |
| `/mis-evaluados` | Lista y gestión de evaluados |
| `/mi-evaluacion` | Consulta de evaluación propia |
| `/objetivos/asignacion` | Asignación de objetivos |
| `/objetivos/asignacion/biblioteca` | Catálogo/biblioteca de objetivos |
| `/objetivos/avance` | Seguimiento de avance |
| `/evaluacion/9x9` | Matriz 9x9 (desempeño × potencial) |
| `/evaluacion/9x9/competencias` | Competencias en matriz 9x9 |
| `/evaluacion/9x9/jerarquia` | Vista jerárquica 9x9 |
| `/perfil` | Perfil de usuario |
| `/rh/ciclos` | Gestión de ciclos (RRHH) |
| `/rh/pilares` | Pilares de competencias |
| `/rh/pilares/[id]/competencias` | Competencias por pilar |
| `/rh/criterios-escala` | Criterios y escalas |
| `/rh/niveles-aceptacion` | Niveles de aceptación |
| `/rh/evaluaciones` | Gestión de evaluaciones (RRHH) |
| `/rh/jerarquia` | Jerarquía organizacional |
| `/dev/requisitos` | Dev tools (solo `import.meta.env.DEV`) |

## Generación de cliente API

```bash
pnpm gen:api
```

Lee los 7 specs en `../api/openapi/` y genera tipos en `src/lib/api/schemas/`:
- `auth.d.ts`
- `cycle.d.ts`
- `goals.d.ts`
- `evaluations.d.ts`
- `competency.d.ts`
- `org-hierarchy.d.ts`

## Variables de entorno

```bash
# .env (opcional, para build)
VITE_SHOW_DEV_TOOLBAR=true   # Mostrar toolbar de desarrollo
```

## Docker

```bash
# Desarrollo (hot reload)
docker compose -f docker-compose.dev.yml up --build web

# Solo web (puerto 8080)
docker compose -f docker-compose.web-dev.yml up --build

# Producción
docker compose up --build web
```

## Notas de implementación

- Selector de persona y fase en desarrollo (`import.meta.env.DEV`) — no aparece en producción.
- Tema `sed` DaisyUI con colores: primario `#1e3a5f`, secundario `#4a90a4`, acento `#c4a035`.
- `box-shadow` decorativo desactivado globalmente.
- Code-splitting por ruta (lazy loading automático de SvelteKit).
- Skeletons en carga; errores tipados desde OpenAPI.