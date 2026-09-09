# Tasks: SED Roles y Privacidad

- [ ] Migración + seed `evaluation_profiles` (Gerente/Coordinador visibles, upsert idempotente)
- [ ] `rbac.go`: implementar `ProfileNameToRole` + tests tabla (incl. desconocido → colaborador)
- [ ] `mobonet_sync.go`: `jobTitleToProfileName` case-insensitive + fallback jefe → colaborador + tests
- [ ] Índice `lower(job_title)` en mapping (si tabla existe) o documentar por qué no
- [ ] Gates de privacidad: colaborador no lista/detalla evals avance/medio-año ajenas (403); jefe recomendado + RH fallback
- [ ] `api/openapi/auth.yaml`: documentar roles, alcances y 403
- [ ] UI: filtros por rol + `viewerMode` (self/team/all) ocultando evaluaciones ajenas al colaborador
- [ ] Validación: `medio-anio` interno (ASCII, sin ñ) vs display `medio año`; `openspec validate --all`; `go test ./...`
