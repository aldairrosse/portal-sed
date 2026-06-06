# Proposal: A5 — Annual Evaluation UI

## Intent

Build the year-end evaluation screen (`/mi-evaluacion`) for employees to self-rate competencies (1–5 scale) and close goals, plus the RH formal evaluation view (`/rh/evaluaciones`) with side-by-side self-rating vs RH rating comparison. Driven by fixture data; no backend, login, or email notifications.

## Non-Goals

- Login or authentication
- Backend persistence (OpenAPI stubs only)
- Email notifications
- 9×9 potential matrix (jefe records potential separately)

## Scope

### In Scope
- Phase guard: all mutations gated to `cyclePhase === 'cierre'`
- Competency self-rating: 3 pillars, 8 competencies, scale 1–5, acceptance level reference, optional comment per competency
- Goal closure: final progress (reuses `progress` field), self-assessment comment per goal, status freeze
- RH evaluation form: RH rates competencies, sees employee self-rating alongside, adds RH assessment
- Comparison table (RH view): employee self-rating vs RH rating vs acceptance level per competency
- Evaluation status badges: pending / in-progress / completed per employee
- Employee picker (RH view): AssigneePicker component reused

### Out of Scope
- 9×9 potencial matrix
- Email notifications
- Backend persistence
- Login/auth

## Approach

**Phase 1 — Foundation (~5 tasks)**
1. Resolve CyclePhase naming: update `evaluation.ts` to use `'asignacion'|'avance'|'cierre'` (unify on `goal.ts` form since it's already used in stores)
2. Create `evaluation-result.ts` with `CompetencyRating`, `GoalClosure`, `EvaluationStatus` types
3. Create `evaluationStore.svelte.ts` with phase-gated mutations
4. Update `goalsStore` `getGoalPermissions()` for `cierre` phase
5. Update `cycle.json` phase to `'cierre'`

**Phase 2 — Employee Self-Evaluation (~6 tasks)**
6. Build `ScaleRatingSelector` (1–5 radio with level labels)
7. Build `CompetencyRatingCard` (pillar card with competencies + scale selector + comment)
8. Build `GoalClosureCard` (read-only goal + final progress + self-assessment textarea)
9. Create fixtures: `evaluations/self-evaluations.json` (2–3 employees, mixed completion states)
10. Assemble `/mi-evaluacion` page with two sections (competencies + goal closure)
11. Test with dev toolbar phase switching

**Phase 3 — RH Formal Evaluation (~6 tasks)**
12. Build `EvaluationStatusBadge` component
13. Build `ComparisonTable` (self-rating vs RH rating vs acceptance level)
14. Create fixtures: `evaluations/rh-evaluations.json`, `evaluations/goal-closures.json`
15. Assemble `/rh/evaluaciones` page with employee picker + evaluation form + comparison
16. Test with dev toolbar

**Phase 4 — Integration (~3 tasks)**
17. Verify dev toolbar phase switching works across all screens
18. Test with all 8 evaluation profiles
19. Visual polish + accessibility (WCAG 2.1 AA)

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `web/src/routes/mi-evaluacion/+page.svelte` | New | Self-evaluation screen |
| `web/src/routes/rh/evaluaciones/+page.svelte` | New | RH formal evaluation screen |
| `web/src/lib/types/evaluation-result.ts` | New | Rating, closure, status types |
| `web/src/lib/stores/evaluationStore.svelte.ts` | New | Evaluation state store |
| `web/src/lib/stores/goalsStore.svelte.ts` | Modified | Cierre phase permissions |
| `web/src/lib/fixtures/goals/cycle.json` | Modified | Phase set to `'cierre'` |
| `web/src/lib/fixtures/evaluations/` | New | Self/RH evaluation fixtures |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| CyclePhase naming conflict causes store type mismatch | High | Resolve first; update `evaluation.ts` to match `goal.ts` form |
| `getGoalPermissions()` missing `cierre` case | Medium | Add in Phase 1; test with employee + RH views |
| Fixture data volume for 8 profiles | Low | Prioritize 3 profiles (colaborador, jefe, vendedor) with varying states |

## Rollback Plan

- Revert `evaluation.ts` CyclePhase alias if conflict causes type errors
- Delete `evaluationStore.svelte.ts` and revert page imports if store approach doesn't work
- Revert `cycle.json` to `'avance'` if phase logic blocks testing

## Dependencies

- `competencyStore` (A2) — pillars, competencies, acceptance levels, scale criteria
- `goalsStore` (A3/A4) — goal progress, category cards, goal row components
- `AssigneePicker` (existing) — employee selection in RH view
- `ProgressIndicator` (existing) — final progress display
- `CategoryCard` / `GoalRow` (A4) — reuse in read-only mode for goal closure

## Success Criteria

- [ ] Employee can self-rate all 8 competencies with scale 1–5 and comments
- [ ] Employee can add closing self-assessment per goal
- [ ] RH can select any employee and see their self-evaluation status
- [ ] RH can rate competencies and see side-by-side comparison (self vs RH vs acceptance)
- [ ] RH can add RH assessment per goal
- [ ] All mutations blocked outside `cierre` phase
- [ ] Dev toolbar phase switching works
- [ ] Works with all 8 profiles via fixtures