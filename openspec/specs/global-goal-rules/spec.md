# global-goal-rules Specification

## Purpose
Support assigning global goals by role via profile-linked rules.

## Requirements

### Requirement: Global goal rules SHALL support assignment by role

`global_goal_rules.rule_type` SHALL accept a third value `role` alongside `department` and `min_direct_reports`, with a nullable `profile_id` FK to `evaluation_profiles(id)` (indexed). `ExecuteRules` SHALL match active employees whose `profile_id` equals the rule's `profile_id`. `CreateRuleRequest` SHALL accept `profile_id` and validate `rule_type` as `oneof=department min_direct_reports role`; OpenAPI `goals-api.yaml` SHALL document the new enum value and field.

#### Scenario: Assign global goal by role

- WHEN RH creates a global goal with rule `{ rule_type: 'role', profile_id: '<jefe-profile>' }`
- THEN every active employee with that `profile_id` is assigned the goal on execution
