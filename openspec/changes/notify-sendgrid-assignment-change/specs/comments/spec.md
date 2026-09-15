# comments — delta
## ADDED Requirements
### Requirement: Solicitar cambio notifica al colaborador
Al crear change-request/proposal del jefe (`commentchange/handler.go:CreateChangeRequest|CreateProposal`) el sistema SHALL enviar `Send(change_requested)` al email del colaborador con título+mensaje+CTA full-path vía `notify.Sender` inyectado.
#### Scenario: jefe solicita cambio
- WHEN `POST /assignments/{id}/change-request` con `{title,message}` THEN email a colaborador contiene `Title/Message/CTAURL=https://*/objetivos/asignacion` y response 201 aun si SendGrid falla (log, no bloquea).
#### Scenario: sin auto-email
- WHEN autor == dueño THEN skip Send (reusa `notifyOwner` guard).
