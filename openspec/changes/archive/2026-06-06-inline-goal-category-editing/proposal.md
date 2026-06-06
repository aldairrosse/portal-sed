# Proposal: inline-goal-category-editing

## Intent
Replace modals (CategoryFormModal, GoalFormModal) with inline forms for creating/editing categories and goals in the "Asignación anual" view. Reduces cognitive friction by keeping context visible while editing.

## Scope
### In scope
- Delete CategoryFormModal.svelte and GoalFormModal.svelte
- Modify CategoryCard.svelte (inline edit category + inline create goal)
- Modify GoalRow.svelte (inline edit goal with KPI checkboxes)
- Modify +page.svelte (inline create category, remove modal state, add isAnyInlineEditing coordination)
- Create goalValidation.ts (shared validation functions)

### Out of scope
- RequestChangeModal, GoalCommentModal, KpiFormModal unchanged
- Store and types unchanged
- Reader mode, medio-anio, fin-anio phases unchanged

## Approach
1. Local state per component — each CategoryCard and GoalRow manages its own $state
2. Single form open at a time — isAnyInlineEditing coordination via $bindable
3. Same validation rules as modals
4. CustomSelect for unit selection
5. KPIs shown as inline checkboxes (no popover/modal)
