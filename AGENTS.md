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

1. Catálogo de objetivos  
2. Mis evaluados  
3. Mi evaluación  
4. Objetivos (fijación / evaluación / metas)

Cada módulo = al menos una spec OpenSpec antes de UI o handlers.

## Documentación humana

- Principios: `principles/`
- Arranque: `docs/get-started/`
- Este archivo prevalece para agentes si hay conflicto con README salvo decisiones de producto en OpenSpec.

## Idioma

- UI y mensajes de usuario: español.
- Código y comentarios técnicos: inglés o español según consistencia del archivo.
