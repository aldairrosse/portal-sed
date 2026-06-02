# Arquitectura

## Visión

Sistema web en dos servicios desplegables: API Go y frontend Vite/Svelte, con PostgreSQL como almacén principal.

## Capas backend (futuro `api/`)

- `cmd/server` — entrada.
- `internal/handler` — HTTP, validación, mapeo DTO.
- `internal/service` — reglas de negocio.
- `internal/repository` — Ent/GORM + transacciones.
- `internal/auth` — sesión, RBAC.
- `internal/notify` — cola/correo (async).

## Capas frontend (futuro `web/`)

- `src/routes` — páginas por módulo SED.
- `src/lib/api` — cliente OpenAPI tipado.
- `src/lib/stores` — estado UI mínimo; servidor es fuente de verdad.
- `src/lib/components` — UI reutilizable DaisyUI.

## Escalabilidad (miles / millones de usuarios)

- Particionar por **organización/tenant** en todas las tablas críticas.
- Lecturas: réplicas PostgreSQL + caché (Redis) para sesión y catálogos estables.
- Escrituras: colas para correo y reportes pesados.
- API stateless horizontal detrás de load balancer.
- Frontend en CDN; assets con hash largo.

## Despliegue

- Docker Compose en dev: `db`, `api`, `web`, opcional `mailpit`.
- Prod: ECS Fargate o VMs + RDS + ALB; secrets en SSM/Parameter Store.

## Pendiente de spec OpenSpec

- Multi-tenant vs single-tenant.
- SSR vs SPA pura (hoy SPA; evaluar SvelteKit si SEO/TTFB lo exigen).
