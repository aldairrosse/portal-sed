# reconciliation-gaps Specification

## Purpose

Documentar y rastrear los **gaps de reconciliación** entre la especificación OpenAPI, los artefactos OpenSpec en `openspec/changes/`, y la implementación real en Go. Esta spec sirve como fuente de verdad para cerrar inconsistencias antes de que se conviertan en deuda técnica.

**Alcance**: Tres gaps identificados que bloquean la confianza en el contrato API↔Handler↔Spec.

## Gap 1: Referencia rota a `CompetencyResultItem` en OpenAPI

### Contexto

El archivo `api/openapi/evaluations-and-9x9.yaml` referencia un schema `CompetencyResultItem` que **no existe** en `components.schemas`.

### Evidencia

- **Ubicación**: `api/openapi/evaluations-and-9x9.yaml` línea ~296 (en la respuesta de `GET /evaluations/competency-results`)
- **Referencia**: `{ $ref: "#/components/schemas/CompetencyResultItem" }`
- **Realidad**: El schema no está definido en `components.schemas` del mismo archivo
- **Existencia en código Go**: El tipo `dto.CompetencyResultItem` **sí existe** en `api/internal/dto/evaluation/evaluation_dto.go:112` y se usa en `evaluation_service.go:616`

### Impacto

- Generación de clientes TS falla (schema inválido)
- Documentación Swagger/UI muestra error
- Contrato API no valida contra implementación real

### Requisito de corrección

El sistema SHALL definir `CompetencyResultItem` en `components.schemas` del OpenAPI, alineado 1:1 con el DTO Go.

#### Scenario: Schema CompetencyResultItem definido en OpenAPI

- GIVEN el archivo `api/openapi/evaluations-and-9x9.yaml`
- WHEN se agrega `CompetencyResultItem` a `components.schemas` con campos: `id` (string, uuid), `name` (string), `profileName` (string), `selfRatingAvg` (number|null), `rhRatingAvg` (number|null), `status` (string)
- AND se referencia correctamente desde la respuesta de `GET /evaluations/competency-results`
- THEN la especificación OpenAPI pasa validación (`openapi-spec-validator` o `swagger-codegen`)
- AND la generación de tipos TypeScript produce `CompetencyResultItem` idéntico al DTO Go

---

## Gap 2: Changes con artefactos incompletos en `openspec/changes/`

### Contexto

Dos changes en `openspec/changes/` no tienen el set completo de artefactos requeridos por el flujo OpenSpec (proposal → tasks → spec → implementación).

### Gap 2a: `oauth-auth` — solo `exploration.md`

| Artefacto | Estado | Notas |
|-----------|--------|-------|
| `exploration.md` | ✅ Existe | Análisis completo de opciones, recomendación OIDC |
| `proposal.md` | ❌ Falta | Decisión de proveedor (Keycloak/Azure/Google), estrategia de provisioning |
| `tasks.md` | ❌ Falta | Lista de tareas de implementación |
| `spec.md` | ❌ Falta | Spec duradera en `openspec/specs/` |

**Impacto**: No hay decisión aprobada ni plan de ejecución. El cambio está estancado en exploración.

### Gap 2b: `add-pillar-type` — `proposal.md` + `tasks.md` sin `spec.md`

| Artefacto | Estado | Notas |
|-----------|--------|-------|
| `proposal.md` | ✅ Existe | Alcance, validación, precondiciones definidas |
| `tasks.md` | ✅ Existe | 8 tareas, varias marcadas como completadas |
| `spec.md` | ❌ Falta | No hay spec duradera en `openspec/specs/` |

**Impacto**: Aunque la implementación parece avanzada (tasks marcadas), no hay spec canónica que documente el comportamiento final para referencia futura, onboarding, o validación de regresiones.

### Requisito de corrección

El sistema SHALL tener un `spec.md` en `openspec/specs/` para cada change que alcance implementación, y un `proposal.md` + `tasks.md` para cada change en exploración activa.

#### Scenario: oauth-auth tiene proposal y tasks

- GIVEN el change `oauth-auth` en `openspec/changes/`
- WHEN se crea `proposal.md` con: proveedor OIDC elegido, estrategia de provisioning (match by email), preservación de dev-login, decisión single-provider
- AND se crea `tasks.md` con tareas concretas (OIDCAdapter, handlers, routes, OpenAPI, frontend, wiring, deps, env)
- THEN el change puede moverse a implementación con `/opsx:apply`

#### Scenario: add-pillar-type tiene spec duradera

- GIVEN el change `add-pillar-type` con implementación completada
- WHEN se crea `openspec/specs/pillar-type/spec.md` (o similar) documentando: modelo de datos, endpoints, DTOs, reglas de validación, comportamiento frontend
- AND el change se archiva con `/opsx:archive`
- THEN la spec sirve como referencia canónica y el change sale de `openspec/changes/`

---

## Gap 3: Rutas OpenAPI sin handlers Go correspondientes

### Contexto

El archivo `api/openapi/evaluations-and-9x9.yaml` define endpoints que **no tienen handler implementado** en `api/internal/handler/evaluation/`.

### Endpoints faltantes

| Endpoint OpenAPI | OperationId | Handler Go | Estado |
|------------------|-------------|------------|--------|
| `GET /evaluations/competency-results` | `getCompetencyResults` | `GetCompetencyResults` | ✅ Existe en `evaluation_handler.go:487` |
| `GET /evaluations/summary` | `getEvaluationSummary` | `GetEvaluationSummary` | ✅ Existe en `evaluation_handler.go:560` |

**Corrección**: Tras revisión, los dos endpoints **sí tienen handlers**. El gap real es la **falta de definición en OpenAPI** para el endpoint `GET /evaluations/competency-results` (el handler existe pero no está en el YAML).

### Evidencia real

- **Handler**: `GetCompetencyResults` en `api/internal/handler/evaluation/evaluation_handler.go:487`
- **Ruta registrada**: Verificar en `routes.go` si `GET /evaluations/competency-results` está cableado
- **OpenAPI**: **No existe** el path `/evaluations/competency-results` en `evaluations-and-9x9.yaml`

### Impacto

- Contrato API no refleja la API real
- Clientes generados no tienen el endpoint
- Tests de contrato (si existen) fallarían

### Requisito de corrección

El sistema SHALL tener correspondencia 1:1 entre paths en OpenAPI y handlers registrados en Chi.

#### Scenario: Endpoint competency-results documentado en OpenAPI

- GIVEN el handler `GetCompetencyResults` existe y está cableado en `routes.go`
- WHEN se agrega el path `/evaluations/competency-results` con `GET` a `evaluations-and-9x9.yaml`
- AND se define request/response usando `CompetencyResultsResponse` (ya existe en DTOs)
- THEN la especificación OpenAPI cubre el 100% de handlers públicos de evaluación

#### Scenario: Validación continua API↔Handler

- GIVEN cualquier cambio en handlers o routes
- WHEN se ejecuta validación (script o CI)
- THEN se detecta automáticamente: handlers sin path OpenAPI, paths OpenAPI sin handler
- AND el pipeline falla si hay desincronización

---

## Requirements

### Requirement: Integridad del contrato OpenAPI

El archivo `api/openapi/evaluations-and-9x9.yaml` SHALL ser un contrato válido OpenAPI 3.1 sin referencias rotas, y SHALL cubrir todos los endpoints públicos expuestos por handlers Go.

#### Scenario: Validación de especificación completa

- GIVEN el archivo `api/openapi/evaluations-and-9x9.yaml`
- WHEN se ejecuta `openapi-spec-validator` o equivalente
- THEN no hay errores de referencias no resueltas
- AND todos los `operationId` en paths tienen handler correspondiente en Go
- AND todos los handlers públicos en `evaluation_handler.go` tienen path en el YAML

### Requirement: Completeness de changes OpenSpec

Todo change en `openspec/changes/` que pase de exploración a implementación SHALL tener: `exploration.md` (opcional), `proposal.md`, `tasks.md`, y al completarse, un `spec.md` en `openspec/specs/`.

#### Scenario: Change completo tiene todos los artefactos

- GIVEN un change en `openspec/changes/<name>/`
- WHEN el change está en estado "implementado" o "completado"
- THEN existen: `proposal.md`, `tasks.md`
- AND existe `openspec/specs/<domain>/spec.md` con la especificación duradera

### Requirement: Sincronización API↔Handler

La correspondencia entre paths OpenAPI y handlers Go SHALL ser verificable automáticamente.

#### Scenario: Detección de drift API-Handler

- GIVEN un script de validación (o step en CI)
- WHEN se comparan `operationId` en OpenAPI vs handlers registrados en `routes.go`
- THEN cualquier discrepancia se reporta como error
- AND el pipeline bloquea merge si hay drift

---

## Non-goals

- **Implementar los fixes**: esta spec solo documenta los gaps; la corrección se hace en changes separados.
- **Cambiar el flujo OpenSpec**: solo exige cumplimiento del flujo existente.
- **Validar otros archivos OpenAPI**: scope limitado a `evaluations-and-9x9.yaml` y los dos changes mencionados.
- **Generar código**: no se genera cliente ni server stub aquí.

---

## Acceptance Criteria

| Criterio | Verificación |
|----------|--------------|
| `CompetencyResultItem` definido en OpenAPI | `grep -n "CompetencyResultItem:" api/openapi/evaluations-and-9x9.yaml` retorna línea en `components.schemas` |
| `oauth-auth` tiene `proposal.md` y `tasks.md` | `ls openspec/changes/oauth-auth/` muestra 3+ archivos |
| `add-pillar-type` tiene `spec.md` en `openspec/specs/` | `ls openspec/specs/` incluye directorio con `spec.md` para pillar-type |
| Endpoint `/evaluations/competency-results` en OpenAPI | `grep -A5 "competency-results" api/openapi/evaluations-and-9x9.yaml` muestra path GET |
| Validación OpenAPI pasa | `openapi-spec-validator api/openapi/evaluations-and-9x9.yaml` exit code 0 |