# Design: C3 — competency-framework-api

## 1. Overview

C3 delivers the **REST API** for managing the competency catalog — pillars, competencies, scale criteria, level definitions, competency acceptance levels per profile, and evaluation profiles. This is the HR administration API consumed by the RH admin UI (screen A2) and feeds the annual assignment (A3) and formal evaluation (C6) modules.

**Deliverables:**
- 16 HTTP endpoints under `/api/v1/` implemented in Go with Chi.
- Handler, service, and repository layers for `Pillar`, `Competency`, `ScaleCriterion`, `CompetencyAcceptanceLevel`, `LevelDefinition`, and `EvaluationProfile`.
- OpenAPI 3.1 specification (YAML) generated and validated.
- Middleware for rate limiting, idempotency, optimistic locking, and read-replica routing.
- Redis caching strategy for static catalogs and the competency tree.
- Comprehensive test suite: unit, concurrency, load, idempotency, and cache ETag.

**Dependencies:**
- **C1** (`data-model-core`) — tables and indexes must exist.
- **B2** (`competency-framework`) — domain decisions (no pillar weight, single catalog, independent goal categories).
- **C7** (`auth-rbac`) — referenced via `TODO(auth:C7)` markers; no auth implementation in this change.

---

## 2. Package Structure

```
api/internal/
├── handler/competency/
│   ├── pillar_handler.go
│   ├── competency_handler.go
│   ├── scale_handler.go
│   ├── catalog_handler.go
│   ├── acceptance_handler.go
│   ├── routes.go
│   └── handler_test.go          // table-driven tests for all 16 endpoints
├── service/competency/
│   ├── pillar_service.go
│   ├── competency_service.go
│   ├── scale_service.go
│   ├── catalog_service.go
│   ├── acceptance_service.go
│   ├── interfaces.go            // service contracts (interfaces)
│   └── service_test.go
├── repository/competency/
│   ├── pillar_repo.go
│   ├── competency_repo.go
│   ├── scale_repo.go
│   ├── catalog_repo.go
│   ├── acceptance_repo.go
│   ├── interfaces.go            // repository contracts
│   └── repo_test.go
├── middleware/
│   ├── rate_limiter.go
│   ├── idempotency.go
│   ├── optimistic_lock.go
│   ├── read_replica.go
│   └── timeout.go
└── pkg/
    ├── cursor/
    │   ├── cursor.go            // cursor encoding/decoding (base64 json)
    │   └── cursor_test.go
    ├── etag/
    │   ├── etag.go              // hash generation for ETag
    │   └── etag_test.go
    └── errors/
        ├── app_error.go         // typed application errors with codes
        └── errors_test.go
```

---

## 3. Handler Layer

All handlers are in package `handler/competency`. Each handler accepts `http.ResponseWriter` and `*http.Request` (Chi pattern). DTOs are defined in `api/internal/dto/competency/`.

### 3.1 DTOs

```go
// api/internal/dto/competency/pillar_dto.go

type PillarListItem struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	CompetencyCount int    `json:"competency_count"`
}

type PillarDetail struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Competencies []CompetencyLite  `json:"competencies,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type CreatePillarRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty" validate:"max=2000"`
}

type UpdatePillarRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty" validate:"max=2000"`
}

type CompetencyLite struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type CompetencyDetail struct {
	ID            string            `json:"id"`
	PillarID      string            `json:"pillar_id"`
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	ScaleCriteria map[int][]string  `json:"scale_criteria,omitempty"` // level -> []descriptions
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type CreateCompetencyRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty" validate:"max=2000"`
}

type UpdateCompetencyRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty" validate:"max=2000"`
	PillarID    string `json:"pillar_id,omitempty" validate:"omitempty,uuid"`
}

type ScaleCriterionItem struct {
	Level       int    `json:"level" validate:"required,min=1,max=5"`
	Description string `json:"description" validate:"required,max=2000"`
}

type ScaleCriteriaBulkRequest struct {
	Criteria []ScaleCriterionItem `json:"criteria" validate:"required,dive"`
}

type LevelDefinitionItem struct {
	Level       int    `json:"level"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type EvaluationProfileItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type AcceptanceLevelItem struct {
	ID           string `json:"id"`
	CompetencyID string `json:"competency_id"`
	ProfileID    string `json:"profile_id"`
	Level        int    `json:"level"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UpsertAcceptanceRequest struct {
	CompetencyID string `json:"competency_id" validate:"required,uuid"`
	ProfileID    string `json:"profile_id" validate:"required,uuid"`
	Level        int    `json:"level" validate:"required,min=1,max=5"`
}

type ErrorResponse struct {
	Error struct {
		Code    string   `json:"code"`
		Message string   `json:"message"`
		Details []string `json:"details,omitempty"`
		TraceID string   `json:"trace_id,omitempty"`
	} `json:"error"`
}
```

### 3.2 Endpoints

| # | Method | Path | Handler Function | Auth Marker |
|---|--------|------|------------------|-------------|
| 1 | `GET` | `/api/v1/pillars` | `ListPillars` | `TODO(auth:C7)` — any authenticated org user |
| 2 | `POST` | `/api/v1/pillars` | `CreatePillar` | `TODO(auth:C7)` — `rh` or `admin` |
| 3 | `GET` | `/api/v1/pillars/:id` | `GetPillar` | `TODO(auth:C7)` — any authenticated org user |
| 4 | `PUT` | `/api/v1/pillars/:id` | `UpdatePillar` | `TODO(auth:C7)` — `rh` or `admin` |
| 5 | `DELETE` | `/api/v1/pillars/:id` | `DeletePillar` | `TODO(auth:C7)` — `rh` or `admin` |
| 6 | `GET` | `/api/v1/pillars/:pillarId/competencies` | `ListCompetenciesByPillar` | `TODO(auth:C7)` — any authenticated org user |
| 7 | `POST` | `/api/v1/pillars/:pillarId/competencies` | `CreateCompetency` | `TODO(auth:C7)` — `rh` or `admin` |
| 8 | `GET` | `/api/v1/competencies/:id` | `GetCompetency` | `TODO(auth:C7)` — any authenticated org user |
| 9 | `PUT` | `/api/v1/competencies/:id` | `UpdateCompetency` | `TODO(auth:C7)` — `rh` or `admin` |
| 10 | `DELETE` | `/api/v1/competencies/:id` | `DeleteCompetency` | `TODO(auth:C7)` — `rh` or `admin` |
| 11 | `GET` | `/api/v1/competencies/:id/scale-criteria` | `GetScaleCriteria` | `TODO(auth:C7)` — any authenticated org user |
| 12 | `POST` | `/api/v1/competencies/:id/scale-criteria` | `UpsertScaleCriteria` | `TODO(auth:C7)` — `rh` or `admin` |
| 13 | `GET` | `/api/v1/levels` | `ListLevels` | `TODO(auth:C7)` — any authenticated user |
| 14 | `GET` | `/api/v1/acceptance-levels` | `ListAcceptanceLevels` | `TODO(auth:C7)` — any authenticated org user |
| 15 | `POST` | `/api/v1/acceptance-levels` | `UpsertAcceptanceLevel` | `TODO(auth:C7)` — `rh` or `admin` |
| 16 | `GET` | `/api/v1/profiles` | `ListProfiles` | `TODO(auth:C7)` — any authenticated user |

### 3.3 Handler Details

#### `GET /api/v1/pillars` — `ListPillars(w, r)`

**Query Parameters:**
- `include` (string, optional): `"competencies"` — embeds competencies array.
- `cursor` (string, optional): base64-encoded JSON cursor.
- `limit` (int, optional): default `20`, max `100`, min `1`.

**Request:** None.

**Response `200 OK`:**
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Liderazgo",
      "description": "Competencias de liderazgo",
      "competency_count": 5
    }
  ],
  "pagination": {
    "next_cursor": "eyJpZCI6Inh4eCJ9",
    "has_more": true
  }
}
```

**Validation Rules:**
- `limit` parsed as integer; if > 100 clamped to 100; if < 1 clamped to 1.
- `include` validated against allowlist `{"competencies"}`.

**Error Mapping:**
| Condition | HTTP | Error Code |
|-----------|------|------------|
| Invalid `limit` | `400` | `INVALID_PARAMETER` |
| Invalid `cursor` | `400` | `INVALID_CURSOR` |
| Timeout | `408` | `REQUEST_TIMEOUT` |

---

#### `POST /api/v1/pillars` — `CreatePillar(w, r)`

**Headers:**
- `Idempotency-Key: <uuid>` (required for idempotency).

**Request Body:** `CreatePillarRequest`

**Response `201 Created`:** `PillarDetail` (without competencies).

**Validation Rules:**
- `name` required, 1-255 chars.
- `description` max 2000 chars.
- `name` unique in `Pillar` (DB unique constraint + app-level check).

**Error Mapping:**
| Condition | HTTP | Error Code |
|-----------|------|------------|
| Invalid JSON | `400` | `INVALID_REQUEST` |
| Validation fails | `400` | `VALIDATION_ERROR` |
| Duplicate name | `409` | `DUPLICATE_NAME` |
| Idempotency key reused with different payload | `409` | `IDEMPOTENCY_KEY_CONFLICT` |
| Rate limit | `429` | `RATE_LIMIT_EXCEEDED` |

---

#### `GET /api/v1/pillars/:id` — `GetPillar(w, r)`

**Path Parameter:** `id` (UUID).

**Query Parameters:** `include=competencies` (optional).

**Response `200 OK`:** `PillarDetail`.

**Error Mapping:**
| Condition | HTTP | Error Code |
|-----------|------|------------|
| Invalid UUID | `400` | `INVALID_PARAMETER` |
| Pillar not found | `404` | `PILLAR_NOT_FOUND` |

---

#### `PUT /api/v1/pillars/:id` — `UpdatePillar(w, r)`

**Headers:**
- `If-Match: <updated_at>` (required, optimistic lock token).

**Path Parameter:** `id` (UUID).

**Request Body:** `UpdatePillarRequest`

**Response `200 OK`:** `PillarDetail`.

**Validation Rules:**
- Same as `CreatePillar`.
- `If-Match` must match current `updated_at` of the row.

**Error Mapping:**
| Condition | HTTP | Error Code |
|-----------|------|------------|
| Missing `If-Match` | `428` | `PRECONDITION_REQUIRED` |
| Optimistic lock conflict | `409` | `CONCURRENT_UPDATE` |
| Pillar not found | `404` | `PILLAR_NOT_FOUND` |
| Duplicate name | `409` | `DUPLICATE_NAME` |

---

#### `DELETE /api/v1/pillars/:id` — `DeletePillar(w, r)`

**Path Parameter:** `id` (UUID).

**Query Parameter:** `force` (bool, default `false`). If `false`, the pillar must be empty (no competencies). If `true`, performs cascade delete.

**Response `204 No Content` on success.**

**Error Mapping:**
| Condition | HTTP | Error Code |
|-----------|------|------------|
| Pillar not found | `404` | `PILLAR_NOT_FOUND` |
| Pillar has competencies and `force=false` | `409` | `PILLAR_HAS_COMPETENCIES` |
| Advisory lock conflict / concurrent delete | `409` | `CONCURRENT_DELETE` |

---

#### `GET /api/v1/pillars/:pillarId/competencies` — `ListCompetenciesByPillar(w, r)`

**Path Parameter:** `pillarId` (UUID).

**Query Parameters:** `cursor`, `limit` (default 20, max 100).

**Response `200 OK`:**
```json
{
  "data": [
    {
      "id": "...",
      "name": "Comunicación efectiva",
      "description": "..."
    }
  ],
  "pagination": { "next_cursor": "...", "has_more": false }
}
```

**Error Mapping:**
| Condition | HTTP | Error Code |
|-----------|------|------------|
| Invalid UUID | `400` | `INVALID_PARAMETER` |
| Pillar not found | `404` | `PILLAR_NOT_FOUND` |

---

#### `POST /api/v1/pillars/:pillarId/competencies` — `CreateCompetency(w, r)`

**Headers:** `Idempotency-Key: <uuid>` (required).

**Path Parameter:** `pillarId` (UUID).

**Request Body:** `CreateCompetencyRequest`

**Response `201 Created`:** `CompetencyDetail` (without scale criteria).

**Error Mapping:**
| Condition | HTTP | Error Code |
|-----------|------|------------|
| Pillar not found | `404` | `PILLAR_NOT_FOUND` |
| Duplicate name | `409` | `DUPLICATE_NAME` |
| Idempotency key conflict | `409` | `IDEMPOTENCY_KEY_CONFLICT` |

---

#### `GET /api/v1/competencies/:id` — `GetCompetency(w, r)`

**Path Parameter:** `id` (UUID).

**Response `200 OK`:** `CompetencyDetail` with `scale_criteria` grouped by level.

**Error Mapping:**
| Condition | HTTP | Error Code |
|-----------|------|------------|
| Invalid UUID | `400` | `INVALID_PARAMETER` |
| Competency not found | `404` | `COMPETENCY_NOT_FOUND` |

---

#### `PUT /api/v1/competencies/:id` — `UpdateCompetency(w, r)`

**Headers:** `If-Match: <updated_at>` (required).

**Path Parameter:** `id` (UUID).

**Request Body:** `UpdateCompetencyRequest`

**Response `200 OK`:** `CompetencyDetail`.

**Validation Rules:**
- If `pillar_id` is provided, new pillar must exist.

**Error Mapping:**
| Condition | HTTP | Error Code |
|-----------|------|------------|
| Missing `If-Match` | `428` | `PRECONDITION_REQUIRED` |
| Competency not found | `404` | `COMPETENCY_NOT_FOUND` |
| Target pillar not found | `404` | `PILLAR_NOT_FOUND` |
| Concurrent update | `409` | `CONCURRENT_UPDATE` |
| Duplicate name | `409` | `DUPLICATE_NAME` |

---

#### `DELETE /api/v1/competencies/:id` — `DeleteCompetency(w, r)`

**Path Parameter:** `id` (UUID).

**Query Parameter:** `force` (bool, default `false`). If `false`, competency must have no scale criteria.

**Response `204 No Content` on success.**

**Error Mapping:**
| Condition | HTTP | Error Code |
|-----------|------|------------|
| Competency not found | `404` | `COMPETENCY_NOT_FOUND` |
| Has scale criteria and `force=false` | `409` | `COMPETENCY_HAS_CRITERIA` |

---

#### `GET /api/v1/competencies/:id/scale-criteria` — `GetScaleCriteria(w, r)`

**Path Parameter:** `id` (UUID).

**Response `200 OK`:**
```json
{
  "competency_id": "...",
  "criteria": {
    "1": ["No cumple con los requisitos mínimos"],
    "2": ["Requiere supervisión constante"],
    "3": ["Cumple con las expectativas"],
    "4": ["Supera las expectativas"],
    "5": ["Desempeño excepcional"]
  },
  "version": 3
}
```

**Error Mapping:**
| Condition | HTTP | Error Code |
|-----------|------|------------|
| Competency not found | `404` | `COMPETENCY_NOT_FOUND` |

---

#### `POST /api/v1/competencies/:id/scale-criteria` — `UpsertScaleCriteria(w, r)`

**Headers:** `Idempotency-Key: <uuid>` (required).

**Path Parameter:** `id` (UUID).

**Request Body:** `ScaleCriteriaBulkRequest`

**Behavior:**
- Replaces **all** existing scale criteria for the competency with the provided array.
- Validates each `level` is in `1-5`.
- Increments `version` field on the competency (or on each `ScaleCriterion` row).

**Response `200 OK`:**
```json
{
  "competency_id": "...",
  "criteria": { /* grouped by level */ },
  "version": 4,
  "updated_at": "2026-06-05T10:00:00Z"
}
```

**Error Mapping:**
| Condition | HTTP | Error Code |
|-----------|------|------------|
| Invalid level | `400` | `INVALID_LEVEL` |
| Duplicate levels in array | `400` | `DUPLICATE_LEVEL` |
| Competency not found | `404` | `COMPETENCY_NOT_FOUND` |
| Idempotency key conflict | `409` | `IDEMPOTENCY_KEY_CONFLICT` |

---

#### `GET /api/v1/levels` — `ListLevels(w, r)`

**Request:** None.

**Response `200 OK`:** Array of `LevelDefinitionItem`.

**Headers:**
- Response: `ETag: "levels:v1:<hash>"`
- If `If-None-Match` matches, returns `304 Not Modified`.

---

#### `GET /api/v1/acceptance-levels` — `ListAcceptanceLevels(w, r)`

**Query Parameters:**
- `profile_id` (UUID, optional)
- `competency_id` (UUID, optional)

**Response `200 OK`:** Array of `AcceptanceLevelItem`.

**Error Mapping:**
| Condition | HTTP | Error Code |
|-----------|------|------------|
| Invalid UUID in query | `400` | `INVALID_PARAMETER` |

---

#### `POST /api/v1/acceptance-levels` — `UpsertAcceptanceLevel(w, r)`

**Request Body:** `UpsertAcceptanceRequest`

**Behavior:** Upsert on `(competency_id, profile_id)` unique key.

**Response `200 OK`:** `AcceptanceLevelItem` (updated or created).

**Validation Rules:**
- `level` must be 1-5.

**Error Mapping:**
| Condition | HTTP | Error Code |
|-----------|------|------------|
| Invalid level | `400` | `INVALID_LEVEL` |
| Invalid FKs | `404` | `RESOURCE_NOT_FOUND` |

---

#### `GET /api/v1/profiles` — `ListProfiles(w, r)`

**Request:** None.

**Response `200 OK`:** Array of `EvaluationProfileItem`.

**Headers:**
- Response: `ETag: "profiles:v1:<hash>"`
- If `If-None-Match` matches, returns `304 Not Modified`.

---

## 4. Service Layer

### 4.1 Interfaces

```go
// api/internal/service/competency/interfaces.go

package competency

import (
	"context"
	"time"

	"sed-evaluacion-desempeno/api/internal/dto/competency"
	"sed-evaluacion-desempeno/api/internal/ent"
)

// PillarService defines business rules for pillars.
type PillarService interface {
	List(ctx context.Context, opts ListOptions) (*ListResult[competency.PillarListItem], error)
	Get(ctx context.Context, id string, includeCompetencies bool) (*competency.PillarDetail, error)
	Create(ctx context.Context, req competency.CreatePillarRequest) (*competency.PillarDetail, error)
	Update(ctx context.Context, id string, req competency.UpdatePillarRequest, ifMatch time.Time) (*competency.PillarDetail, error)
	Delete(ctx context.Context, id string, force bool) error
}

// CompetencyService defines business rules for competencies.
type CompetencyService interface {
	ListByPillar(ctx context.Context, pillarID string, opts ListOptions) (*ListResult[competency.CompetencyLite], error)
	Get(ctx context.Context, id string) (*competency.CompetencyDetail, error)
	Create(ctx context.Context, pillarID string, req competency.CreateCompetencyRequest) (*competency.CompetencyDetail, error)
	Update(ctx context.Context, id string, req competency.UpdateCompetencyRequest, ifMatch time.Time) (*competency.CompetencyDetail, error)
	Delete(ctx context.Context, id string, force bool) error
}

// ScaleService defines business rules for scale criteria.
type ScaleService interface {
	GetByCompetency(ctx context.Context, competencyID string) (*competency.ScaleCriteriaResponse, error)
	Upsert(ctx context.Context, competencyID string, req competency.ScaleCriteriaBulkRequest) (*competency.ScaleCriteriaResponse, error)
}

// CatalogService defines read-only access to static catalogs.
type CatalogService interface {
	ListLevels(ctx context.Context) ([]competency.LevelDefinitionItem, error)
	ListProfiles(ctx context.Context) ([]competency.EvaluationProfileItem, error)
}

// AcceptanceService defines business rules for competency acceptance levels.
type AcceptanceService interface {
	List(ctx context.Context, filter AcceptanceFilter) ([]competency.AcceptanceLevelItem, error)
	Upsert(ctx context.Context, req competency.UpsertAcceptanceRequest) (*competency.AcceptanceLevelItem, error)
}

// Shared types
type ListOptions struct {
	Cursor  string
	Limit   int
	Include []string
}

type ListResult[T any] struct {
	Data       []T
	NextCursor string
	HasMore    bool
}

type AcceptanceFilter struct {
	ProfileID    *string
	CompetencyID *string
}
```

### 4.2 Implementation Patterns

#### PillarService — Cascade Delete

```go
// api/internal/service/competency/pillar_service.go

func (s *pillarService) Delete(ctx context.Context, id string, force bool) error {
	return s.repo.WithTx(ctx, func(tx *ent.Tx) error {
		// 1. Lock pillar row
		p, err := tx.Pillar.Query().Where(pillar.IDEQ(id)).ForUpdate().Only(ctx)
		if err != nil {
			return appErr.NewNotFound("PILLAR_NOT_FOUND", "pillar not found", err)
		}

		// 2. Count competencies
		compCount, err := p.QueryCompetencies().Count(ctx)
		if err != nil {
			return err
		}

		if compCount > 0 && !force {
			return appErr.NewConflict("PILLAR_HAS_COMPETENCIES",
				"pillar contains competencies", map[string]any{
					"pillar_id":        id,
					"competencies_count": compCount,
					"action_required":  "eliminar competencias primero o usar force=true",
				})
		}

		// 3. Cascade delete (Ent edges with Cascade handle child deletion)
		// Ent will delete competencies -> scale criteria + acceptance levels.
		return tx.Pillar.DeleteOne(p).Exec(ctx)
	})
}
```

#### CompetencyService — Move Between Pillars

```go
func (s *competencyService) Update(ctx context.Context, id string, req competency.UpdateCompetencyRequest, ifMatch time.Time) (*competency.CompetencyDetail, error) {
	return s.repo.WithTx(ctx, func(tx *ent.Tx) error {
		c, err := tx.Competency.Query().Where(competency.IDEQ(id)).ForUpdate().Only(ctx)
		if err != nil {
			return appErr.NewNotFound("COMPETENCY_NOT_FOUND", "competency not found", err)
		}

		// Optimistic lock check
		if c.UpdatedAt != ifMatch {
			return appErr.NewConflict("CONCURRENT_UPDATE", "optimistic lock failed", nil)
		}

		updater := tx.Competency.UpdateOne(c)
		updater.SetName(req.Name)
		updater.SetDescription(req.Description)

		if req.PillarID != "" {
			// Validate target pillar exists
			exists, err := tx.Pillar.Query().Where(pillar.IDEQ(req.PillarID)).Exist(ctx)
			if err != nil || !exists {
				return appErr.NewNotFound("PILLAR_NOT_FOUND", "target pillar not found", err)
			}
			updater.SetPillarID(req.PillarID)
		}

		updated, err := updater.Save(ctx)
		if err != nil {
			return mapEntError(err)
		}
		return s.mapToDetail(ctx, updated)
	})
}
```

#### ScaleService — Bulk Write (Replace All)

```go
func (s *scaleService) Upsert(ctx context.Context, competencyID string, req competency.ScaleCriteriaBulkRequest) (*competency.ScaleCriteriaResponse, error) {
	return s.repo.WithTx(ctx, func(tx *ent.Tx) error {
		// Verify competency exists
		comp, err := tx.Competency.Query().Where(competency.IDEQ(competencyID)).Only(ctx)
		if err != nil {
			return appErr.NewNotFound("COMPETENCY_NOT_FOUND", "competency not found", err)
		}

		// Validate no duplicate levels in request
		seen := make(map[int]struct{})
		for _, item := range req.Criteria {
			if item.Level < 1 || item.Level > 5 {
				return appErr.NewValidation("INVALID_LEVEL", "level must be 1-5", nil)
			}
			if _, ok := seen[item.Level]; ok {
				return appErr.NewValidation("DUPLICATE_LEVEL", "duplicate level in request", nil)
			}
			seen[item.Level] = struct{}{}
		}

		// Bulk delete existing
		_, err = tx.ScaleCriterion.Delete().Where(scalecriterion.CompetencyIDEQ(competencyID)).Exec(ctx)
		if err != nil {
			return err
		}

		// Bulk insert new
		builders := make([]*ent.ScaleCriterionCreate, len(req.Criteria))
		for i, item := range req.Criteria {
			builders[i] = tx.ScaleCriterion.Create().
				SetCompetencyID(competencyID).
				SetPillarID(comp.PillarID).
				SetLevel(item.Level).
				SetDescription(item.Description)
		}
		_, err = tx.ScaleCriterion.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return err
		}

		// Increment version on competency (or maintain row-level version in ScaleCriterion)
		_, err = tx.Competency.UpdateOne(comp).SetUpdatedAt(time.Now()).Save(ctx)
		if err != nil {
			return err
		}

		return s.GetByCompetency(ctx, competencyID)
	})
}
```

#### AcceptanceService — Upsert

```go
func (s *acceptanceService) Upsert(ctx context.Context, req competency.UpsertAcceptanceRequest) (*competency.AcceptanceLevelItem, error) {
	return s.repo.WithTx(ctx, func(tx *ent.Tx) error {
		// Validate FKs
		compExists, err := tx.Competency.Query().Where(competency.IDEQ(req.CompetencyID)).Exist(ctx)
		if err != nil || !compExists {
			return appErr.NewNotFound("COMPETENCY_NOT_FOUND", "competency not found", err)
		}
		profExists, err := tx.EvaluationProfile.Query().Where(evaluationprofile.IDEQ(req.ProfileID)).Exist(ctx)
		if err != nil || !profExists {
			return appErr.NewNotFound("PROFILE_NOT_FOUND", "profile not found", err)
		}

		// Check existing
		existing, err := tx.CompetencyAcceptanceLevel.Query().
			Where(competencyacceptancelevel.CompetencyIDEQ(req.CompetencyID)).
			Where(competencyacceptancelevel.ProfileIDEQ(req.ProfileID)).
			Only(ctx)

		if ent.IsNotFound(err) {
			// Create
			created, err := tx.CompetencyAcceptanceLevel.Create().
				SetCompetencyID(req.CompetencyID).
				SetProfileID(req.ProfileID).
				SetLevel(req.Level).
				Save(ctx)
			if err != nil {
				return mapEntError(err)
			}
			return s.mapToDTO(created), nil
		} else if err != nil {
			return err
		}

		// Update
		updated, err := tx.CompetencyAcceptanceLevel.UpdateOne(existing).SetLevel(req.Level).Save(ctx)
		if err != nil {
			return mapEntError(err)
		}
		return s.mapToDTO(updated), nil
	})
}
```

---

## 5. Repository Layer

### 5.1 Interfaces

```go
// api/internal/repository/competency/interfaces.go

package competency

import (
	"context"
	"time"

	"sed-evaluacion-desempeno/api/internal/ent"
)

type TxFunc func(tx *ent.Tx) error

type PillarRepo interface {
	WithTx(ctx context.Context, fn TxFunc) error
	List(ctx context.Context, cursor string, limit int, includeCompetencies bool) ([]*ent.Pillar, string, error)
	Get(ctx context.Context, id string, includeCompetencies bool) (*ent.Pillar, error)
	Create(ctx context.Context, name, description string) (*ent.Pillar, error)
	Update(ctx context.Context, id string, name, description string, ifMatch time.Time) (*ent.Pillar, error)
	Delete(ctx context.Context, id string) error
	CountCompetencies(ctx context.Context, pillarID string) (int, error)
}

type CompetencyRepo interface {
	WithTx(ctx context.Context, fn TxFunc) error
	ListByPillar(ctx context.Context, pillarID string, cursor string, limit int) ([]*ent.Competency, string, error)
	Get(ctx context.Context, id string) (*ent.Competency, error)
	Create(ctx context.Context, pillarID, name, description string) (*ent.Competency, error)
	Update(ctx context.Context, id string, name, description, pillarID string, ifMatch time.Time) (*ent.Competency, error)
	Delete(ctx context.Context, id string) error
}

type ScaleRepo interface {
	WithTx(ctx context.Context, fn TxFunc) error
	GetByCompetency(ctx context.Context, competencyID string) ([]*ent.ScaleCriterion, error)
	ReplaceAll(ctx context.Context, competencyID string, criteria []*ScaleCriterionInput) error
}

type ScaleCriterionInput struct {
	Level       int
	Description string
}

type CatalogRepo interface {
	ListLevels(ctx context.Context) ([]*ent.LevelDefinition, error)
	ListProfiles(ctx context.Context) ([]*ent.EvaluationProfile, error)
}

type AcceptanceRepo interface {
	WithTx(ctx context.Context, fn TxFunc) error
	List(ctx context.Context, competencyID, profileID *string) ([]*ent.CompetencyAcceptanceLevel, error)
	Upsert(ctx context.Context, competencyID, profileID string, level int) (*ent.CompetencyAcceptanceLevel, error)
}
```

### 5.2 Eager Loading Patterns

```go
// api/internal/repository/competency/pillar_repo.go

func (r *pillarRepo) Get(ctx context.Context, id string, includeCompetencies bool) (*ent.Pillar, error) {
	q := r.client.Pillar.Query().Where(pillar.IDEQ(id))
	if includeCompetencies {
		q = q.WithCompetencies(func(q *ent.CompetencyQuery) {
			q.Order(ent.Asc(competency.FieldName))
		})
	}
	return q.Only(ctx)
}

func (r *pillarRepo) List(ctx context.Context, cursor string, limit int, includeCompetencies bool) ([]*ent.Pillar, string, error) {
	q := r.client.Pillar.Query().
		Order(ent.Asc(pillar.FieldName)).
		Limit(limit + 1) // +1 to detect hasMore

	if includeCompetencies {
		q = q.WithCompetencies()
	}

	if cursor != "" {
		decoded, err := cursorpkg.Decode(cursor)
		if err != nil {
			return nil, "", err
		}
		q = q.Where(pillar.NameGT(decoded.Name))
	}

	results, err := q.All(ctx)
	if err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(results) > limit {
		results = results[:limit]
		last := results[len(results)-1]
		nextCursor = cursorpkg.Encode(cursorpkg.Cursor{Name: last.Name})
	}

	return results, nextCursor, nil
}
```

### 5.3 Read Replica Routing

```go
// api/internal/repository/competency/pillar_repo.go

func (r *pillarRepo) List(ctx context.Context, cursor string, limit int, includeCompetencies bool) ([]*ent.Pillar, string, error) {
	client := r.readClient // pgx pool configured to read replica
	q := client.Pillar.Query().
		Order(ent.Asc(pillar.FieldName)).
		Limit(limit + 1)
	// ...
}
```

---

## 6. Middleware

### 6.1 Rate Limiter

```go
// api/internal/middleware/rate_limiter.go

package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(rdb *redis.Client, writeRPS, readRPS int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isWrite := r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete || r.Method == http.MethodPatch
			key := "rate:" + r.Method + ":" + r.RemoteAddr
			limit := readRPS
			if isWrite {
				limit = writeRPS
			}

			pipe := rdb.Pipeline()
			incr := pipe.Incr(r.Context(), key)
			pipe.Expire(r.Context(), key, time.Second)
			_, err := pipe.Exec(r.Context())
			if err != nil {
				http.Error(w, `{"error":{"code":"RATE_LIMIT_ERROR"}}`, http.StatusInternalServerError)
				return
			}

			if incr.Val() > int64(limit) {
				http.Error(w, `{"error":{"code":"RATE_LIMIT_EXCEEDED"}}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

**Applied to router:**
```go
// api/internal/handler/competency/routes.go

func RegisterRoutes(r chi.Router, deps *Dependencies) {
	r.Use(middleware.Timeout(3 * time.Second)) // GET
	r.Use(middleware.RateLimiter(deps.Redis, 50, 500))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/pillars", func(r chi.Router) {
			r.Get("/", handler.ListPillars)
			r.With(middleware.Idempotency(deps.Redis)).Post("/", handler.CreatePillar)
			r.Get("/{id}", handler.GetPillar)
			r.With(middleware.OptimisticLock).Put("/{id}", handler.UpdatePillar)
			r.Delete("/{id}", handler.DeletePillar)
			r.Get("/{pillarId}/competencies", handler.ListCompetenciesByPillar)
			r.With(middleware.Idempotency(deps.Redis)).Post("/{pillarId}/competencies", handler.CreateCompetency)
		})
		// ... etc
	})
}
```

### 6.2 Idempotency

```go
// api/internal/middleware/idempotency.go

func Idempotency(rdb *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				http.Error(w, `{"error":{"code":"IDEMPOTENCY_KEY_REQUIRED"}}`, http.StatusBadRequest)
				return
			}
			if !isValidUUID(key) {
				http.Error(w, `{"error":{"code":"IDEMPOTENCY_KEY_INVALID"}}`, http.StatusBadRequest)
				return
			}

			redisKey := "idempotency:" + key
			cached, err := rdb.Get(r.Context(), redisKey).Result()
			if err == nil && cached != "" {
				w.Header().Set("X-Idempotency-Replay", "true")
				w.Write([]byte(cached))
				return
			}
			if err != redis.Nil {
				http.Error(w, `{"error":{"code":"INTERNAL_ERROR"}}`, http.StatusInternalServerError)
				return
			}

			// Store response after successful execution
			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, r)

			body := rec.Body.Bytes()
			if rec.Code < 300 {
				rdb.SetEX(r.Context(), redisKey, string(body), 24*time.Hour)
			}
			w.WriteHeader(rec.Code)
			for k, v := range rec.Header() {
				w.Header()[k] = v
			}
			w.Write(body)
		})
	}
}
```

### 6.3 Optimistic Lock

```go
// api/internal/middleware/optimistic_lock.go

func OptimisticLock(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			ifMatch := r.Header.Get("If-Match")
			if ifMatch == "" {
				http.Error(w, `{"error":{"code":"PRECONDITION_REQUIRED"}}`, http.StatusPreconditionRequired)
				return
			}
			if _, err := time.Parse(time.RFC3339Nano, ifMatch); err != nil {
				http.Error(w, `{"error":{"code":"INVALID_PRECONDITION"}}`, http.StatusBadRequest)
				return
			}
			ctx := context.WithValue(r.Context(), "if-match", ifMatch)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		next.ServeHTTP(w, r)
	})
}
```

---

## 7. OpenAPI 3.1 Spec

```yaml
# api/openapi/competency-framework-api.yaml
openapi: 3.1.0
info:
  title: SED Competency Framework API
  version: 1.0.0
  description: REST API for managing the competency catalog (C3).
servers:
  - url: https://api.sed.example.com/api/v1
    description: Production
paths:
  /pillars:
    get:
      operationId: listPillars
      summary: List pillars
      parameters:
        - name: include
          in: query
          schema:
            type: string
            enum: [competencies]
        - name: cursor
          in: query
          schema:
            type: string
        - name: limit
          in: query
          schema:
            type: integer
            default: 20
            maximum: 100
            minimum: 1
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PillarListResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '408':
          $ref: '#/components/responses/RequestTimeout'
    post:
      operationId: createPillar
      summary: Create a new pillar
      parameters:
        - name: Idempotency-Key
          in: header
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreatePillarRequest'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PillarDetail'
        '400':
          $ref: '#/components/responses/BadRequest'
        '409':
          $ref: '#/components/responses/Conflict'
        '429':
          $ref: '#/components/responses/RateLimit'
  /pillars/{id}:
    get:
      operationId: getPillar
      summary: Get pillar by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
        - name: include
          in: query
          schema:
            type: string
            enum: [competencies]
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PillarDetail'
        '404':
          $ref: '#/components/responses/NotFound'
    put:
      operationId: updatePillar
      summary: Update pillar
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
        - name: If-Match
          in: header
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdatePillarRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PillarDetail'
        '404':
          $ref: '#/components/responses/NotFound'
        '409':
          $ref: '#/components/responses/Conflict'
        '428':
          $ref: '#/components/responses/PreconditionRequired'
    delete:
      operationId: deletePillar
      summary: Delete pillar
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
        - name: force
          in: query
          schema:
            type: boolean
            default: false
      responses:
        '204':
          description: No Content
        '404':
          $ref: '#/components/responses/NotFound'
        '409':
          $ref: '#/components/responses/Conflict'
  /pillars/{pillarId}/competencies:
    get:
      operationId: listCompetenciesByPillar
      summary: List competencies in a pillar
      parameters:
        - name: pillarId
          in: path
          required: true
          schema:
            type: string
            format: uuid
        - name: cursor
          in: query
          schema:
            type: string
        - name: limit
          in: query
          schema:
            type: integer
            default: 20
            maximum: 100
            minimum: 1
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CompetencyListResponse'
        '404':
          $ref: '#/components/responses/NotFound'
    post:
      operationId: createCompetency
      summary: Create competency in a pillar
      parameters:
        - name: pillarId
          in: path
          required: true
          schema:
            type: string
            format: uuid
        - name: Idempotency-Key
          in: header
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateCompetencyRequest'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CompetencyDetail'
        '404':
          $ref: '#/components/responses/NotFound'
        '409':
          $ref: '#/components/responses/Conflict'
  /competencies/{id}:
    get:
      operationId: getCompetency
      summary: Get competency detail
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CompetencyDetail'
        '404':
          $ref: '#/components/responses/NotFound'
    put:
      operationId: updateCompetency
      summary: Update competency
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
        - name: If-Match
          in: header
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateCompetencyRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CompetencyDetail'
        '404':
          $ref: '#/components/responses/NotFound'
        '409':
          $ref: '#/components/responses/Conflict'
        '428':
          $ref: '#/components/responses/PreconditionRequired'
    delete:
      operationId: deleteCompetency
      summary: Delete competency
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
        - name: force
          in: query
          schema:
            type: boolean
            default: false
      responses:
        '204':
          description: No Content
        '404':
          $ref: '#/components/responses/NotFound'
        '409':
          $ref: '#/components/responses/Conflict'
  /competencies/{id}/scale-criteria:
    get:
      operationId: getScaleCriteria
      summary: Get scale criteria for a competency
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ScaleCriteriaResponse'
        '404':
          $ref: '#/components/responses/NotFound'
    post:
      operationId: upsertScaleCriteria
      summary: Replace scale criteria (bulk)
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
        - name: Idempotency-Key
          in: header
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ScaleCriteriaBulkRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ScaleCriteriaResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '404':
          $ref: '#/components/responses/NotFound'
        '409':
          $ref: '#/components/responses/Conflict'
  /levels:
    get:
      operationId: listLevels
      summary: List global level definitions
      responses:
        '200':
          description: OK
          headers:
            ETag:
              schema:
                type: string
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/LevelDefinition'
        '304':
          description: Not Modified
  /acceptance-levels:
    get:
      operationId: listAcceptanceLevels
      summary: List acceptance levels
      parameters:
        - name: profile_id
          in: query
          schema:
            type: string
            format: uuid
        - name: competency_id
          in: query
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/AcceptanceLevel'
    post:
      operationId: upsertAcceptanceLevel
      summary: Upsert acceptance level
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpsertAcceptanceRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AcceptanceLevel'
        '400':
          $ref: '#/components/responses/BadRequest'
        '404':
          $ref: '#/components/responses/NotFound'
  /profiles:
    get:
      operationId: listProfiles
      summary: List evaluation profiles
      responses:
        '200':
          description: OK
          headers:
            ETag:
              schema:
                type: string
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/EvaluationProfile'
        '304':
          description: Not Modified

components:
  schemas:
    PillarListItem:
      type: object
      properties:
        id: { type: string, format: uuid }
        name: { type: string }
        description: { type: string }
        competency_count: { type: integer }
    PillarListResponse:
      type: object
      properties:
        data:
          type: array
          items:
            $ref: '#/components/schemas/PillarListItem'
        pagination:
          type: object
          properties:
            next_cursor: { type: string }
            has_more: { type: boolean }
    PillarDetail:
      type: object
      properties:
        id: { type: string, format: uuid }
        name: { type: string }
        description: { type: string }
        competencies:
          type: array
          items:
            $ref: '#/components/schemas/CompetencyLite'
        created_at: { type: string, format: date-time }
        updated_at: { type: string, format: date-time }
    CreatePillarRequest:
      type: object
      required: [name]
      properties:
        name: { type: string, minLength: 1, maxLength: 255 }
        description: { type: string, maxLength: 2000 }
    UpdatePillarRequest:
      type: object
      required: [name]
      properties:
        name: { type: string, minLength: 1, maxLength: 255 }
        description: { type: string, maxLength: 2000 }
    CompetencyLite:
      type: object
      properties:
        id: { type: string, format: uuid }
        name: { type: string }
        description: { type: string }
    CompetencyListResponse:
      type: object
      properties:
        data:
          type: array
          items:
            $ref: '#/components/schemas/CompetencyLite'
        pagination:
          type: object
          properties:
            next_cursor: { type: string }
            has_more: { type: boolean }
    CreateCompetencyRequest:
      type: object
      required: [name]
      properties:
        name: { type: string, minLength: 1, maxLength: 255 }
        description: { type: string, maxLength: 2000 }
    UpdateCompetencyRequest:
      type: object
      required: [name]
      properties:
        name: { type: string, minLength: 1, maxLength: 255 }
        description: { type: string, maxLength: 2000 }
        pillar_id: { type: string, format: uuid }
    CompetencyDetail:
      type: object
      properties:
        id: { type: string, format: uuid }
        pillar_id: { type: string, format: uuid }
        name: { type: string }
        description: { type: string }
        scale_criteria:
          type: object
          additionalProperties:
            type: array
            items: { type: string }
        created_at: { type: string, format: date-time }
        updated_at: { type: string, format: date-time }
    ScaleCriterionItem:
      type: object
      required: [level, description]
      properties:
        level: { type: integer, minimum: 1, maximum: 5 }
        description: { type: string, maxLength: 2000 }
    ScaleCriteriaBulkRequest:
      type: object
      required: [criteria]
      properties:
        criteria:
          type: array
          items:
            $ref: '#/components/schemas/ScaleCriterionItem'
    ScaleCriteriaResponse:
      type: object
      properties:
        competency_id: { type: string, format: uuid }
        criteria:
          type: object
          additionalProperties:
            type: array
            items: { type: string }
        version: { type: integer }
        updated_at: { type: string, format: date-time }
    LevelDefinition:
      type: object
      properties:
        level: { type: integer, minimum: 1, maximum: 5 }
        label: { type: string }
        description: { type: string }
    EvaluationProfile:
      type: object
      properties:
        id: { type: string, format: uuid }
        name: { type: string }
        description: { type: string }
    AcceptanceLevel:
      type: object
      properties:
        id: { type: string, format: uuid }
        competency_id: { type: string, format: uuid }
        profile_id: { type: string, format: uuid }
        level: { type: integer, minimum: 1, maximum: 5 }
        created_at: { type: string, format: date-time }
        updated_at: { type: string, format: date-time }
    UpsertAcceptanceRequest:
      type: object
      required: [competency_id, profile_id, level]
      properties:
        competency_id: { type: string, format: uuid }
        profile_id: { type: string, format: uuid }
        level: { type: integer, minimum: 1, maximum: 5 }
    Error:
      type: object
      properties:
        error:
          type: object
          properties:
            code: { type: string }
            message: { type: string }
            details:
              type: array
              items: { type: string }
            trace_id: { type: string }
  responses:
    BadRequest:
      description: Bad Request
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    NotFound:
      description: Not Found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    Conflict:
      description: Conflict
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    PreconditionRequired:
      description: Precondition Required
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    RateLimit:
      description: Too Many Requests
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    RequestTimeout:
      description: Request Timeout
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
```

---

## 8. Concurrency Details

### 8.1 Database Connection Management

- **pgx connection pool:**
  ```go
  config, _ := pgxpool.ParseConfig(connStr)
  config.MaxConns = 25
  config.MinConns = 5
  config.MaxConnLifetime = 1 * time.Hour
  config.MaxConnIdleTime = 30 * time.Minute
  pool, _ := pgxpool.NewWithConfig(ctx, config)
  ```
- The `EntClient` is injected per request via `chi` context. Writes use `ent.Tx` with explicit isolation.
- **Read replicas:** GET endpoints use a separate `ent.Client` backed by a read-replica pool. Router logic: `dbRole = read` selects replica; `dbRole = write` selects primary.
- **Metrics:** Prometheus counters `pgx_pool_conns_busy`, `pgx_pool_conns_idle`, `pgx_pool_wait_duration_ms` exposed at `/metrics`.

### 8.2 Optimistic Locking

- `Pillar` and `Competency` tables have `updated_at` (timestamp with time zone).
- `PUT` requests must include `If-Match: <updated_at>` where `<updated_at>` is `RFC3339Nano`.
- Repository checks `updated_at` equality in the `UPDATE` `WHERE` clause:
  ```sql
  UPDATE pillars SET name = $1, description = $2, updated_at = now()
  WHERE id = $3 AND updated_at = $4
  ```
- If `RowsAffected == 0`, repository returns `CONCURRENT_UPDATE`.

### 8.3 Advisory Lock on Pillar Deletion

```go
func (r *pillarRepo) Delete(ctx context.Context, id string) error {
	return r.WithTx(ctx, func(tx *ent.Tx) error {
		// Advisory lock: hash pillar ID to int64
		lockID := hashUUIDToInt64(id)
		_, err := tx.ExecContext(ctx, "SELECT pg_advisory_lock($1)", lockID)
		if err != nil {
			return err
		}
		defer tx.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", lockID)

		// ... delete logic
		return tx.Pillar.DeleteOneID(id).Exec(ctx)
	})
}
```

### 8.4 Version Field on ScaleCriterion

- `ScaleCriterion` table includes `version integer default 1`.
- On `POST /competencies/:id/scale-criteria`, the service increments `version` on the competency row (or on each criterion row if per-row versioning is desired).
- This allows concurrent editors to detect drift: if a second `PUT` arrives with an older `version`, return `CONCURRENT_UPDATE`.

### 8.5 Transaction Design

#### Pillar Delete Cascade
```sql
BEGIN;
  SELECT pg_advisory_lock(hash('pillar_id'));
  SELECT * FROM pillars WHERE id = 'pillar_id' FOR UPDATE;
  -- validate competencies count (app-level)
  DELETE FROM scale_criteria WHERE competency_id IN (
    SELECT id FROM competencies WHERE pillar_id = 'pillar_id'
  );
  DELETE FROM competency_acceptance_levels WHERE competency_id IN (
    SELECT id FROM competencies WHERE pillar_id = 'pillar_id'
  );
  DELETE FROM competencies WHERE pillar_id = 'pillar_id';
  DELETE FROM pillars WHERE id = 'pillar_id';
COMMIT;
```

#### Competency Update + Scale Criteria
```sql
BEGIN;
  SELECT * FROM competencies WHERE id = 'comp_id' FOR UPDATE;
  -- update competency fields
  DELETE FROM scale_criteria WHERE competency_id = 'comp_id';
  INSERT INTO scale_criteria (...) VALUES (...);
  UPDATE competencies SET updated_at = now() WHERE id = 'comp_id';
COMMIT;
```

**Isolation Level:** `READ COMMITTED` (default in PostgreSQL). Explicit locking (`FOR UPDATE`, `pg_advisory_lock`) is used for critical paths.

---

## 9. Caching Strategy

### 9.1 Redis Key Patterns

| Key | Data | TTL | Invalidation Trigger |
|-----|------|-----|---------------------|
| `competency:tree:v1` | Full JSON tree (pillars -> competencies -> criteria) | 1h | Any `POST`, `PUT`, `DELETE` on pillars, competencies, or scale criteria |
| `levels:v1` | Array of `LevelDefinition` (5 records) | 24h | Never (static seed data) |
| `profiles:v1` | Array of `EvaluationProfile` (8 records) | 24h | Never (static seed data) |
| `idempotency:<uuid>` | Cached response body | 24h | TTL expiration |
| `rate:<method>:<ip>` | Request counter | 1s | TTL expiration |

### 9.2 ETag Generation

```go
// api/internal/pkg/etag/etag.go

package etag

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func Generate(prefix string, data []byte) string {
	h := sha256.New()
	h.Write(data)
	return fmt.Sprintf("%s:%s", prefix, hex.EncodeToString(h.Sum(nil))[:16])
}
```

- `GET /api/v1/levels`: ETag = `etag.Generate("levels", jsonBytes)`
- `GET /api/v1/profiles`: ETag = `etag.Generate("profiles", jsonBytes)`
- If client sends `If-None-Match` matching the generated ETag, return `304 Not Modified` with no body.

### 9.3 Cache Invalidation

```go
// api/internal/service/competency/cache_invalidator.go

func (i *invalidator) InvalidateTree(ctx context.Context) {
	i.rdb.Del(ctx, "competency:tree:v1")
}

// Called after every successful write:
// - CreatePillar, UpdatePillar, DeletePillar
// - CreateCompetency, UpdateCompetency, DeleteCompetency
// - UpsertScaleCriteria
```

---

## 10. Testing Strategy

### 10.1 Unit Tests

- **Table-driven tests** in `api/internal/handler/competency/handler_test.go` for every endpoint.
- **Mocked services** using `mockgen` or manual fakes.
- **Coverage targets:** handler layer ≥ 80%, service layer ≥ 85%, repository layer ≥ 70%.

Example test structure:
```go
func TestCreatePillar(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		setupMock  func(*mockcompetency.MockPillarService)
		wantStatus int
		wantCode   string
	}{
		{
			name: "success",
			body: `{"name":"Liderazgo"}`,
			setupMock: func(m *mockcompetency.MockPillarService) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&competency.PillarDetail{ID: "...", Name: "Liderazgo"}, nil)
			},
			wantStatus: 201,
		},
		{
			name:       "duplicate name",
			body:       `{"name":"Liderazgo"}`,
			setupMock: func(m *mockcompetency.MockPillarService) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, appErr.NewConflict("DUPLICATE_NAME", "...", nil))
			},
			wantStatus: 409,
			wantCode:   "DUPLICATE_NAME",
		},
		// ...
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// ...
		})
	}
}
```

### 10.2 Concurrency Tests

- **50 goroutines** attempt to `UpsertScaleCriteria` on the same competency simultaneously.
- **Assertion:** final `version` == 50, no data loss, no duplicate levels.
- **Tool:** `go test -race` + `sync.WaitGroup`.

```go
func TestConcurrentScaleCriteriaUpdate(t *testing.T) {
	ctx := context.Background()
	compID := seedCompetency(t)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := competency.ScaleCriteriaBulkRequest{
				Criteria: []competency.ScaleCriterionItem{{Level: 1, Description: fmt.Sprintf("desc-%d", i)}},
			}
			svc.Upsert(ctx, compID, req)
		}(i)
	}
	wg.Wait()

	res, _ := svc.GetByCompetency(ctx, compID)
	assert.Equal(t, 50, res.Version)
}
```

### 10.3 Load Tests

- **Target:** `GET /api/v1/pillars` sustains **500 req/s** for 60s with p95 latency < 200ms.
- **Tool:** `k6` script in `api/tests/load/pillars_list.js`.

```js
// api/tests/load/pillars_list.js
import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 500 },
    { duration: '60s', target: 500 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<200'],
  },
};

export default function () {
  const res = http.get('http://api:8080/api/v1/pillars?limit=20');
  check(res, { 'status is 200': (r) => r.status === 200 });
}
```

### 10.4 Idempotency Tests

- **Test:** Send `POST /api/v1/pillars` with same `Idempotency-Key` twice.
- **Assertion:** Second request returns `201` with identical body and `X-Idempotency-Replay: true`; no duplicate row in DB.

### 10.5 Cache ETag Tests

- **Test:** `GET /api/v1/levels` → capture `ETag`. Second request with `If-None-Match: <ETag>` → `304`.
- **Test:** `GET /api/v1/profiles` → same pattern.

### 10.6 Pool Metrics Tests

- **Test:** Expose `/metrics` endpoint. Assert Prometheus text contains `pgx_pool_conns_busy` and `pgx_pool_conns_idle`.

### 10.7 N+1 Tests

- **Test:** `GET /api/v1/pillars?include=competencies` triggers **exactly 1 or 2 queries** (count + data with eager load), verified via `sqlmock` or `pgx` query logging.
- **Assertion:** No looped queries per pillar.

---

## Success Criteria

- [ ] All 16 endpoints implemented with Chi and validated against OpenAPI 3.1.
- [ ] OpenAPI spec generated; TypeScript types generated in `web/src/lib/api/`.
- [ ] Unit tests cover happy path, validation errors, duplicates, invalid levels, cascade delete.
- [ ] Concurrency test: 50 goroutines update same competency; no data loss, version consistent.
- [ ] Load test: `GET /api/v1/pillars` sustains 500 req/s for 60s with p95 < 200ms.
- [ ] Idempotency test: replay with same key returns identical response without duplicate.
- [ ] ETag cache test: `GET /levels` and `GET /profiles` return `304` on second request.
- [ ] Prometheus pool metrics exposed and verified.
- [ ] No N+1: `?include=competencies` resolves in a single query or eager load.
- [ ] READMEs in `internal/service/competency/` and `internal/handler/competency/` document cascade flow and locking.
