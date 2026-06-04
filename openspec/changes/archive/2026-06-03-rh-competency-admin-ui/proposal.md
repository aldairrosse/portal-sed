# Proposal: rh-competency-admin-ui (A2)

## Intent
Habilitar a RH a administrar el marco único de competencias: pilares, competencias por categoría, escala 1-5 con criterios por competencia y categoría, y niveles de aceptación por perfil de evaluación. UI-first con fixtures JSON, sin API ni persistencia. Desbloquea A3 (asignación de inicio de año).

## Scope

### In Scope
- CRUD visual de pilares únicos de la empresa (mock).
- CRUD visual de competencias por pilar/categoría.
- Matriz criterios escala 1-5: competencia × categoría, editable.
- Niveles de aceptación por perfil de evaluación (8 perfiles).
- Catálogo único de pilares/competencias visible desde cualquier perfil (decisión #2).
- Categorías sin ponderación: solo agrupan y muestran escalas (decisión #1).
- Fixtures JSON estáticas; carga síncrona en dev.
- Reutilizar shell, tokens y selectores dev de A1.
- Estados UI: skeleton, vacío, error (heredados).
- Sentence case, sin box-shadow decorativo, sin border-left/right.

### Out of Scope (Non-goals)
- Asignación masiva de competencias a empleados.
- Importación desde Excel.
- API real ni persistencia (recargar restaura fixtures).
- Autenticación y RBAC real (selector dev persona activo).
- Ponderación o suma de pesos en categorías/pilares.
- Cierre definitivo de "criterios por competencia/perfil" (regla abierta `openspec/config.yaml` l.97).
- Notificaciones email (no aplica a admin de catálogo).

## Capabilities

### New Capabilities
- `competency-admin-ui`: pantallas RH para pilares, competencias por categoría, criterios 1-5 y niveles de aceptación por perfil; UI-first con fixtures JSON; refleja decisiones #1 y #2.

### Modified Capabilities
- None — shell, tokens y selector dev de A1 se reutilizan sin cambios de spec.

## Approach
- UI-first estricto: cero llamadas de red, fixtures importadas en `web/src/lib/fixtures/`.
- Tipos TS centralizados en `web/src/lib/types/competency.ts` para que B2 (spec duradera `competency-framework`) y C3 (`competency-framework-api`) los adopten sin acoplar a la estructura de mock.
- Rutas lazy en `web/src/routes/rh/...` (paths exactos en design), protegidas por filtro shell (perfil = `rh`).
- Componentes DaisyUI para tablas, formularios, modales; sin estilos inline ni CSS raw (`principles/styles-and-ui.md`).
- Accesibilidad WCAG 2.1 AA heredada de A1; labels explícitos en inputs administrativos.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `web/src/routes/rh/pillars/` (tentativo) | New | Lista y editor de pilares |
| `web/src/routes/rh/competencies/` (tentativo) | New | Lista y editor de competencias |
| `web/src/routes/rh/scale-criteria/` (tentativo) | New | Matriz criterios 1-5 |
| `web/src/routes/rh/acceptance-levels/` (tentativo) | New | Niveles por perfil |
| `web/src/lib/fixtures/{pillars,competencies,acceptance-levels}.json` | New | Seed data |
| `web/src/lib/types/competency.ts` | New | Tipos compartidos con B2/C3 |
| `openspec/specs/competency-admin-ui/spec.md` | New | Spec duradera de la capability |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Forma de datos cambie al crear B2 (`competency-framework`) | Med | Tipos TS centralizados; fixtures aisladas |
| Criterios por perfil cierren distinto al supuesto | Med | Supuesto explícito en design; UI editable |
| RH espera asignación masiva (non-goal) | Low | Non-goal visible en landing del módulo |
| Categorías muestren peso (violación decisión #1) | Low | Acceptance test verifica ausencia de inputs de peso |

## Rollback Plan
- Borrar `web/src/routes/rh/...` y fixtures `web/src/lib/fixtures/{pillars,competencies,acceptance-levels}.json`.
- Sin cambios en specs durables de A1 → no requiere rearchive.
- Si ya archivado el change: abrir nuevo change correctivo, no revertir archivos.

## Dependencies
- A1 (`ui-shell-and-design-tokens`): COMPLETED, archivado en `openspec/changes/archive/2026-06-02-ui-shell-and-design-tokens/`.
- Specs durables A1: `ui-shell`, `design-tokens`, `dev-persona-cycle`.
- B2 (`competency-framework`): paralelo en Fase B; alinear nombres de campos al cierre.
- Regla abierta (`openspec/config.yaml` l.97): criterios de escala por perfil en detalle.

## Success Criteria
- [x] 4 pantallas RH renderizan con fixtures, sin errores de consola ni type-check.
- [x] CRUD visual de pilares y competencias (mock).
- [x] Matriz 1-5 editable: 8 perfiles × N competencias × M categorías.
- [x] Niveles de aceptación editables por perfil.
- [x] Sin input de peso en categorías (decisión #1).
- [x] Catálogo idéntico desde cualquier perfil (decisión #2).
- [x] Selector dev = `rh` muestra sección; otros perfiles la ocultan (filtro shell A1).
- [x] Estados skeleton, vacío, error operativos.
- [x] Sentence case, sin sombras decorativas, sin border-left/right.
- [x] `pnpm run lint` y `tsc` sin errores.
- [x] WCAG 2.1 AA: foco visible, labels, contraste.
