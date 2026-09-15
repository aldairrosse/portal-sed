# Roles y autenticación

## Por definir en OpenSpec (checklist)

- [ ] Modelo de identidad: email corporativo, SSO (Azure AD / Okta), o ambos.
- [ ] Roles base: empleado, líder/evaluador, RH, administrador org, superadmin plataforma.
- [ ] Permisos granulares: leer/escribir catálogo, fijar objetivos, evaluar, ver resultados agregados.
- [ ] Alcance: solo subordinados directos vs árbol completo vs por departamento.

## Principios

- Autenticación en backend; frontend solo envía cookies httpOnly o headers acordados.
- Autorización **siempre** en servicio Go, no solo ocultando botones en UI.
- Sesión con rotación y expiración; logout invalida servidor.
- MFA para roles RH/admin si política de seguridad lo pide.

## Flujos mínimos

1. Login → sesión → redirect a dashboard según rol.
2. Cambio de ciclo activo en contexto de sesión o header `X-Cycle-Id` validado en API.
3. Impersonación (solo admin) auditada si aplica soporte.

## Skill / spec sugerida OpenSpec

`openspec/specs/identity-access.md` — primer change recomendado antes de cualquier módulo de negocio.
