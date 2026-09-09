---
title: "SED Roles y Privacidad"
status: draft
change: sed-roles-privacidad
---

# Proposal: SED Roles y Privacidad

## Why

El sync Mobonet asigna perfiles sin regla visible: solo existe `colaborador` y el mapeo `job_title` es sensible a mayúsculas, dejando Gerentes/Coordinadores como colaboradores sin alcance de equipo. Además no hay regla de privacidad: un colaborador podría ver evaluaciones de avance/medio-año ajenas o el flujo de jefe recomendado / RH fallback. Sin esto se bloquea RBAC real y la confianza en evaluaciones.

## What

- Hacer visibles 2 roles: `Gerente` y `Coordinador` (además de los existentes) con alcance de equipo (descendientes).
- Upsert de mapping `job_title` → perfil **case-insensitive** con fallback al jefe (manager_id) cuando el título no matchea.
- Privacidad colaborador: NO ve evaluaciones de avance ni de medio-año ajenas; solo las propias. Flujo jefe recomendado con fallback RH cuando no hay jefe asignado.
- Sembrar `evaluation_profiles` con los perfiles visibles y documentar contrato en `auth.yaml` + `viewerMode` en UI.

## Impact

- `api/`: `auth/rbac.go` (ProfileNameToRole), `mobonet_sync.go` (jobTitleToProfileName), migración + seed `evaluation_profiles`, gates en handlers de evaluación.
- `api/openapi/auth.yaml`: roles visibles y códigos 403 por alcance.
- `web/`: filtros de menú/tabla por rol, `viewerMode` (team/all/self) para ocultar evaluaciones ajenas al colaborador.
- Sin cambio en cómputo de tiers ni matriz 9×9 (lo consume como scope).
