# Technical Design: inline-goal-category-editing

## Architecture Overview
Replace modals with inline forms in CategoryCard and GoalRow. +page.svelte coordinates via isAnyInlineEditing flag.

## Component Changes
- CategoryFormModal.svelte: DELETED
- GoalFormModal.svelte: DELETED
- goalValidation.ts: CREATED (validateCategory, validateGoal, UNIT_OPTIONS)
- +page.svelte: Removed 5 state vars, added inline state, modified handlers
- CategoryCard.svelte: Added inline category edit + goal creation forms
- GoalRow.svelte: Added inline goal edit with KPI checkboxes

## Data Flow
+page → CategoryCard (onSaveCategory, onSaveGoal, isAnyInlineEditing) → GoalRow (onSaveGoal, editingGoalId)
