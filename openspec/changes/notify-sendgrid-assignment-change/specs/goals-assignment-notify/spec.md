# goals-and-weighting + goal-assignment-status — delta
## ADDED Requirements
### Requirement: Enviar asignación idempotente
Al terminar asignación (`UpdateAssignmentStatus→submitted`) el sistema SHALL notificar al jefe UNA sola vez por `(employee_id,cycle_id)`; doble-click/retry SHALL NOT reenviar (guard: si `status` ya `submitted/enviada` o `submitted_at!=nil` skip; + `Idempotency-Key` middleware existente).
#### Scenario: doble click
- WHEN dos `POST submit` seguidos THEN un solo `Send(assignment_submitted)` al email del jefe.
### Requirement: Solicitar cambio
Al crear propuesta del jefe (`CreateChangeRequest`/`CreateProposal`) el sistema SHALL notificar al colaborador con título+mensaje+CTA full-path.
#### Scenario: propuesta jefe
- WHEN jefe crea propuesta THEN `Send(change_requested)` al email del colaborador con `Title/Message/CTAURL`.
