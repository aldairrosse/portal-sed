# Add Role Assignment Rule

## Goal

Add a third `global_goal_rules.rule_type` value, `role`, so RH can assign a global goal to every employee whose profile matches (e.g. Jefe, Director), alongside the existing `department` and `min_direct_reports` rule types.

## Non-goals

- Fixing the incomplete `min_direct_reports` execution branch (out of scope; note below).
- Editing rules via `PUT /goals/global/{id}` (existing API does not support it; out of scope).
- Adding a new UI route; the rule picker lives inside the existing `GlobalGoalCreateForm` modal.

## Scope

| Layer | Change |
|-------|--------|
| DB | Migration `000031_add_rule_type_role`: `ALTER TYPE rule_type ADD VALUE 'role'`; add nullable `profile_id UUID REFERENCES evaluation_profiles(id)` + index on `global_goal_rules(profile_id)` |
| Ent schema | `GlobalGoalRule` gains nillable `profile_id` field + edge to `EvaluationProfile`; add `RuleTypeRole` const; run `go run entgo.io/ent/cmd/ent generate ./internal/schema` |
| Repo | `CreateGlobalGoal` + `GetGlobalGoal` map `profile_id`; `ExecuteRules` gains `RuleTypeRole` branch: match `employee.ProfileID == *rule.ProfileID AND employee.IsActive(true)` |
| DTO/validation | `CreateRuleRequest` gains `ProfileID *uuid.UUID`; `rule_type` validator becomes `oneof=department min_direct_reports role` |
| Frontend | `RULE_TYPE_OPTIONS` + `Por rol`; `RuleRow` gains `profileId`; conditional profile picker fed by `competencyStore.getProfiles()`; payload ternary + validation; types in `web/src/lib/api/globalGoals.ts` |
| OpenAPI | Document `rule_type` enum incl. `role` + `profile_id` in `goals-api.yaml` (note: global rules endpoints currently undocumented — document them) |

## Gotchas (documented constraints)

- **Postgres enum migration**: `ALTER TYPE ... ADD VALUE` cannot run inside the same transaction that uses the new value (goose limitation). Keep migration 000031 isolated and do not reference `'role'` in the same migration file's DML.
- **`UpdateGlobalGoalRequest` has no rules**: the existing PUT cannot edit rules; do not rely on it. Rule changes happen via create + delete only.
- **`min_direct_reports` branch is a stub**: `ExecuteRules` (global_repo.go:273-282) fetches ALL active employees without a real report-count check. Do NOT copy it as the template for `role`; the `department` branch (265-272) is the correct pattern.
- **Profile names are slugs** (`colaborador`, `jefe`): display via `PROFILE_LABELS` mapping (`$lib/types/evaluation`); store/compare the raw slug.
- **No tests for ExecuteRules**: add a minimal table test for the new branch; do not assume existing coverage.
- **Double enum validation**: Ent's `RuleTypeValidator` AND the PG enum both reject unknown values; keep both in sync when adding `role`.
- **competencyStore may not be loaded** in the goals page context: load profiles explicitly (`getProfiles()`) before rendering the role picker; do not assume the store is initialized.

## Precondition

`evaluation_profiles` table exists (yes — `employees.profile_id` FK already references it, required non-null). `GET /api/v1/profiles` endpoint already exists.

## Status

Draft — proposal only.
