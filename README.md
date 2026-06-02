# SED — Portal de evaluación de desempeño

Portal web para fijación y evaluación de objetivos de empleados (catálogo, mis evaluados, mi evaluación, objetivos).

**Estado:** especificación y principios. Sin implementación de aplicación aún.

## Stack (objetivo)

| Capa | Tecnología |
|------|------------|
| Frontend | Vite, Svelte 5, DaisyUI, TypeScript |
| Backend | Go 1.22+, router Chi, ORM (Ent recomendado; GORM alternativa) |
| Base de datos | PostgreSQL |
| Contratos | OpenAPI 3.1 (fuente de verdad en API) + tipos TS generados |
| Specs / SDD | [OpenSpec](https://github.com/Fission-AI/OpenSpec) |
| Contenedores | Docker multi-stage (futuro `api/`, `web/`) |

## Requerimientos funcionales (resumen)

- **Catálogo de objetivos:** lista, alta por indicador, descripción y categoría.
- **Mis evaluados:** desempeño y puesto, lista de empleados, evaluación por tipo de objetivo, fijación/evaluación, notificaciones por correo.
- **Mi evaluación:** consulta de la evaluación propia.
- **Objetivos:** fijación y evaluación con metas, por desempeño, etc.

Detalle y reglas de negocio: carpeta `openspec/specs/` (Spec-Driven Development).

## Estructura del repositorio (actual)

```
sed-evaluacion-desempeno/
├── AGENTS.md              # Reglas para agentes y humanos
├── README.md
├── principles/            # Decisiones y estándares (sin código de app)
├── docs/get-started/      # Guías de arranque
├── openspec/              # SDD: specs, changes, archive
├── api/                   # (futuro) servicio Go
└── web/                   # (futuro) Vite + Svelte
```

## OpenSpec (Spec-Driven Development)

Tras clonar o abrir el proyecto, inicializa o actualiza la integración con Cursor:

```bash
cd sed-evaluacion-desempeno
npm install -g @fission-ai/openspec@latest   # Node >= 20.19
openspec init --tools cursor --force
```

Flujo sugerido:

1. `/opsx:propose` — nueva capacidad (ej. autenticación, módulo catálogo).
2. Revisar artefactos en `openspec/changes/<nombre>/`.
3. `/opsx:apply` — implementación cuando la spec esté aprobada.
4. `/opsx:archive` — cerrar cambio completado.

Ver `docs/get-started/openspec.md`.

## Cómo levantar (cuando exista código)

Hoy solo hay documentación. Cuando existan `api/` y `web/`:

### Prerrequisitos

- Go 1.22+
- Node 20.19+ y **pnpm** (recomendado)
- Docker y Docker Compose
- PostgreSQL 15+ (local o contenedor)

### Desarrollo local (previsto)

```bash
# Base de datos
docker compose up -d db

# API Go (hot reload con air)
cd api && air

# Frontend
cd web && pnpm install && pnpm dev
```

### Producción (previsto)

- Imagen `api`: binario Go estático (multi-stage).
- Imagen `web`: build Vite servido por nginx o contenedor Node según decisión en spec.
- Proxy (Caddy/Traefik/Nginx): `/api` → Go, `/` → frontend.
- AWS: ECS Fargate + RDS PostgreSQL + ALB.

Detalle: `docs/get-started/local-setup.md` y `principles/architecture.md`.

## Contratos API ↔ frontend

- OpenAPI generado o mantenido desde el backend.
- Cliente TS con `openapi-typescript` + `openapi-fetch` (sin axios pesado).
- Paginación cursor-based en listas grandes; campos mínimos en listados, detalle en segundo request.
- Caché HTTP y ETag donde aplique; sin duplicar reglas de negocio en el cliente.

Ver `principles/contracts-api.md`.

## Contribuir

1. Leer `AGENTS.md` y `principles/README.md`.
2. Proponer cambio vía OpenSpec antes de código nuevo.
3. Commits imperativos: `Add`, `Fix`, `Refactor` + descripción.

## Licencia

Por definir.
