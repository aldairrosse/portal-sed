## ADDED Requirements

### Requirement: OAuth/SSO exploration SHALL be recorded as docs-only decision input

This change is an exploration record (see `exploration.md`): the platform already has session-based auth (`session_token` httpOnly cookie, RBAC, `SSOAdapter` noop) and no OAuth provider is wired yet. No code, migration, or endpoint changes are introduced by this change; it SHALL serve only as the documented gap analysis and approach input for a future OAuth implementation change.

#### Scenario: Validation accepts docs-only change

- WHEN `openspec validate --all --strict` parses this change
- THEN it finds at least one requirement with a scenario and reports no missing-delta error
