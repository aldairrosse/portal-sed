# Seguridad

## Threat model (ligero)

- Acceso horizontal entre empleados/organizaciones.
- Escalación de privilegios vía API sin RBAC.
- Inyección SQL/XSS (mitigar ORM + escape Svelte + CSP).
- Abuso de login y scraping de listados.

## Controles

| Control | Implementación |
|---------|----------------|
| AuthN | Sesión segura o SSO |
| AuthZ | RBAC en capa service |
| Transporte | TLS terminado en proxy |
| Headers | HSTS, CSP, X-Frame-Options |
| Secrets | Env / SSM, nunca en repo |
| Dependencias | pnpm audit, govulncheck en Go |
| Imágenes | Usuario no-root, Trivy en CI |
| Logs | Sin PII en texto plano; trace_id |

## Datos sensibles

- Evaluaciones = datos personales/laborales; retención y borrado según política RH (spec legal).

## Skill OpenSpec

`security-baseline` — checklist de aceptación por cada change.
