# Add Pillar Type

## Goal

Extend the `pillars` table with a `type` column (`competencias` | `metas`) so pillars can be shared between the competency framework and the goal-setting module without duplicating tables or forms.

## Non-goals

- Moving competencies or metas to a separate table.
- Changing the competency CRUD flow.
- Adding a new route/view for goal pillars (reuses existing pilares page).

## Scope

| Layer | Change |
|-------|--------|
| DB | New column `type` (NOT NULL, default `competencias`, CHECK constraint), index on `type` |
| Ent schema | `Pillar.Fields()` adds `Enum("type").Values("competencias", "metas").Default("competencias")` |
| Migration | `000020_add_pillar_type` |
| DTOs | `PillarListItem`, `PillarDetail`, `CreatePillarRequest`, `UpdatePillarRequest` gain `Type` field |
| Service | `ListOptions.Type *string`; `List` filters by type; `Create`/`Update` set type |
| Repo | Query filter by `pillar.TypeEQ(type)` when set; rejects type change with competencies |
| Handler | `GET /pillars?type=competencias\|metas` (optional param, omit = all) |
| OpenAPI | Add `type` query param to `GET /pillars`, add `type` to schemas |
| Frontend types | `Pillar.type: 'competencias' \| 'metas'` |
| Frontend store | `load(type?)` sends `?type=...`; omits `include=competencies` for metas |
| Frontend page | `selectedType` triggers refetch via `$effect`; `handleNew()` passes current tab type |

## Validation rules

- Invalid `type` param → 400.
- Changing type from `competencias` to `metas` when `competency_count > 0` → 409 `PILLAR_HAS_COMPETENCIES`.
- `GET /pillars?type=metas&include=competencies` → competencies ignored (empty list).

## Precondition

Existing pillars have no `type` column. Migration defaults all to `competencias`.

## Status

- [x] Ent schema + regen
- [x] Migration 000020
- [x] DTOs
- [x] Service + repo (filter, validation on type change)
- [x] Handler (?type= param)
- [x] OpenAPI spec
- [x] Frontend types, store, page
