# Design: rh-competency-admin-ui (A2)

## Technical Approach

Four RH admin screens built on the A1 shell, using Svelte 5 runes for local CRUD state backed by synchronous JSON fixture imports. No network calls. Routes under `/rh/` with lazy code-splitting. A single `competencyStore.svelte.ts` holds all entities in `$state` and exposes mutation functions — reload resets to fixture defaults. Types in `competency.ts` are shaped for forward compatibility with B2 (`competency-framework`) and C3 (`competency-framework-api`).

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|-------------|-----------|
| State management | Single `competencyStore.svelte.ts` with `$state` runes | Per-route local state; Svelte stores (legacy) | Centralized CRUD avoids prop-drilling across nested routes; runes are the Svelte 5 pattern established by `devContext.svelte.ts` |
| Fixture loading | Static `import` of JSON at store init | `fetch()` at route load; dynamic `import()` | Zero network overhead; synchronous; matches spec requirement; reload resets state |
| Route nesting | `/rh/pilares/[id]/competencias` as SvelteKit dynamic route | Flat `/rh/competencias?pilar=X` | SvelteKit convention; cleaner URLs; pillar context via params |
| Menu restructure | Replace `/rh/competencias` entry with 3 sidebar items (Pilares, Criterios, Niveles) | Keep single entry with sub-tabs | Spec defines 4 screens; competencies are accessed from pillar list, not sidebar |
| Component granularity | One form modal per entity (Pillar, Competency, ScaleCriterion) | Generic reusable form component | Only 3 forms — abstraction cost exceeds benefit; each has distinct fields |
| Delete confirmation | Shared `ConfirmDeleteModal.svelte` | `window.confirm()`; inline alert | Reusable across all CRUD screens; accessible; consistent UX |

## Data Flow

```
fixtures/*.json ──import──→ competencyStore.svelte.ts ──$derived──→ Route pages
                                   ↑                                      │
                                   └──── mutation fns (add/update/delete) ─┘
```

Store initializes from fixtures on first access. Route pages read via `$derived` getters. User actions call store mutation functions. No persistence — page reload resets to fixture state.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `web/src/lib/types/competency.ts` | Create | `Pillar`, `Competency`, `ScaleCriterion`, `AcceptanceLevel` interfaces |
| `web/src/lib/fixtures/competency/pillars.json` | Create | 3 seed pillars |
| `web/src/lib/fixtures/competency/competencies.json` | Create | 2-3 competencies per pillar |
| `web/src/lib/fixtures/competency/scale-criteria.json` | Create | Level 1-5 descriptions per competency × pillar |
| `web/src/lib/fixtures/competency/acceptance-levels.json` | Create | Labels/descriptions per profile × level |
| `web/src/lib/stores/competencyStore.svelte.ts` | Create | `$state` store with CRUD operations for all entities |
| `web/src/lib/components/competency/PillarTable.svelte` | Create | Table with edit/delete actions |
| `web/src/lib/components/competency/CompetencyTable.svelte` | Create | Table filtered by pillar param |
| `web/src/lib/components/competency/ScaleCriteriaMatrix.svelte` | Create | Competency × pillar grid with edit modal trigger |
| `web/src/lib/components/competency/AcceptanceLevelEditor.svelte` | Create | Profile selector + level 1-5 editor |
| `web/src/lib/components/competency/PillarFormModal.svelte` | Create | Create/edit pillar modal |
| `web/src/lib/components/competency/CompetencyFormModal.svelte` | Create | Create/edit competency modal |
| `web/src/lib/components/competency/ScaleCriterionModal.svelte` | Create | Edit 5 levels for one competency × pillar cell |
| `web/src/lib/components/competency/ConfirmDeleteModal.svelte` | Create | Shared delete confirmation dialog |
| `web/src/routes/rh/pilares/+page.svelte` | Create | Pillar list + CRUD page |
| `web/src/routes/rh/pilares/[id]/competencias/+page.svelte` | Create | Competency list for a pillar |
| `web/src/routes/rh/criterios-escala/+page.svelte` | Create | Scale criteria matrix page |
| `web/src/routes/rh/niveles-aceptacion/+page.svelte` | Create | Acceptance levels editor page |
| `web/src/routes/rh/competencias/+page.svelte` | Delete | Replaced by `/rh/pilares` entry point |
| `web/src/lib/nav/menuConfig.ts` | Modify | Replace "Competencias" entry; add Pilares, Criterios escala, Niveles aceptación |

## Interfaces / Contracts

```typescript
// web/src/lib/types/competency.ts

export interface Pillar {
  id: string;
  name: string;
  description: string;
}

export interface Competency {
  id: string;
  name: string;
  description: string;
  pillarId: string;
}

export interface ScaleCriterion {
  competencyId: string;
  pillarId: string;
  level: 1 | 2 | 3 | 4 | 5;
  description: string;
}

export interface AcceptanceLevel {
  profileId: EvaluationProfile;
  level: 1 | 2 | 3 | 4 | 5;
  label: string;
  description: string;
}
```

Store exposes: `getPillars()`, `addPillar(p)`, `updatePillar(id, p)`, `deletePillar(id)`, `getCompetenciesByPillar(pillarId)`, `addCompetency(c)`, `updateCompetency(id, c)`, `deleteCompetency(id)`, `getScaleCriteria(competencyId, pillarId)`, `updateScaleCriterion(sc)`, `getAcceptanceLevels(profileId)`, `updateAcceptanceLevel(al)`. Cascading delete: removing a pillar removes its competencies and scale criteria.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Store CRUD operations, cascade delete, validation | Vitest with direct store function calls |
| Unit | Type safety of fixtures against interfaces | `tsc` compile check on fixture imports |
| Integration | Route renders correct data from store | Vitest + Svelte Testing Library per route |
| E2E | CRUD flow: create pillar → add competency → edit criterion | Playwright (deferred to verify phase) |

## Migration / Rollout

No migration required. This is a greenfield UI module. Rollback: delete `web/src/routes/rh/` new routes, `web/src/lib/fixtures/competency/`, `web/src/lib/stores/competencyStore.svelte.ts`, `web/src/lib/types/competency.ts`, and `web/src/lib/components/competency/`. Restore `menuConfig.ts` and `/rh/competencias` placeholder.

## Open Questions

- [ ] Spec lists 9 profiles (includes `director-general`) but existing `evaluation.ts` has 8 — follow existing 8 until B4 (`org-hierarchy`) resolves.
- [ ] `openspec/config.yaml` l.97 open rule: scale criteria by profile in detail — current design uses criteria by competency × pillar only; profile-specific criteria deferred to spec closure.
