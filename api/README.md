# API (Go) — Implementado

Servicio REST con **Chi**, **Ent ORM**, **PostgreSQL**, **OpenAPI 3.1**.

## Estructura

```
api/
├── cmd/
│   ├── server/      # Servidor HTTP principal
│   └── import/      # CLI de importación/seed
├── internal/        # Lógica de dominio, handlers, repositorios
├── openapi/         # 7 specs YAML (fuente de verdad)
├── .air.toml        # Hot reload (air)
├── Dockerfile       # Multi-stage producción
├── Dockerfile.dev   # Desarrollo con hot reload
├── entc.yaml        # Config Ent codegen
├── go.mod / go.sum
└── README.md
```

## Dominios OpenAPI (`api/openapi/`)

| Archivo | Dominio | Descripción |
|---------|---------|-------------|
| `auth.yaml` | Autenticación | Login, logout, refresh, sesión httpOnly, SSO (Keycloak/OIDC) |
| `cycle.yaml` | Ciclos | CRUD ciclos de evaluación, estados, fechas |
| `goals-api.yaml` | Objetivos | Catálogo, asignación, avance, metas, indicadores |
| `evaluations-and-9x9.yaml` | Evaluaciones | Evaluaciones por empleado, matriz 9x9 (desempeño × potencial) |
| `competency-framework-api.yaml` | Competencias | Pilares, competencias, criterios, niveles de aceptación |
| `org-hierarchy.yaml` | Organización | Jerarquía, empleados, puestos, reportes |
| `activity-logs.yaml` | Auditoría | Logs de actividad, cambios en evaluaciones |

## Comandos

### Desarrollo (hot reload)

```bash
cd api
air                    # Requiere air instalado (go install github.com/air-verse/air@latest)
```

### Build

```bash
cd api
go build -o ./tmp/main ./cmd/server   # Binario en tmp/
```

### Generar código Ent (tras cambios de schema)

```bash
cd api
go generate ./...   # o: ent generate ./internal/ent/schema
```

### Validar specs OpenAPI

```bash
# Instalar validador (una vez)
npm install -g @redocly/openapi-cli@latest

# Validar todos los specs
cd api/openapi
redocly lint *.yaml
```

### Generar cliente TypeScript (para frontend)

```bash
cd web
pnpm gen:api
# Genera tipos en src/lib/api/schemas/*.d.ts desde cada .yaml
```

## Variables de entorno (`.env`)

```bash
DATABASE_URL=postgres://sed:sed@localhost:5432/sed?sslmode=disable
ENV=development
CORS_ORIGINS=http://localhost:5173
SSO_KC_ISSUER=...
SSO_CLIENT_ID=...
SSO_CLIENT_SECRET=...
SSO_REDIRECT_URI=...
SSO_POST_LOGOUT_URI=...
```

## Docker

```bash
# Desarrollo (con hot reload)
docker compose -f docker-compose.dev.yml up --build api

# Producción
docker compose up --build api
```

## Puertos

- API: `8080` (interno) → expuesto en `8080` (dev) o via proxy `8088` (prod)
- Health check: `GET /healthz`

## Principios

- Handlers delgados → lógica en `internal/`
- Validación: Go `validator` + esquemas OpenAPI
- RBAC en backend siempre
- Migraciones versionadas con Ent
- Índices compuestos alineados a queries de listado

Ver `principles/` y `AGENTS.md`.