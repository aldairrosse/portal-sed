# Archive Report: KPI Library Route

**Date**: 2026-06-05
**Change**: kpi-library-route
**Status**: Completed

## Summary

Converted "Biblioteca de KPI" from a modal (`KpiFormModal`) to a dedicated sub-route at `/objetivos/asignacion/biblioteca` with DaisyUI breadcrumbs, a KPI table, and inline add/edit/delete capabilities.

## Artifacts

| Artifact | Path |
|----------|------|
| Proposal | `openspec/changes/kpi-library-route/proposal.md` |
| Tasks | `openspec/changes/kpi-library-route/tasks.md` |

## Files Changed

| File | Impact | Description |
|------|--------|-------------|
| `web/src/routes/objetivos/asignacion/biblioteca/+page.svelte` | New | KPI library route with table, inline form, inline edit, delete modal |
| `web/src/routes/objetivos/asignacion/+page.svelte` | Modified | Button navigates to route instead of opening modal |

## Decisions

- Breadcrumbs: "Asignación anual > Biblioteca" using DaisyUI `text-sm breadcrumbs`
- Inline form (no modal) for adding KPIs — button disabled while form is open
- Inline edit mode per row (one at a time)
- Delete with DaisyUI confirmation modal
- Table headers: `text-xs font-semibold text-base-content/60`
- Badges: `bg-primary/20 text-primary` for Unidad and Dirección
- Direction icons: `TrendingUp` / `TrendingDown` from lucide-svelte

## Verification

- `svelte-check` passes (0 errors, 11 pre-existing warnings)
- All 6 seed KPIs display in table
- Inline add/edit/delete flows work end-to-end
- Breadcrumbs navigate correctly

## Commits

1. `09c3f42` feat(goals): convert KPI library from modal to sub-route with breadcrumbs
2. `ebacfd8` style(kpi-table): headers with subtle text, primary badges, direction icons
3. (pending) style(kpi-table): bg-primary/20 in badges
