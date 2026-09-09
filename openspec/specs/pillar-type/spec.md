# pillar-type Specification

## Purpose
Discriminate pillars by type so competencias and metas are stored and filtered separately.

## Requirements

### Requirement: Pillars SHALL carry a type discriminating competencias from metas

The `pillars` table and `Pillar` Ent schema SHALL store `type` (`competencias`|`metas`, NOT NULL, default `competencias`, CHECK-constrained, indexed). `GET /pillars?type=...` SHALL filter by type (omit = all); DTOs `PillarListItem`, `PillarDetail`, `CreatePillarRequest`, `UpdatePillarRequest` SHALL include `Type`. Changing type from `competencias` to `metas` while `competency_count > 0` SHALL return 409 `PILLAR_HAS_COMPETENCIES`; an invalid `type` param SHALL return 400.

#### Scenario: Filter metas pillars

- WHEN a client calls `GET /pillars?type=metas`
- THEN only pillars with `type = 'metas'` are returned, without competencies included

#### Scenario: Blocked type change with competencies

- WHEN updating a `competencias` pillar that has competencies to `type = 'metas'`
- THEN the API returns 409 `PILLAR_HAS_COMPETENCIES` and the type is unchanged
