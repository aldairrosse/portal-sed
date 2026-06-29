## ADDED Requirements

### Requirement: Role-based tree filtering

The system SHALL filter the visible org tree nodes based on the user's evaluator scope. The frontend SHALL call `GET /evaluator-scopes?evaluatorId={me}` to obtain the scope, then filter the full tree to show only nodes in `scopeData.orgNodeIds`.

#### Scenario: Director-general sees full tree

- **WHEN** user with profile `director-general` loads `/evaluacion/9x9/jerarquia`
- **THEN** evaluator scope returns `scopeType: 'department'` with all orgNodeIds in the tree
- **AND** full org tree renders with all nodes visible
- **AND** all nodes are selectable

#### Scenario: Director sees subtree

- **WHEN** user with profile `director` loads `/evaluacion/9x9/jerarquia`
- **THEN** evaluator scope returns `scopeType: 'department'` with orgNodeIds under their node
- **AND** only nodes within their subtree render in the tree
- **AND** nodes outside their scope are hidden (not just disabled)
- **AND** the tree root is the director's own node (or nearest ancestor in scope)

#### Scenario: Jefe sees direct reports only

- **WHEN** user with profile `jefe` loads `/evaluacion/9x9/jerarquia`
- **THEN** evaluator scope returns `scopeType: 'individual'` with employeeIds of direct reports
- **AND** system renders a table of direct reports instead of the org tree
- **AND** table columns are: nombre, puesto, perfil
- **AND** no tree component is rendered

#### Scenario: RH sees full tree

- **WHEN** user with profile `rh` loads `/rh/jerarquia`
- **THEN** evaluator scope returns all orgNodeIds (full access)
- **AND** full org tree renders with all nodes visible
- **AND** selecting a node shows area metrics from the backend API

### Requirement: Evaluator scope fetch on store init

The `orgHierarchyStore` SHALL fetch the evaluator scope for the current user when loading the org tree. The scope SHALL be used to filter which nodes are visible and selectable.

#### Scenario: Store fetches scope before rendering tree

- **WHEN** `orgHierarchyStore.load()` is called
- **THEN** system calls `GET /evaluator-scopes?evaluatorId={currentUserId}`
- **AND** stores `scopeData.orgNodeIds` for filtering
- **AND** filters the loaded tree to only include nodes whose IDs are in `orgNodeIds`
- **AND** root of filtered tree is the highest-depth node the user has access to

#### Scenario: Scope fetch failure

- **WHEN** evaluator-scopes endpoint returns 404
- **THEN** store sets error state "No se pudo cargar tu alcance de evaluación"
- **AND** tree does not render
- **AND** an EmptyState component shows the error message

#### Scenario: Scope fetch network error

- **WHEN** evaluator-scopes endpoint returns network error
- **THEN** store sets error state "Error de conexión al cargar alcance"
- **AND** tree does not render

#### Scenario: Scope returns empty orgNodeIds

- **WHEN** evaluator-scopes returns `scopeData.orgNodeIds: []`
- **THEN** store renders EmptyState "No tienes nodos asignados en tu alcance"
- **AND** no tree nodes render

### Requirement: Profile-specific view routing

The system SHALL render different views based on user profile:
- `director-general`, `director`, `rh`: Show org tree with detail panel
- `jefe`: Show table of direct reports (no tree)

#### Scenario: Jefe route renders table

- **WHEN** user with profile `jefe` navigates to `/evaluacion/9x9/jerarquia`
- **THEN** page renders a table of direct reports with columns: nombre, puesto, perfil
- **AND** no org tree component is rendered
- **AND** selecting a row shows detail panel for that employee

#### Scenario: Director route renders tree

- **WHEN** user with profile `director` navigates to `/evaluacion/9x9/jerarquia`
- **THEN** page renders the org tree (filtered to their subtree)
- **AND** selecting a node shows detail panel with metrics

#### Scenario: Profile change handling

- **WHEN** user has profile `jefe` but navigates to `/rh/jerarquia`
- **THEN** page checks profile and redirects to `/evaluacion/9x9/jerarquia`
- **AND** renders the jefe table view

### Requirement: Tree filtering algorithm

The filtering SHALL work as follows:
1. Load full tree from `GET /org-trees/{treeId}/nodes?format=nested`
2. Load scope from `GET /evaluator-scopes?evaluatorId={me}`
3. Collect `scopeData.orgNodeIds` into a Set
4. Walk the tree recursively, keeping only nodes whose `id` is in the Set
5. If a node is kept, keep all its ancestors (to maintain tree structure)
6. The resulting filtered tree's root is the highest node in the user's scope

#### Scenario: Filtering preserves ancestors

- **WHEN** user scope includes nodes at depth 3 and 4
- **THEN** their parent at depth 2 is included (even if not in scope)
- **AND** the grandparent at depth 1 is included
- **AND** the tree root at depth 0 is included

#### Scenario: Filtering removes out-of-scope branches

- **WHEN** user scope includes node A but not node B (sibling of A)
- **THEN** node A and its subtree render
- **AND** node B and its subtree do not render
- **AND** their common parent renders (as ancestor of A)

## Acceptance Criteria

1. `orgHierarchyStore.load()` calls `GET /evaluator-scopes?evaluatorId={me}` before rendering tree
2. Tree is filtered to only show nodes in `scopeData.orgNodeIds` (plus ancestors)
3. Jefe profile renders table of direct reports, not tree
4. Director/DG/RH profiles render filtered or full tree
5. Scope fetch failure shows EmptyState with error message
6. Empty scope shows EmptyState with appropriate message
7. `pnpm run check` passes
