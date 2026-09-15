# Design: notify-sendgrid-assignment-change

## Context
Ver `proposal.md - Why`. Estado actual verificado: `notify.Sender/NoopSender/SMTPSender+templateFS` (`api/internal/service/notify/`), `commentchange/handler.go:notifyOwner+appBaseURL`, `AssignmentRepo` con advisory-lock + `UpdateAssignmentStatus(status,submitted_at)`, middleware Idempotency existente. Falta provider SendGrid y disparo en 2 eventos.

## Goals / Non-Goals
- Goals: provider SendGrid tras `Sender`, templates `assignment_submitted/change_requested` con CTA full-path, idempotencia Enviar asignación, memoria Engram compartida en subagentes.
- Non-Goals: worker/cola persistente (sigue goroutine in-process), JWT, multi-tenant, rediseño UI.

## Decisions
1. **SendGrid tras `Sender` (no SMTP nuevo):** `sendgrid_sender.go: type SendGridSender struct{client *sendgrid.Client; from, baseURL string; tmpl map[Template]*template.Template}` reusando `templateFS go:embed templates/*.tmpl`. Alternativa SMTP existente: descartada (credenciales + deliverability). Fallback: sin `SENDGRID_API_KEY` → `NoopSender` + log.
2. **Wiring env en `main.go`:** `SENDGRID_API_KEY/MAIL_FROM/APP_BASE_URL`; `NewSendGridSender` solo si key≠""; inyectar mismo `Sender` a `commentchange.NewHandler` y `goal.NewHandler`. Ej: `var s notify.Sender = notify.NoopSender{}; if k:=os.Getenv("SENDGRID_API_KEY"); k!="" { s,_ = notify.NewSendGridSender(k, from, base) }`.
3. **Templates + tokens light SED:** reutilizar `web/src/app.css` (`#ffffff/#f6f8fc/#334155/#2e63e8/#4a90a4/#d97706`), sin box-shadow. Campos `{{.Title}} {{.Message}} {{.CTAURL}} {{.CTALabel}}`.
4. **Helper CTA full-path `absURL`:** `func FullAssignmentURL(base string) string { return strings.TrimSuffix(base,"/")+"/objetivos/asignacion" }` junto a `appBaseURL`; valida `https?://`.
5. **Idempotencia Enviar asignación:** guard `if row.Status=="submitted" || row.SubmittedAt!=nil { skip Send }` + advisory-lock `hashEmployeeCycle` existente + `Idempotency-Key` middleware. Solicitar cambio sin guard (cada propuesta notifica).

## Risks / Trade-offs
- [SDK nuevo `sendgrid-go` en go.mod] → fijar versión, fallback Noop si init falla.
- [Doble-click concurrente] → guard status/submitted_at dentro del lock; test doble POST.
- [CTA relativa rota en email] → validación `https?://` + test render.
- [Goroutine pierde email al reiniciar] → aceptado; documentado como `ponytail:` upgrade a tabla notifications.

## Migration Plan
1. Merge change → `go mod tidy`, `.env.example` con 3 vars. 2. Deploy sin key = Noop (seguro). 3. Set key en prod → demo endpoint verifica. Rollback: vaciar key vuelve a Noop.

## Open Questions
- Ninguna bloqueante; remitente verificado SendGrid se define en `MAIL_FROM`.
