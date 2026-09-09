# Tasks: Add Role Assignment Rule

- [ ] 1. Ent schema: add nillable `profile_id` + edge to `EvaluationProfile` on `GlobalGoalRule`; add `RuleTypeRole` const; `go generate`
- [ ] 2. Migration `000031_add_rule_type_role`: `ALTER TYPE rule_type ADD VALUE 'role'` (isolated, no DML using it) + `profile_id` column + index
- [ ] 3. DTO + service: `CreateRuleRequest.ProfileID *uuid.UUID`; validator `oneof=department min_direct_reports role`; map in `CreateGlobalGoal` (service + repo)
- [ ] 4. Repo `ExecuteRules`: add `RuleTypeRole` branch matching `employee.ProfileID == *rule.ProfileID AND IsActive(true)` (follow the `department` branch pattern, NOT the min_direct_reports stub)
- [ ] 5. Repo `GetGlobalGoal`: map `profile_id` in rule DTO response
- [ ] 6. OpenAPI `goals-api.yaml`: document global rules endpoints, `rule_type` enum incl. `role`, `profile_id`; regenerate TS types
- [ ] 7. Frontend `GlobalGoalCreateForm`: `RULE_TYPE_OPTIONS` + `Por rol`; `RuleRow.profileId`; profile picker (load `getProfiles()` explicitly); validation + payload ternary
- [ ] 8. Test: minimal table test for `ExecuteRules` role branch
- [ ] 9. Verify: `go test ./...` + `pnpm build`

## Verification notes

- Migration isolation: `'role'` must not be referenced in DML inside migration 000031.
- Do not extend `UpdateGlobalGoalRequest` with rules (existing limitation).
- Keep Ent validator + PG enum in sync.
