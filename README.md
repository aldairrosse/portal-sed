# SED — Portal de evaluación de desempeño

Portal web para fijación y evaluación de objetivos de empleados (catálogo, mis evaluados, mi evaluación, objetivos).

**Estado:** implementado — backend Go + frontend SvelteKit operativos.

## Stack

| Capa | Tecnología |
|------|------------|
| Frontend | Vite, Svelte 5, SvelteKit (SPA), DaisyUI 5, Tailwind CSS 4, TypeScript |
| Backend | Go 1.22+, router Chi, ORM Ent, PostgreSQL |
| Base de datos | PostgreSQL 15+ |
| Contratos | OpenAPI 3.1 (fuente de verdad en `api/openapi/`) + tipos TS generados |
| Specs / SDD | [OpenSpec](https://github.com/Fission-AI/OpenSpec) |
| Contenedores | Docker multi-stage (`api/`, `web/`) |

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
├── api/                   # Go + Chi + Ent + OpenAPI 3.1 (7 specs)
│   ├── cmd/               # Entry points (server, import)
│   ├── internal/          # Lógica de dominio y handlers
│   ├── openapi/           # 7 specs YAML (auth, cycle, goals, evaluations, competency, org-hierarchy, activity-logs)
│   └── .air.toml          # Hot reload config
└── web/                   # Vite + SvelteKit (SPA) + DaisyUI
    ├── src/routes/        # 22+ rutas (login, mis-evaluados, mi-evaluacion, objetivos, evaluacion/9x9, rh/*, perfil, dev)
    ├── src/lib/api/       # Cliente TS generado desde OpenAPI
    └── package.json       # Scripts: dev, build, check, lint, format, gen:api, test
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

## Cómo levantar (desarrollo local)

### Prerrequisitos

- Go 1.22+
- Node 20.19+ y **pnpm** (recomendado)
- Docker y Docker Compose
- PostgreSQL 15+ (local o contenedor)

### Opción A: Docker Compose (recomendado)

```bash
# Levantar todo (DB + API + Web + Proxy)
docker compose -f docker-compose.dev.yml up --build
# API en http://localhost:8080, Web en http://localhost:5173 (o proxy en 8088)
```

### Opción B: Procesos nativos

```bash
# 1. Base de datos
docker compose up -d postgres   # o usar docker-compose.dev.yml solo postgres

# 2. API Go (hot reload con air)
cd api && air

# 3. Frontend
cd web && pnpm install && pnpm dev
```

### Generar cliente TypeScript (frontend)

```bash
cd web && pnpm gen:api
# Genera tipos en src/lib/api/schemas/*.d.ts desde api/openapi/*.yaml
```

## Contratos API ↔ frontend

- OpenAPI mantenido en `api/openapi/` (7 archivos YAML, uno por dominio).
- Cliente TS con `openapi-typescript` + `openapi-fetch` en `web/src/lib/api/`.
- Paginación cursor-based en listas grandes; campos mínimos en listados, detalle en segundo request.
- Caché HTTP y ETag donde aplique; sin duplicar reglas de negocio en el cliente.

Ver `principles/contracts-api.md`.

## Dominios OpenAPI (api/openapi/)

| Archivo | Dominio |
|---------|---------|
| `auth.yaml` | Autenticación y sesiones |
| `cycle.yaml` | Ciclos de evaluación |
| `goals-api.yaml` | Objetivos (catálogo, asignación, avance) |
| `evaluations-and-9x9.yaml` | Evaluaciones y matriz 9x9 |
| `competency-framework-api.yaml` | Marco de competencias (pilares, criterios, niveles) |
| `org-hierarchy.yaml` | Jerarquía organizacional |
| `activity-logs.yaml` | Auditoría / logs de actividad |

## Contribuir

1. Leer `AGENTS.md` y `principles/README.md`.
2. Proponer cambio vía OpenSpec antes de código nuevo.
3. Commits imperativos: `Add`, `Fix`, `Refactor` + descripción.

## Licencia

Por definir.