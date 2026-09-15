# notifications/email — delta
## ADDED Requirements
### Requirement: SendGrid provider via Sender
El sistema SHALL enviar email vía `SendGridSender implements notify.Sender` configurado por `SENDGRID_API_KEY`, `MAIL_FROM`, `APP_BASE_URL`; sin key SHALL usar `NoopSender` y log.
#### Scenario: sin key no rompe
- WHEN `SENDGRID_API_KEY=""` THEN usa `NoopSender` y requests retornan 2xx sin email.
### Requirement: Templates + CTA full-path
Templates `assignment_submitted`, `change_requested` SHALL rendir header/título/descripción/CTA con `{{.Title}} {{.Message}} {{.CTAURL}} {{.CTALabel}}`; `CTAURL` SHALL ser absolute URL `FullAssignmentURL(APP_BASE_URL)="/objetivos/asignacion"`.
#### Scenario: CTA absoluto
- WHEN render THEN `CTAURL` inicia con `http(s)://` y contiene `/objetivos/asignacion`.
### Requirement: Tokens light SED
HTML SHALL usar solo tokens `sed` light: `#ffffff/#f6f8fc/#334155/#2e63e8/#4a90a4/#d97706`; sin box-shadow.
#### Scenario: demo visual
- WHEN demo con datos reales THEN header/título/descripción/CTA visibles y correctos.
