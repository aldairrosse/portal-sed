# Tasks: notify-sendgrid-assignment-change

Orden: a→b→c→d→e→f→g. Compartir memoria Engram `notify/sendgrid-plan` en subagentes.

- [ ] (a) config/env + SendGrid provider
  Files: `api/internal/service/notify/sendgrid_sender.go`, `api/cmd/server/main.go`, `.env.example`, `api/go.mod`
  Ej: `s,_ := notify.NewSendGridSender(os.Getenv("SENDGRID_API_KEY"), os.Getenv("MAIL_FROM"), os.Getenv("APP_BASE_URL"))`
  Acept: sin key → NoopSender + 2xx; con key → SendGridSender inyectado en goal+commentchange handlers. Dep: design. Primero.

- [ ] (b) templates + tokens + CTA helper
  Files: `api/internal/service/notify/templates/*.tmpl`, `api/internal/service/notify/sender.go` (const Template), `handler/commentchange/handler.go` (`FullAssignmentURL`)
  Ej: `{{.Title}} {{.Message}} <a href="{{.CTAURL}}">{{.CTALabel}}</a>` con `#ffffff/#f6f8fc/#334155/#2e63e8`
  Acept: render header/título/descripción/CTA; CTA inicia `https?://` y contiene `/objetivos/asignacion`. Dep: a.

- [ ] (c) Enviar asignación + anti-spam
  Files: `api/internal/handler/goal/*`, `api/internal/repository/goal/assignment_repo.go`
  Ej: `if row.Status=="submitted"||row.SubmittedAt!=nil {skip} else {Send(assignment_submitted)}` bajo advisory-lock
  Acept: doble `POST submit` → un solo email al jefe. Dep: a,b.

- [ ] (d) Solicitar cambio
  Files: `api/internal/handler/commentchange/handler.go`
  Ej: `h.notifier.Send(ctx, Notification{Template:"change_requested", Data:{Title,Message,CTAURL:FullAssignmentURL(base)}})`
  Acept: propuesta jefe → email colaborador con título/mensaje/CTA; autor==dueño skip. Dep: a,b.

- [ ] (e) test demo SendGrid + visual
  Files: endpoint demo temporal + `go test ./internal/service/notify/`
  Ej: `POST /_demo/sendgrid?to=x@y` render real
  Acept: email llega / log Noop; header/título/descripción/CTA visibles. Dep: a-d.

- [ ] (f) validación backend ambos eventos
  Files: `*_test.go` goal + commentchange
  Acept: test doble-click (1 send), test propuesta (1 send + CTA absoluto), fallo SendGrid no rompe 2xx/201. Dep: c,d.

- [ ] (g) docs/OpenAPI
  Files: `api/openapi/*.yaml`, `.env.example`
  Acept: `validate --all --strict` verde; ops documentan `SENDGRID_API_KEY/MAIL_FROM/APP_BASE_URL`. Dep: c-f. Cierra change.
