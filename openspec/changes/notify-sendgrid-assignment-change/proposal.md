# Proposal: notify-sendgrid-assignment-change

## Why
Colaborador→jefe ("Enviar asignación") y jefe→colaborador ("Solicitar cambio") hoy no envían email; se pierde trazabilidad y CTA a `/objetivos/asignacion`. Se usa SendGrid con templates + CTA full-path.

## What Changes
- `SendGridSender` detrás de `notify.Sender`; wiring por env en `api/cmd/server/main.go` (fallback `NoopSender`).
- Templates `assignment_submitted`, `change_requested` (+ demo `sendgrid_demo`) con header/título/descripción/CTA y tokens light SED.
- Helper `FullAssignmentURL(base)` reutilizando `appBaseURL` (`APP_BASE_URL`).
- Evento 1 idempotente: una sola vez por asignación (guard `status=submitted` + advisory lock existente en `AssignmentRepo`).
- Evento 2 con título/mensaje/CTA desde propuesta del jefe.

## Capabilities
- **New Capabilities**: `notifications/email` (provider, templates, CTA, demo visual).
- **Modified Capabilities**: `goals-and-weighting` (disparo Enviar asignación + idempotencia), `comments` (disparo Solicitar cambio vía change-request/proposal).
- **Modified Capabilities**: `goal-assignment-status` (guard estado submitted).

## Impact
`api/internal/service/notify/*`, `api/internal/handler/commentchange/handler.go`, `api/internal/handler/goal/*`, `api/cmd/server/main.go`, `.env.example`, `api/openapi/*.yaml`, `web/src/app.css` (solo tokens referencia).
