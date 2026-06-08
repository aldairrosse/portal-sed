# Proposal: data-model-core

## Intent

Define the PostgreSQL data model via Ent ORM for all SED entities needed by domain specs B1–B5 and future APIs C2–C6. This change produces the schema files, migrations, and index definitions that underpin the entire backend. It is a **pure data model** change — no UI, no API handlers, no auth, no fixtures.

**Decisiones reflejadas:** #1 (no ponderación en pilares), #2 (catálogo único de pilares/competencias para todos los perfiles), #3 (medio año: editar metas, prohibido eliminar), #4 (fin de año: 9×9 del jefe independiente de evaluación RH), #5 (categorías de metas custom independientes de pilares), #6 (KPIs N:M con metas), #7 (RH también tiene metas), #8 (jefe puede ver y solicitar cambios, no agregar/borrar metas).

## Scope

### In Scope

- Ent schema files for ALL entities from B1–B5:
  - **Org hierarchy**: `Organization`, `OrgNode`, `Employee`, `EvaluatorScope`
  - **Evaluation lifecycle**: `Cycle`, `PhaseDefinition`, `PhaseTransition`
  - **Competency framework**: `Pillar`, `Competency`, `ScaleCriterion`, `LevelDefinition`, `CompetencyAcceptanceLevel`, `EvaluationProfile`
  - **Goals and weighting**: `GoalCategory`, `Goal`, `KPI`, `GoalKpiLink`, `GoalAssignment`
  - **9×9 matrix**: `NineBoxMatrix`, `NineBoxEntry`, `NineBoxQuadrant`, `NineBoxScale`
  - **Evaluation**: `Evaluation`, `EvaluationCompetency` (linking employee to competencies for a cycle)
- Edge definitions (Ent relations) between all entities
- Database indices aligned to listing queries for C2–C6
- Versioned migrations (Ent generate + Atlas)
- Audit fields (`created_at`, `updated_at`, `created_by`, `updated_by`) on evaluation and goal entities
- Enums for: `Phase` (`asignacion`, `avance`, `cierre`), `GoalUnit` (`porcentaje`, `moneda`, `numero`), `GoalState`, `EvaluationState`, `ProfileType`
- Constraints at DB level: `targetValue > 0`, weights ∈ [0, 100], unique constraints

### Out of Scope

- **UI** of any kind — pure data model
- **API handlers** — routes and controllers are C2–C6
- **Authentication / RBAC** — session, login, permissions are C7
- **Fixtures / seed data** — Phase A mocks are separate
- **HTTP middleware**, rate limiting, validation at HTTP layer
- **Email / notifications** — async notification infrastructure
- **Audit log table** beyond basic `created_by`/`updated_by` fields
- **Soft delete** unless explicitly required by a spec

## Capabilities

This change introduces the following **new capabilities** (each becomes a spec in `openspec/specs/` after archive):

| Capability | Description |
|------------|-------------|
| `org-hierarchy-data` | PostgreSQL schema for org tree, employees, evaluator scopes |
| `evaluation-lifecycle-data` | PostgreSQL schema for cycles, phases, transitions |
| `competency-framework-data` | PostgreSQL schema for pillars, competencies, scale criteria, acceptance levels |
| `goals-and-weighting-data` | PostgreSQL schema for goal categories, goals, KPIs (N:M), assignments |
| `manager-9x9-data` | PostgreSQL schema for nine-box matrices, entries, quadrants, scales |

## Approach

1. **Ent schema files** in `api/internal/schema/` — one file per entity, with `entc` annotations for edges, indices, and constraints
2. **Edges** defined using Ent's `edge.Resolver` pattern — explicit relation names, cascade delete where domain requires
3. **Indices** aligned to the listing queries defined in `openspec/config.yaml` and the future C2–C6 API specs
4. **Migrations** generated via `ent generate` + Atlas (or `golang-migrate` for apply), stored in `api/migrations/`
5. **Enums** as PostgreSQL `enum` types via Ent schema annotations
6. **JSONB** for flexible metadata fields (e.g., `PhaseDefinition.allowedActions`, `EvaluatorScope.scopeData`) — with GIN index when query-heavy

### Key decisions

| Topic | Decision | References |
|-------|----------|------------|
| `Organization` vs `OrgNode` | OrgNode is the node in the tree; Organization is the root/tentant-level entity | B4 spec |
| `Employee.managerId` | Self-referential FK for hierarchy traversal | B4 spec |
| `GoalCategory` owned by `Employee` | One employee owns the category; no shared categories | B5, #8 |
| `GoalKpiLink` | Explicit join table, not embedded JSONB — supports queries like "all goals using KPI X" | C4 future API |
| `NineBoxEntry.quadrant` | Stored denormalized (computed on write, not read) — avoids recalculating on every read | C6 |
| Cascade delete | `Competency` → cascades to `ScaleCriterion`, `CompetencyAcceptanceLevel` | B2 spec |
| Audit fields | `created_at`, `updated_at` on all evaluation-domain entities; `created_by`, `updated_by` on `Goal`, `Evaluation` | principles/data-and-orm.md |
| Indices | Composite indices per listing query pattern; partial indices where phase-specific | C2–C6 API specs |

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `api/internal/schema/` | New | Ent schema files (one per entity) |
| `api/internal/schema/edges/` | New | Explicit edge definitions |
| `api/migrations/` | New | Versioned SQL migrations |
| `api/internal/` | Modified | Package structure for future handlers/services/repos |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Ent generate conflicts with existing types | Low | No existing Ent files yet; clean slate |
| Index mismatch with future C2–C6 APIs | Medium | Align indices to the listing query patterns documented in config.yaml and each spec; review before C2–C6 |
| Cascade delete unintended data loss | Low | Explicit `edge.Annotations{Ref: "...", OnDelete: cascade}` per relation; review with domain team |
| Enum migration ordering | Medium | Run `CREATE TYPE` before `CREATE TABLE` referencing them; Atlas handles this |

## Rollback Plan

1. Revert migration files in `api/migrations/` to previous version
2. Run `migrate down` to rollback last migration
3. If in production: Point-in-time recovery (PITR) using PostgreSQL WAL
4. Delete the newly created schema files (`api/internal/schema/*.go`)

## Dependencies

- **B1–B5** domain specs must be finalized before this change applies (they are)
- **principles/data-and-orm.md** — index naming and ORM conventions
- **principles/architecture.md** — package structure `api/internal/`
- **Go 1.22+**, `ent`, `Atlas` or `golang-migrate` installed in dev environment

## Success Criteria

- [ ] All 20+ Ent schema files exist in `api/internal/schema/`
- [ ] Every entity has correct edges pointing to its neighbors
- [ ] Composite indices exist for all documented listing queries
- [ ] `ent generate` completes without errors
- [ ] `atlas migrate diff` (or equivalent) produces a valid SQL migration file
- [ ] Audit fields (`created_at`, `updated_at`, `created_by`, `updated_by`) present on all evaluation-domain tables
- [ ] Enum types for `Phase`, `GoalUnit`, `GoalState`, `EvaluationState`, `ProfileType` created
- [ ] No FK cycles that would prevent PostgreSQL from inferring a deterministic delete order
- [ ] `openspec validate --all` passes