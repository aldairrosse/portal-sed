# Proposal: KPI Library Route

## Intent

Convert "Biblioteca de KPI" from a modal (`KpiFormModal`) to a dedicated sub-route at `/objetivos/asignacion/biblioteca` with breadcrumbs. The route shows a table of all KPIs, with an inline form for adding new KPIs (no modal). Only one "add" action can be in progress at a time — after saving, the form resets and the user can add another.

## Non-Goals

- Login or authentication
- Backend persistence (fixture-only)
- KPI linking to goals (already exists in goal form)
- KPI categories or grouping

## Scope

### In Scope
- New route `/objetivos/asignacion/biblioteca` as a child of asignacion
- Breadcrumbs: "Asignación anual > Biblioteca" using DaisyUI `text-sm breadcrumbs`
- Table listing all KPIs (name, description, unit, direction, target value, actions)
- Inline form for adding a new KPI (appears below table, no modal)
- Inline edit mode per row (click Edit → row becomes editable)
- Delete with confirmation (DaisyUI `modal`)
- "Agregar KPI" button disabled while an add form is open
- Link from asignacion page ("Biblioteca de KPI" button → route navigation)
- Return link/navigation back to asignacion

### Out of Scope
- Bulk import/export
- KPI search/filter (6 KPIs don't need it)
- KPI usage stats (which goals reference each KPI)

## Approach

**Phase 1 — Route + Table (~3 tasks)**
1. Create `/objetivos/asignacion/biblioteca/+page.svelte` with breadcrumbs and KPI table
2. Reuse `getKpis()` from goalsStore for data
3. Table columns: Nombre, Unidad, Dirección, Meta, Acciones (Editar/Eliminar)

**Phase 2 — Inline Form (~3 tasks)**
4. Build inline add form below the table (same fields as KpiFormModal but rendered inline)
5. Build inline edit mode per row (toggle row to editable state)
6. "Agregar KPI" button state: disabled when form is open, enabled after save

**Phase 3 — Integration (~2 tasks)**
7. Update asignacion page: "Biblioteca de KPI" button navigates to route (not modal)
8. Remove or deprecate `KpiFormModal` (or keep for future use)

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `web/src/routes/objetivos/asignacion/biblioteca/+page.svelte` | New | KPI library route with table + inline form |
| `web/src/routes/objetivos/asignacion/+page.svelte` | Modified | Button navigates instead of opening modal |
| `web/src/lib/components/goals/KpiFormModal.svelte` | Deprecated | Replaced by inline form in route |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Breadcrumbs don't match DaisyUI patterns | Low | Use standard `text-sm breadcrumbs` with links |
| Inline form feels disconnected from table | Medium | Place form directly below table with visual grouping |
| User confusion on "only one add at a time" | Low | Disable button + visual feedback when form is open |

## Success Criteria

- [ ] Clicking "Biblioteca de KPI" navigates to `/objetivos/asignacion/biblioteca`
- [ ] Breadcrumbs show "Asignación anual > Biblioteca"
- [ ] Table lists all 6 seed KPIs
- [ ] Inline form appears when clicking "Agregar KPI"
- [ ] Form saves to goalsStore (addKpi)
- [ ] "Agregar KPI" button disabled while form is open
- [ ] Edit toggles row to editable state inline
- [ ] Delete shows confirmation modal
