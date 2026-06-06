# Delta Spec: inline-goal-category-editing

## Domain: goal-assignment-ui

## Requirements

### REQ-001: Inline category creation (P0)
User can create a category directly in the page without opening a modal.

### REQ-002: Inline category editing (P0)
User can edit a category directly in the card. Pencil converts header to inputs.

### REQ-003: Inline goal creation (P0)
User can create a goal directly in the card without opening a modal.

### REQ-004: Inline goal editing (P0)
User can edit a goal directly in the row. Each td transforms to inputs with inline KPI checkboxes.

### REQ-005: State coordination (P0)
Only one inline form open at a time via isAnyInlineEditing flag.

### REQ-006: Disable buttons during editing (P0)
New buttons disabled while any form is open.

### REQ-007: Delete modal files (P0)
CategoryFormModal.svelte and GoalFormModal.svelte removed.

### REQ-008: No store changes (P0)
goalsStore.svelte.ts unchanged.

### REQ-009: No changes to out-of-scope modals (P0)
RequestChangeModal, GoalCommentModal, KpiFormModal unchanged.

### REQ-010: Success/error feedback (P1)
Operations show success/error feedback via alerts.
