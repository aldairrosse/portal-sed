# SED — Roadmap de specs (UI-first)

Decisiones de producto cerradas, convenciones OpenSpec y orden de changes con contexto para `/opsx:propose`.

**Estado del repo:** especificación; sin código de aplicación hasta aprobar cada change.

---

## Decisiones de producto (cerradas)

| # | Tema | Decisión |
|---|------|----------|
| 1 | **Ponderación** | **Doble ponderación 100%:** categorías de metas suman 100%, y metas dentro de cada categoría (ponderadas) suman 100%. Categorías de metas son **independientes** de pilares de competencias. |
| 2 | **Pilares** | **Todos los perfiles** (corporativo, tiendas, RH, etc.) usan el **mismo catálogo de pilares** y competencias. El perfil de evaluación cambia criterios de escala y niveles de aceptación, no el catálogo base. |
| 3 | **Medio año** | Permite **editar metas** y registrar **avances**. **No** permite **eliminar** metas ya creadas (solo ajuste de campos permitidos en spec). |
| 4 | **Fin de año** | **Autoevaluación del empleado** + **evaluación RH** (competencias / cierre formal). El **jefe** evalúa **potenciales** para matriz **9×9** (desempeño vs potencial), no sustituye la evaluación RH de competencias. |
| 5 | **Categorías de metas** | Cada **usuario define sus propias categorías** de metas (independientes de pilares). Categorías son agrupaciones custom para organizar metas. |
| 6 | **KPIs** | KPIs son **indicadores** (numérico, porcentaje o moneda) que pueden vincularse a **1 o más metas**. Cada meta puede tener un KPI asociado. |
| 7 | **Roles evaluables** | **RH también tiene metas** y puede ser evaluado. Todos los perfiles (incluido RH) son evaluables con el mismo modelo de metas. |
| 8 | **Jerarquía de edición** | Cada usuario define sus metas, categorías, ponderaciones y KPIs. **Jefe/director/gerente pueden VER** definiciones de personas a cargo y **SOLICITAR CAMBIOS** (ajustar KPIs y ponderaciones), **NO borrar ni agregar metas**. |

### Jerarquías (contexto, sin decisión pendiente)

- **Corporativa:** colaborador → jefe → director → director general.
- **Retail:** vendedor → gerente de tienda → divisional → regional.
- Ambas conviven; el evaluador y “mis evaluados” dependen del nodo en el árbol (spec `org-hierarchy`).

### Ciclo anual (3 fases)

| Fase | Quién actúa | Qué ocurre |
|------|-------------|------------|
| **Inicio de año** | RH + empleado | RH asigna competencias por pilar; empleado define metas (KPI % o $) con ponderación 100%. |
| **Medio año** | Empleado (+ jefe según spec de visibilidad) | Edición de metas y avances; sin borrado de metas. |
| **Fin de año** | Empleado, jefe (9×9), RH | Autoevaluación; jefe califica potencial 9×9; RH evaluación formal y cierre. |

---

## Convenciones OpenSpec

### Nombres

- **Changes (carpeta):** `kebab-case`, prefijo por capa si ayuda: `ui-*`, luego dominio sin prefijo, luego `*-api`, `identity-access`.
- **Specs duraderas (`openspec/specs/`):** mismo kebab-case; una spec por bounded context o módulo de pantalla estable.

### Flujo

1. `/opsx:propose "<texto del bloque Contexto propose>"` — genera change en `openspec/changes/<nombre>/`.
2. Revisar y aprobar artefactos (`proposal`, `design`, `tasks`, specs delta).
3. `/opsx:apply` — solo cuando el change esté aprobado.
4. `/opsx:archive` — al cerrar; sincronizar spec duradera en `openspec/specs/`.

### UI-first (esta ruta)

- Pantallas primero con **fixtures JSON** en `web/` (cuando exista) y **selector de persona** en dev (sin login).
- **OpenAPI stub** o tipos manuales alineados a spec; API real en fase C.
- **Non-goals** obligatorio en cada proposal: qué no incluye el change (ej. “sin SSO”, “sin persistencia”).
- Español en textos de producto y specs; código en inglés o español consistente por archivo.
- Alineación con `openspec/config.yaml` y `principles/`; si hay conflicto, gana OpenSpec tras revisión explícita.

### Perfiles de evaluación (fixture / RBAC futuro)

`colaborador` · `jefe` · `vendedor` · `gerente-tienda` · `divisional` · `regional` · `director` · `director-general` · `rh`

### Validaciones de negocio (referencia en specs)

- Metas: suma de pesos = 100% (o validación de monto total si aplica).
- Categorías: sin suma de pesos; solo visualización de escalas y competencias asignadas.
- Medio año: `UPDATE` metas y avances; **sin** `DELETE` de metas.
- Fin año: tres vías paralelas documentadas — autoevaluación, 9×9 jefe, cierre RH.

---

## Orden de specs y contexto para propose

Ejecutar en orden. No saltar fases A→B sin cerrar decisiones de dominio que afecten la UI siguiente.

### Fase A — Pantallas (fixtures, sin backend ni auth)

#### A1 — `ui-shell-and-design-tokens`

**Contexto propose:**

```text
/opsx:propose "Shell UI SED: layout autenticado, menú por perfil de evaluación, design tokens DaisyUI, selector de persona en dev (8 perfiles), selector de fase de ciclo (inicio/medio/fin año). UI-first sin login ni API. Non-goals: SSO, persistencia, backend."
```

**Entrega:** navegación lazy, estados vacío/error/skeleton, tokens (color, radius, tipografía), convención sentence case.

**Fixtures:** lista de perfiles y fases; sin datos de evaluación aún.

---

#### A2 — `rh-competency-admin-ui`

**Contexto propose:**

```text
/opsx:propose "Pantallas RH: administración de pilares únicos para toda la empresa, competencias por categoría, escala 1-5 con criterios por competencia y categoría, niveles de aceptación por perfil de evaluación. UI con fixtures JSON, sin API. Non-goals: asignación masiva a empleados, importación Excel."
```

**Entrega:** CRUD visual de pilares y competencias (mock); matriz criterios escala × perfil.

**Depende de:** A1.

**Refleja decisiones:** #2 pilares únicos; categorías sin ponderación (#1).

---

#### A3 — `annual-assignment-ui`

**Contexto propose:**

```text
/opsx:propose "Pantalla inicio de año: empleado crea y edita metas agrupadas en categorías custom (independientes de pilares). Doble ponderación 100%: categorías suman 100%, metas dentro de cada categoría suman 100%. KPIs indicadores (numérico/porcentaje/moneda) vinculables a 1+ metas. Unidades porcentaje o moneda por meta. Todos los perfiles (incluido RH) pueden tener metas. Jefe/director/gerente pueden ver definiciones de personas a cargo y solicitar cambios (ajustar KPIs y ponderaciones, no borrar/agregar metas). Fixtures JSON, sin API. Non-goals: medio año, evaluación final, 9x9."
```

**Entrega:** formulario metas por categoría, KPI vinculado, doble validación 100%, vista de solo lectura para jefes.

**Depende de:** A1.

**Refleja decisiones:** #1 (doble ponderación), #2 (catálogo único), #5 (categorías custom), #6 (KPIs), #7 (roles evaluables), #8 (jerarquía de edición).

---

#### A4 — `mid-year-progress-ui`

**Contexto propose:**

```text
/opsx:propose "Pantalla medio de año: edición de metas existentes y registro de avances por meta, sin eliminar metas. Semáforo o indicador de avance por pilar. Fixtures con metas precargadas. Non-goals: evaluación 1-5 final, 9x9, borrado de metas."
```

**Entrega:** edición campos de meta, avance %/valor, bloqueo de eliminar meta.

**Depende de:** A3 (mismas fixtures extendidas).

**Refleja decisiones:** #3.

---

#### A5 — `annual-evaluation-ui`

**Contexto propose:**

```text
/opsx:propose "Pantalla fin de año: autoevaluación empleado (competencias escala 1-5 y cierre de metas), vista RH para evaluación formal del empleado. Fixtures por perfil. Non-goals: login, persistencia, notificaciones email."
```

**Entrega:** flujo autoevaluación; panel RH evaluación; estados completado/pendiente.

**Depende de:** A3, A4.

**Refleja decisiones:** #4 (parte empleado + RH).

---

#### A6 — `manager-9x9-ui`

**Contexto propose:**

```text
/opsx:propose "Pantalla jefe: matriz 9x9 desempeño vs potencial para colaboradores a cargo, lista desde jerarquía mock corporativa y retail. Solo calificación de potencial/desempeño para cuadrante, no reemplaza evaluación RH. Fixtures JSON. Non-goals: API org real, agregados empresa."
```

**Entrega:** grid o selector cuadrante 9×9 por evaluado; lista “mis evaluados” mock.

**Depende de:** A1, A5 (mismo ciclo fin de año).

**Refleja decisiones:** #4 (parte jefe).

---

#### A7 — `my-evaluatees-ui` (opcional en paralelo tras A4)

**Contexto propose:**

```text
/opsx:propose "Lista mis evaluados para jefe/gerente/divisional: árbol mock corporativo y retail, acceso a fijación y seguimiento de metas de subordinados en inicio y medio año. Fixtures. Non-goals: auth, API jerarquía real."
```

**Depende de:** A1, A3, A4.

---

### Fase B — Dominio (specs duraderas, poco o sin código)

Ejecutar en paralelo con A2–A5 o justo antes de `/opsx:apply` de la pantalla relacionada.

| Orden | Spec duradera | Contexto propose |
|-------|---------------|------------------|
| B1 | `evaluation-lifecycle` | `/opsx:propose "Ciclo anual SED: fases inicio/medio/fin, transiciones de estado, quién edita en cada fase, prohibición eliminar metas en medio año. Sin implementación."` |
| B2 | `competency-framework` | `/opsx:propose "Marco de competencias: pilares únicos, escala 1-5, criterios por competencia y categoría, niveles de aceptación por perfil. Categorías sin ponderación. Sin UI ni API."` |
| B3 | `goals-and-weighting` | `/opsx:propose "Metas y KPIs: unidades porcentaje y moneda, ponderación solo en metas suma 100%, reglas de edición medio año sin delete. Sin implementación."` |
| B4 | `org-hierarchy` | `/opsx:propose "Jerarquía dual corporativa y retail, evaluador, alcance mis evaluados, perfiles de evaluación. Sin auth ni API."` |
| B5 | `manager-9x9` | `/opsx:propose "Matriz 9x9: ejes desempeño y potencial, cuadrantes, quién califica (jefe), separación de evaluación RH. Sin implementación."` |

---

### Fase C — Backend y auth (reemplazar mocks)

| Orden | Change | Contexto propose |
|-------|--------|------------------|
| C1 | `data-model-core` | `/opsx:propose "Modelo de datos PostgreSQL/Ent: organización, empleado, ciclo, fase, pilar, competencia, meta, evaluación, calificación 9x9. Índices para listados. Sin UI."` |
| C2 | `evaluation-lifecycle-api` | `/opsx:propose "API REST fases del ciclo anual y reglas de transición según spec evaluation-lifecycle. OpenAPI 3.1."` |
| C3 | `competency-framework-api` | `/opsx:propose "API pilares y competencias RH, asignación a empleados inicio de año. OpenAPI."` |
| C4 | `goals-api` | `/opsx:propose "API metas: CRUD inicio año, update medio año sin delete, validación suma 100%. OpenAPI."` |
| C5 | `org-hierarchy-api` | `/opsx:propose "API árbol organizacional corporativo y retail, mis evaluados por evaluador. OpenAPI."` |
| C6 | `evaluations-and-9x9-api` | `/opsx:propose "API autoevaluación, evaluación RH fin de año, registro 9x9 jefe. OpenAPI."` |
| C7 | `identity-access` | `/opsx:propose "Autenticación sesión httpOnly, RBAC por perfil, reemplazo selector persona dev. OpenAPI login/sesión."` |
| C8 | `wire-api-replace-mocks` | `/opsx:propose "Conectar web/ a cliente openapi-fetch, eliminar fixtures en rutas productivas, mantener mocks solo en dev flag."` |

---

## Resumen visual

```
A1 shell → A2 RH admin → A3 inicio año → A4 medio año → A5 fin año (auto+RH) → A6 9x9 jefe
                ↓              ↑ B1–B5 dominio (paralelo)
C1 data → C2–C6 APIs → C7 auth → C8 wire
```

---

## Checklist antes de `/opsx:apply`

- [ ] Decisiones de este documento reflejadas en proposal y design.
- [ ] Non-goals explícitos.
- [ ] Fixtures documentados (nombres de archivos y perfiles).
- [ ] `openspec validate --all` sin errores.
- [ ] Aprobación humana del change.

---

## Referencias

- Contexto IA: `openspec/config.yaml`
- Principios: `principles/evaluations-domain.md`, `principles/modules-and-screens.md`
- Arranque OpenSpec: `docs/get-started/openspec.md`
- Agentes: `AGENTS.md`
