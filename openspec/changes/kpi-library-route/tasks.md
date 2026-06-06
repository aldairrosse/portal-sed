# Tasks: KPI Library Route

## Task 1: Create route + breadcrumbs + table
- Create `web/src/routes/objetivos/asignacion/biblioteca/+page.svelte`
- Breadcrumbs: Home > Asignación anual > Biblioteca (DaisyUI breadcrumbs)
- Table with columns: Nombre, Unidad, Dirección, Meta, Acciones
- Load KPIs from goalsStore
- Empty state when no KPIs

## Task 2: Inline add form
- "Agregar KPI" button at bottom of table
- When clicked, shows inline form below table (same fields as KpiFormModal)
- Fields: nombre (required), descripción (required), unidad (select), dirección (select), valor objetivo (optional)
- Save calls `addKpi()` from goalsStore
- After save, form hides, button re-enables
- While form is open, button is disabled

## Task 3: Inline edit mode
- "Editar" button per row toggles that row to editable state
- Input fields replace text in the row
- Save/Cancel buttons appear
- Only one row can be in edit mode at a time

## Task 4: Delete with confirmation
- "Eliminar" button per row opens DaisyUI confirmation modal
- Confirm calls `deleteKpi()` from goalsStore

## Task 5: Update asignacion page
- Change "Biblioteca de KPI" button from `onclick={() => (showKpiLibrary = true)}` to `goto('/objetivos/asignacion/biblioteca')`
- Remove `KpiFormModal` import and usage from asignacion page

## Task 6: Verify
- Breadcrumbs navigate correctly
- Table shows all 6 seed KPIs
- Inline add works end-to-end
- Inline edit works
- Delete with confirmation works
- Button disabled state works correctly
