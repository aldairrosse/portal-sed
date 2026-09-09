# Design: SED Roles y Privacidad

## Context

Perfiles actuales en `evaluation_profiles` sin Gerente/Coordinador visibles; sync Mobonet matchea `job_title` exacto (case-sensitive) y deja nulos como colaborador sin fallback. Handlers de evaluación no filtran por viewer (ver `TODO(auth:C7)` análogo en nine-box).

## Decisions

### D1. `rbac.go`: ProfileNameToRole

- `ProfileNameToRole(name string) Role`: normaliza (`TrimSpace` + `ToLower`, sin tildes colapsadas) y mapea `gerente`/`coordinador`/`jefe`/`rh`/`colaborador`/etc. a roles canónicos.
- Desconocido → `colaborador` (mínimo privilegio). Tabla exhaustiva + test unitario.

### D2. `mobonet_sync.go`: jobTitleToProfileName

- `jobTitleToProfileName(title string) string`: case-insensitive (`EqualFold` / lower) contra catálogo `Gerente|Coordinador|Jefe|...`.
- Upsert mapping tabla (si existe) con índice `lower(job_title)`; si no matchea → fallback al perfil del jefe (`manager_id`), si no hay jefe → `colaborador`.

### D3. Seed `evaluation_profiles`

- Migración goose: `INSERT ... ON CONFLICT (name) DO UPDATE` para `gerente`, `coordinador` visibles (`visible=true` / donde aplique), idempotente.

### D4. `auth.yaml`

- Documentar enum de roles visibles, alcance (`team` vs `all` vs `self`) y 403 fuera de alcance en endpoints de evaluación.

### D5. `viewerMode`

- Helper UI/server: `viewerMode = self | team | all` derivado del rol. Colaborador → `self` en avance/medio-año: lista filtra a propias; detalle ajeno → 403.
- Jefe recomendado → evaluador sugerido; si `manager_id` NULL → RH fallback como evaluador.
