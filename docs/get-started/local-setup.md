# Setup local — Desarrollo

Guía para levantar el entorno de desarrollo completo (API + Web + DB).

## Prerrequisitos

- **Go 1.22+** — `go version`
- **Node 20.19+** y **pnpm** — `pnpm --version`
- **Docker** y **Docker Compose** — `docker compose version`
- **PostgreSQL 15+** (se levanta vía Docker)
- **air** (hot reload Go) — `go install github.com/air-verse/air@latest`
- **redocly/openapi-cli** (validar specs) — `npm install -g @redocly/openapi-cli@latest`

## Variables de entorno

Copiar y ajustar:

```bash
cp .env.example .env
```

Variables clave en `.env`:

```bash
# Base de datos
DATABASE_URL=postgres://sed:sed@localhost:5432/sed?sslmode=disable

# API
ENV=development
CORS_ORIGINS=http://localhost:5173
API_PORT=8080

# SSO / Keycloak (OIDC) — requeridos para auth real
SSO_KC_ISSUER=https://sso.example.com/realms/sed
SSO_CLIENT_ID=sed-portal
SSO_CLIENT_SECRET=*****
SSO_REDIRECT_URI=http://localhost:5173/callback
SSO_POST_LOGOUT_URI=http://localhost:5173/login

# Seed/import (opcional)
SSO_SEED_BASE_URL=https://sso.example.com
SSO_SEED_ADMIN_USER=admin
SSO_SEED_ADMIN_PASSWORD=*****
SEED_SSO_DEV_EMPLOYEE_NUMBER=12345
```

## Opción A: Docker Compose (recomendado — todo en contenedores)

```bash
# Levantar DB + API (hot reload) + Web (hot reload) + Proxy
docker compose -f docker-compose.dev.yml up --build

# Servicios:
# - postgres: localhost:5432
# - api:      localhost:8080 (hot reload con air)
# - web:      localhost:5173 (Vite dev server)
# - proxy:    localhost:8088 (nginx: API en /api/, Web en /)
```

Detener: `docker compose -f docker-compose.dev.yml down`

Volver a levantar sin rebuild: `docker compose -f docker-compose.dev.yml up`

## Opción B: Procesos nativos (API y Web fuera de Docker, solo DB en Docker)

### 1. Base de datos

```bash
# Solo PostgreSQL
docker compose -f docker-compose.dev.yml up -d postgres
# Verificar: docker compose -f docker-compose.dev.yml logs -f postgres
```

### 2. API Go (hot reload con air)

```bash
cd api
air
# Escucha en :8080, recarga automática en cambios .go
```

### 3. Frontend

```bash
cd web
pnpm install
pnpm dev
# Vite dev server en http://localhost:5173
```

### 4. Generar cliente TypeScript (tras cambios en OpenAPI)

```bash
cd web
pnpm gen:api
# Regenera src/lib/api/schemas/*.d.ts desde api/openapi/*.yaml
```

## Validar specs OpenAPI

```bash
cd api/openapi
redocly lint *.yaml
```

## Estructura de puertos

| Servicio | Puerto (host) | Notas |
|----------|---------------|-------|
| PostgreSQL | 5432 (dev) / 5435 (prod) | `docker-compose.dev.yml` usa 5432 |
| API Go | 8080 | Health: `GET /healthz` |
| Web (Vite) | 5173 | Solo en dev nativo |
| Proxy (nginx) | 8088 | Une `/api/` → API, `/` → Web |

## Comandos útiles

```bash
# Ver logs
docker compose -f docker-compose.dev.yml logs -f api
docker compose -f docker-compose.dev.yml logs -f web

# Rebuild solo un servicio
docker compose -f docker-compose.dev.yml up --build api

# Entrar a contenedor API
docker compose -f docker-compose.dev.yml exec api sh

# Migraciones Ent (si cambias schema)
cd api && go generate ./...

# Tests
cd api && go test ./...
cd web && pnpm test
```

## Troubleshooting

| Problema | Solución |
|----------|----------|
| `air` no recarga | Verificar `.air.toml` exclude_dir; `air -c .air.toml` |
| Puerto 5432 ocupado | Cambiar en `docker-compose.dev.yml` o parar otro postgres |
| Error CORS | Ajustar `CORS_ORIGINS` en `.env` (incluir `http://localhost:5173`) |
| Tipos TS desactualizados | `cd web && pnpm gen:api` |
| `redocly lint` falla | Revisar YAML en `api/openapi/`; specs deben ser OpenAPI 3.1 válido |

## Producción (referencia)

```bash
# Build imágenes multi-stage
docker compose build

# Levantar stack completo (DB + API + Web + Proxy + certs TLS)
docker compose up -d

# Ver logs
docker compose logs -f
```

Ver `docker-compose.yml` para configuración de producción (puertos, volúmenes, healthchecks, proxy TLS).