# AGENTS.md — SED Evaluación de desempeño

Instrucciones para agentes de IA y desarrolladores. Prioridad: **especificación antes que código**.

## Proyecto

Portal SED: evaluación de empleados, objetivos, ciclos, notificaciones. Escala objetivo: miles a millones de usuarios (multi-tenant o partición por organización según spec).

## Stack obligatorio (cuando se implemente)

| Área | Elección | Notas |
|------|----------|--------|
| Frontend | Vite + Svelte 5 + TypeScript + DaisyUI | SPA; lazy routes; accesibilidad WCAG 2.1 AA donde aplique |
| Paquetes JS | **pnpm** únicamente | Lockfile comprometido; `pnpm audit` en CI |
| Backend | Go + **Chi** | Handlers delgados; lógica en `internal/` |
| ORM | **Ent** (preferido) o GORM | Migraciones versionadas; índices explícitos en spec de datos |
| BD | PostgreSQL | Solo SQL preparado vía ORM/query builders |
| Contratos | **OpenAPI 3.1** | Fuente de verdad; validar requests/responses |
| Auth | Sesión httpOnly + rotación; JWT solo si spec lo exige | RBAC en backend siempre |
| Specs | OpenSpec | No implementar features sin change/spec aprobado |
| Contenedores | Docker multi-stage | Usuario no-root; secrets por env |

## No hacer

- Código de aplicación sin artefacto OpenSpec en `openspec/changes/` (salvo spike acotado y documentado).
- Concatenar SQL crudo con input de usuario.
- Guardar tokens en `localStorage`.
- `npm run build` en React (no aplica); en Svelte usar `pnpm run lint` o `tsc`.
- Commits o push sin pedido explícito del usuario.
- Box-shadow decorativo ni bordes solo izquierda/derecha en UI (ver reglas de diseño del usuario).
- Sobre-ingeniería: helpers de una línea, abstracciones prematuras.

## Arquitectura (objetivo)

```
web/          → Vite + Svelte, consume OpenAPI
api/          → Go, expone REST JSON, Ent/GORM + pgx
openspec/     → specs duraderas y changes activos
principles/   → decisiones transversales
```

- **Bounded contexts:** catálogo, evaluaciones/ciclos, identidad/acceso, notificaciones.
- **API:** REST versionada `/api/v1/`; idempotencia en escrituras críticas.
- **Listados:** paginación por cursor; filtros indexados; proyecciones ligeras.
- **Frontend:** code-splitting por ruta; skeletons en carga; errores tipados desde OpenAPI.

## Contratos y rendimiento

1. Definir OpenAPI antes del handler público.
2. Generar tipos TS (`openapi-typescript`) en `web/src/lib/api/`.
3. Listados: máximo campos necesarios para tabla; detalle en `GET /resource/:id`.
4. Backend: índices compuestos alineados a queries de listado (documentar en spec de BD).
5. Compresión gzip/brotli en proxy; `Cache-Control` en lecturas estables.

## Seguridad

- Validación entrada: Go `validator` + esquemas OpenAPI.
- RBAC: roles → permisos → recursos (empleado, ciclo, objetivo).
- Rate limiting en login y APIs sensibles.
- Auditoría de cambios en evaluaciones (quién, cuándo, qué).
- Dependencias: Renovate/Dependabot; Trivy en imágenes Docker.

## Calidad de código

- Go: `golangci-lint`, tests de tabla en dominio y handlers.
- Svelte/TS: ESLint + Prettier; tests con Vitest + Testing Library.
- E2E: Playwright (flujos críticos: login, fijar objetivo, enviar evaluación).
- Commits: `Add` / `Fix` / `Refactor` + descripción en español o inglés consistente.

## OpenSpec

- Proponer: `/opsx:propose "<descripción>"`
- Tras aprobar spec: `/opsx:apply`
- Al cerrar: `/opsx:archive`
- Validar: `openspec validate --all`

Artefactos en `openspec/specs/` son la verdad a largo plazo; `openspec/changes/` es trabajo en curso.

## Módulos de producto (referencia)

**Implementados (con specs OpenAPI en `api/openapi/` y rutas en `web/src/routes/`):**

1. **Goals / Objetivos** — fijación, avance, asignación, biblioteca (`goals-api.yaml`)
2. **Org-Hierarchy / Jerarquía organizacional** — árbol, evaluados por jefe (`org-hierarchy.yaml`)
3. **Nine-Box 9×9** — matriz 9×9, competencias por empleado (`evaluations-and-9x9.yaml`)
4. **Competency-Framework / Marco de competencias** — pilares, competencias, niveles, criterios (`competency-framework-api.yaml`)
5. **Activity-Logs / Registros de actividad** — auditoría de cambios (`activity-logs.yaml`)
6. **Comment-Changes / Cambios de comentario** — propuestas, revisiones, historial (parte de goals/evaluations)
7. **Catálogo** — catálogo base de objetivos/KPIs
8. **Evaluaciones / Ciclos** — ciclos, criterios de escala, niveles de aceptación (`cycle.yaml`, `evaluations-and-9x9.yaml`)
9. **Identidad / Acceso** — auth, sesión, RBAC (`auth.yaml`)
10. **Notificaciones** — toast, alertas, centro de notificaciones

Cada módulo = al menos una spec OpenAPI en `api/openapi/` antes de UI o handlers.

## Documentación humana

- Principios: `principles/`
- Arranque: `docs/get-started/`
- Este archivo prevalece para agentes si hay conflicto con README salvo decisiones de producto en OpenSpec.

## Cómo orientarse rápido

**Leer primero `principles/`** — AGENTS.md solo enlaza, no duplica:

- `principles/architecture.md` — arquitectura general, bounded contexts, API, frontend
- `principles/contracts-api.md` — OpenAPI, generación de tipos, validación, versionado
- `principles/data-and-orm.md` — Ent/GORM, migraciones, índices, transacciones
- `principles/evaluations-domain.md` — dominio de evaluaciones, ciclos, nine-box, competencias

## Idioma

- UI y mensajes de usuario: español.
- Código y comentarios técnicos: inglés o español según consistencia del archivo.


<!-- headroom:rtk-instructions -->
# RTK (Rust Token Killer) - Token-Optimized Commands

When running shell commands, **always prefix with `rtk`**. This reduces context
usage by 60-90% with zero behavior change. If rtk has no filter for a command,
it passes through unchanged — so it is always safe to use.

## Key Commands
```bash
# Git (59-80% savings)
rtk git status          rtk git diff            rtk git log

# Files & Search (60-75% savings)
rtk ls <path>           rtk read <file>         rtk grep <pattern>
rtk find <pattern>      rtk diff <file>

# Test (90-99% savings) — shows failures only
rtk pytest tests/       rtk cargo test          rtk test <cmd>

# Build & Lint (80-90% savings) — shows errors only
rtk tsc                 rtk lint                rtk cargo build
rtk prettier --check    rtk mypy                rtk ruff check

# Analysis (70-90% savings)
rtk err <cmd>           rtk log <file>          rtk json <file>
rtk summary <cmd>       rtk deps                rtk env

# GitHub (26-87% savings)
rtk gh pr view <n>      rtk gh run list         rtk gh issue list

# Infrastructure (85% savings)
rtk docker ps           rtk kubectl get         rtk docker logs <c>

# Package managers (70-90% savings)
rtk pip list            rtk pnpm install        rtk npm run <script>
```

## Rules
- In command chains, prefix each segment: `rtk git add . && rtk git commit -m "msg"`
- For debugging, use raw command without rtk prefix
- `rtk proxy <cmd>` runs command without filtering but tracks usage
<!-- /headroom:rtk-instructions -->
