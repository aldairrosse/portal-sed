# Design: A6 — Manager 9×9 UI (Enhanced)

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│  Pages                                                                  │
│  /evaluacion/9x9              → NineBoxMatrix (all in-scope employees)  │
│  /evaluacion/9x9/competencias/[employeeId] → CompetencyNetworkView      │
│  /evaluacion/9x9/jerarquia    → OrgHierarchyTree (drill-down)           │
├────────────┬──────────────────┴─────────────────────────────────────────┤
│ nineBox    │ orgHierarchy     │ evaluationStore (existing)              │
│ Store      │ Store            │ competencyStore (existing)              │
│ (NEW)      │ (NEW)            │                                         │
├────────────┴──────────────────┴─────────────────────────────────────────┤
│  Types: nine-box.ts (NEW)  org-hierarchy.ts (NEW)  evaluation.ts (MOD)  │
├─────────────────────────────────────────────────────────────────────────┤
│  Components:                                                            │
│  nine-box/NineBoxMatrix, nine-box/NineBoxEntryCard,                     │
│  nine-box/NineBoxSliders, org-hierarchy/OrgHierarchyTree,               │
│  evaluation/CompetencyNetworkView                                       │
│  Reused: ComparisonTable, EmptyState, AssigneePicker                    │
├─────────────────────────────────────────────────────────────────────────┤
│  Fixtures: nine-box/matrix-entries.json                                 │
│            nine-box/quadrant-definitions.json                           │
│            nine-box/scale-definitions.json                              │
│            org-hierarchy/org-tree.json                                  │
└─────────────────────────────────────────────────────────────────────────┘
```

## Architecture Decisions

| Decision | Options | Tradeoff | Choice |
|----------|---------|----------|--------|
| 9×9 scale | Shared 1-5 (competency) vs independent 1-9 | Independent allows finer granularity for performance/potential | Independent 1-9 scale, separate from competency 1-5 |
| Hierarchy scope | Flat list vs recursive tree | Tree supports arbitrary depth and drill-down UX | Recursive OrgNode tree with getChildren/getDescendants |
| Competency detail | Graph (D3) vs table | Table is simpler, accessible, reusable; graph deferred | Table first (reuse ComparisonTable pattern), graph as future |
| DG drill-down | All directors at once vs one at a time | One-at-a-time reduces visual clutter for large orgs | Drill one director at a time from DG view |
| Store pattern | Single monolithic store vs split | Split follows existing pattern (evaluationStore, competencyStore) | Two stores: nineBoxStore + orgHierarchyStore |

## Types

### New: `web/src/lib/types/nine-box.ts`

```ts
/** 9×9 performance scale (1-9) */
export type NineBoxScale = 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9;

/** Quadrant classification based on performance × potential */
export type NineBoxQuadrant =
  | 'star'           // high perf + high potential (top-right)
  | 'growth'         // high perf + medium potential
  | 'high-potential' // medium perf + high potential
  | 'core-player'    // medium perf + medium potential (center)
  | 'risk'           // low perf + high potential
  | 'effective'      // low perf + medium potential
  | 'underperformer'; // low perf + low potential (bottom-left)

export interface NineBoxEntry {
  id: string;
  employeeId: string;
  employeeName: string;
  profileId: string;
  performance: NineBoxScale;   // X axis (1-9)
  potential: NineBoxScale;     // Y axis (1-9)
  quadrant: NineBoxQuadrant;   // computed from scores
  cycleId?: string;
}

export interface NineBoxQuadrantDef {
  id: NineBoxQuadrant;
  label: string;
  description: string;
  /** DaisyUI-compatible color class for cell background */
  colorClass: string;
  /** Performance range [min, max] inclusive */
  perfRange: [number, number];
  /** Potential range [min, max] inclusive */
  potRange: [number, number];
}

export interface NineBoxMatrix {
  entries: NineBoxEntry[];
  /** Filter: which employees are in scope for current viewer */
  scopeEmployeeIds: string[];
}
```

### New: `web/src/lib/types/org-hierarchy.ts`

```ts
export interface OrgNode {
  id: string;            // employeeId
  name: string;
  profileId: string;     // EvaluationProfile
  managerId: string | null;
  children: OrgNode[];
}
```

### Modified: `web/src/lib/types/evaluation.ts`

Add `'director-general'` to `EvaluationProfile` union, `EVALUATION_PROFILES` array, and `PROFILE_LABELS` record:

```ts
| 'director-general'   // new
// PROFILE_LABELS:
'director-general': 'Director General'
```

### Modified: `web/src/lib/types/goal.ts`

Add to `MANAGER_MAP`:

```ts
director: 'director-general'
```

`director-general` has `managerId: null` (root of hierarchy), so no entry needed for it in MANAGER_MAP.

## Stores

### `web/src/lib/stores/nineBoxStore.svelte.ts`

**State:**
```ts
let entries = $state<NineBoxEntry[]>(structuredClone(matrixEntriesData));
let quadrantDefs = $state<NineBoxQuadrantDef[]>(structuredClone(quadrantDefsData));
```

**Getters:**

| Function | Signature | Notes |
|----------|-----------|-------|
| `getMatrixEntries` | `(scopeIds: string[]) => NineBoxEntry[]` | Filter entries by scope |
| `getEntryByEmployee` | `(employeeId: string) => NineBoxEntry \| undefined` | Single entry |
| `getQuadrantDefs` | `() => NineBoxQuadrantDef[]` | All quadrant definitions |
| `getQuadrantForScores` | `(perf: NineBoxScale, pot: NineBoxScale) => NineBoxQuadrant` | Pure calculation |
| `getEntriesByQuadrant` | `(scopeIds: string[], quadrant: NineBoxQuadrant) => NineBoxEntry[]` | Filter by quadrant |
| `getQuadrantStats` | `(scopeIds: string[]) => Record<NineBoxQuadrant, number>` | Count per quadrant |

**Mutations:**

| Function | Guard | Notes |
|----------|-------|-------|
| `setEntryScores(employeeId, performance, potential)` | phase = `'fin-anio'` | Upsert entry, auto-compute quadrant |
| `bulkSetEntries(entries: NineBoxEntry[])` | — | Replace all (for fixture loading) |

**Quadrant calculation logic:**

The 9×9 grid is split into 3×3 bands:
- Low: 1-3, Medium: 4-6, High: 7-9

| | Pot Low (1-3) | Pot Med (4-6) | Pot High (7-9) |
|---|---|---|---|
| **Perf High (7-9)** | growth | growth | star |
| **Perf Med (4-6)** | effective | core-player | high-potential |
| **Perf Low (1-3)** | underperformer | effective | risk |

```ts
function computeQuadrant(perf: number, pot: number): NineBoxQuadrant {
  const perfBand = perf <= 3 ? 'low' : perf <= 6 ? 'mid' : 'high';
  const potBand = pot <= 3 ? 'low' : pot <= 6 ? 'mid' : 'high';
  const key = `${perfBand}-${potBand}`;
  const QUADRANT_MAP: Record<string, NineBoxQuadrant> = {
    'high-high': 'star',
    'high-mid': 'growth',
    'mid-high': 'high-potential',
    'mid-mid': 'core-player',
    'low-high': 'risk',
    'low-mid': 'effective',
    'low-low': 'underperformer',
    'mid-low': 'effective',
    'high-low': 'growth'
  };
  return QUADRANT_MAP[key] ?? 'core-player';
}
```

### `web/src/lib/stores/orgHierarchyStore.svelte.ts`

**State:**
```ts
let tree = $state<OrgNode>(structuredClone(orgTreeData));
```

**Getters:**

| Function | Signature | Algorithm |
|----------|-----------|-----------|
| `getRoot` | `() => OrgNode` | Return tree root |
| `getChildren` | `(nodeId: string) => OrgNode[]` | BFS to find node, return `.children` |
| `getDescendants` | `(nodeId: string) => OrgNode[]` | Recursive DFS, flatten all descendants |
| `getSubtree` | `(nodeId: string) => OrgNode \| null` | Deep clone of subtree rooted at nodeId |
| `getNodeById` | `(nodeId: string) => OrgNode \| null` | BFS traversal |
| `getScopeIds` | `(nodeId: string) => string[]` | getNode + getDescendants → map to IDs (inclusive) |
| `getDepth` | `(nodeId: string) => number` | Count edges from root to node |
| `getAllLeafIds` | `(nodeId: string) => string[]` | DFS, collect nodes with empty children |

**Mutations:**

| Function | Notes |
|----------|-------|
| `replaceTree(newTree: OrgNode)` | For fixture swap |

**Traversal algorithms:**

```ts
// getDescendants — iterative DFS to avoid stack overflow on deep trees
function getDescendants(root: OrgNode, nodeId: string): OrgNode[] {
  const target = findNode(root, nodeId);
  if (!target) return [];
  const result: OrgNode[] = [];
  const stack = [...target.children];
  while (stack.length > 0) {
    const node = stack.pop()!;
    result.push(node);
    stack.push(...node.children);
  }
  return result;
}

// findNode — BFS
function findNode(root: OrgNode, nodeId: string): OrgNode | null {
  if (root.id === nodeId) return root;
  const queue = [...root.children];
  while (queue.length > 0) {
    const node = queue.shift()!;
    if (node.id === nodeId) return node;
    queue.push(...node.children);
  }
  return null;
}
```

## Components

### New

| Component | Props | Est. lines |
|-----------|-------|:----------:|
| `NineBoxMatrix` | `entries: NineBoxEntry[]`, `quadrantDefs: NineBoxQuadrantDef[]`, `onEntryClick(entry)`, `readonly?` | ~180 |
| `NineBoxEntryCard` | `entry: NineBoxEntry`, `onClose()`, `onNavigateCompetencias?`, `onNavigateJerarquia?` | ~60 |
| `NineBoxSliders` | `performance: NineBoxScale`, `potential: NineBoxScale`, `onChange(perf, pot)`, `disabled?` | ~80 |
| `OrgHierarchyTree` | `root: OrgNode`, `onNodeSelect(node)`, `selectedNodeId?`, `maxDepth?` | ~100 |
| `CompetencyNetworkView` | `employeeId: string`, `employeeName: string` | ~70 |

#### NineBoxMatrix — detailed design

Renders a 9×9 CSS grid. X-axis = performance (1 left → 9 right), Y-axis = potential (1 bottom → 9 top). Each cell shows count badge + colored dot per employee in that cell.

```
Props:
  entries: NineBoxEntry[]
  quadrantDefs: NineBoxQuadrantDef[]
  onEntryClick: (entry: NineBoxEntry) => void
  readonly?: boolean

Events:
  entryclick → CustomEvent<NineBoxEntry>

Slots:
  (none — entry detail handled via NineBoxEntryCard in parent)
```

Accessibility:
- `role="grid"` with `aria-rowcount="9"` and `aria-colcount="9"`
- Each cell: `role="gridcell"` with `aria-label="Performance {p}, Potential {q}, {n} employees"`
- Keyboard: arrow keys navigate cells, Enter opens entry detail
- Color: quadrant backgrounds use DaisyUI `bg-success/20`, `bg-warning/20`, `bg-error/20` with sufficient contrast; employee dots use `badge` component

#### NineBoxSliders — detailed design

Two range inputs (DaisyUI `range` component) for performance and potential. Shows current quadrant label computed from scores.

```
Props:
  performance: NineBoxScale
  potential: NineBoxScale
  onChange: (perf: NineBoxScale, pot: NineBoxScale) => void
  disabled?: boolean
```

#### OrgHierarchyTree — detailed design

Recursive component: renders a node with expand/collapse toggle, children rendered recursively. Uses DaisyUI `tree` / `menu` component for visual hierarchy.

```
Props:
  root: OrgNode
  onNodeSelect: (node: OrgNode) => void
  selectedNodeId?: string
  maxDepth?: number  // default: unlimited

Internal state:
  expandedIds: Set<string>  // tracks which nodes are expanded
```

Accessibility:
- `role="tree"` on container, `role="treeitem"` on each node
- `aria-expanded` on expandable nodes
- Arrow keys: Up/Down navigate siblings, Right expand, Left collapse, Enter select

#### CompetencyNetworkView — detailed design

Reuses `ComparisonTable` pattern. Shows self-rating vs RH rating for all competencies of a single employee, grouped by pillar.

```
Props:
  employeeId: string
  employeeName: string

Data flow:
  competencyStore.getPillars() → getCompetenciesByPillar() →
  evaluationStore.getCompetencyRatings(employeeId) →
  pass to ComparisonTable (showRhColumn=true)
```

## Pages

### `/evaluacion/9x9/+page.svelte` — 9×9 Matrix (REPLACE placeholder)

**Profile guard:** Only `jefe`, `director`, `director-general`, `rh` see the matrix. Others → `EmptyState`.

**Layout:**
1. Header: "Matriz 9×9" + profile badge
2. Scope selector: `NineBoxSliders` (for RH/director to set scores) or read-only view
3. `NineBoxMatrix` with entries scoped by viewer's hierarchy:
   - `jefe` → direct reports only (`getChildren(managerId)`)
   - `director` → all descendants under their span (`getDescendants(directorId)`)
   - `director-general` → all employees in org tree
4. Click entry → `NineBoxEntryCard` popup with links to competencias/jerarquia

**Data flow:** `getProfile()` → `orgHierarchyStore.getScopeIds(viewerId)` → `nineBoxStore.getMatrixEntries(scopeIds)` → render matrix.

### `/evaluacion/9x9/competencias/[employeeId]/+page.svelte`

**Layout:**
1. Breadcrumb: 9×9 → employee name
2. Header: "Competencias de {name}"
3. `CompetencyNetworkView` with `employeeId` from params

### `/evaluacion/9x9/jerarquia/+page.svelte`

**Profile guard:** Only `director`, `director-general` see hierarchy view.

**Layout:**
1. Header: "Jerarquía organizacional"
2. `OrgHierarchyTree` rooted at viewer's scope:
   - `director-general` → full tree from root
   - `director` → subtree from their node
3. Selecting a node → shows summary: employee name, profile, 9×9 position, link to competencias

## Fixtures

### `nine-box/matrix-entries.json`

```json
[
  {
    "id": "nb-emp-colaborador-01",
    "employeeId": "emp-colaborador-01",
    "employeeName": "Juan Pérez",
    "profileId": "colaborador",
    "performance": 5,
    "potential": 6,
    "quadrant": "core-player"
  }
]
```

~12 entries covering all 8 existing employees + DG. Mixed quadrants for visual variety.

### `nine-box/quadrant-definitions.json`

```json
[
  {
    "id": "star",
    "label": "Estrella",
    "description": "Alto desempeño, alto potencial",
    "colorClass": "bg-success/20",
    "perfRange": [7, 9],
    "potRange": [7, 9]
  }
]
```

7 quadrant definitions (star, growth, high-potential, core-player, risk, effective, underperformer).

### `nine-box/scale-definitions.json`

```json
[
  { "level": 1, "label": "Muy bajo", "description": "No cumple expectativas" },
  { "level": 5, "label": "Moderado", "description": "Cumple parcialmente" },
  { "level": 9, "label": "Excepcional", "description": "Supera consistentemente" }
]
```

Key anchor levels (1, 3, 5, 7, 9) with labels for slider tooltips.

### `org-hierarchy/org-tree.json`

```json
{
  "id": "emp-dg-01",
  "name": "Carlos Mendoza",
  "profileId": "director-general",
  "managerId": null,
  "children": [
    {
      "id": "emp-director-01",
      "name": "Laura Torres",
      "profileId": "director",
      "managerId": "emp-dg-01",
      "children": [
        {
          "id": "emp-jefe-01",
          "name": "María García",
          "profileId": "jefe",
          "managerId": "emp-director-01",
          "children": [
            { "id": "emp-colaborador-01", "name": "Juan Pérez", "profileId": "colaborador", "managerId": "emp-jefe-01", "children": [] },
            { "id": "emp-vendedor-01", "name": "Carlos López", "profileId": "vendedor", "managerId": "emp-jefe-01", "children": [] }
          ]
        },
        { "id": "emp-gerente-tienda-01", "name": "Roberto Díaz", "profileId": "gerente-tienda", "managerId": "emp-director-01", "children": [] },
        { "id": "emp-divisional-01", "name": "Ana Martínez", "profileId": "divisional", "managerId": "emp-director-01", "children": [] },
        { "id": "emp-regional-01", "name": "Pedro Sánchez", "profileId": "regional", "managerId": "emp-director-01", "children": [] },
        { "id": "emp-rh-01", "name": "Sofía Ramírez", "profileId": "rh", "managerId": "emp-director-01", "children": [] }
      ]
    }
  ]
}
```

4 levels: DG → Director → Jefe/other direct reports → Colaboradores. Reuses existing employee IDs from `assignments.json`.

## Modifications to Existing Files

| File | Change | Details |
|------|--------|---------|
| `web/src/lib/types/evaluation.ts` | Add `'director-general'` | Add to union type, `EVALUATION_PROFILES` array, `PROFILE_LABELS` record |
| `web/src/lib/types/goal.ts` | Add to `MANAGER_MAP` | `director: 'director-general'` entry |
| `web/src/lib/stores/devContext.svelte.ts` | No change needed | `EvaluationProfile` type already flows through; adding `'director-general'` to the type is sufficient |
| `web/src/routes/evaluacion/9x9/+page.svelte` | Replace | Replace `EmptyState` placeholder with full matrix page |

## Integration Points

| Dependency | How A6 uses it |
|------------|---------------|
| `devContext` | `getProfile()` for profile guard + scope detection, `getPhase()` for mutation guards |
| `competencyStore` | `getPillars()`, `getCompetenciesByPillar()`, `getCompetencyAcceptanceLevel()` for CompetencyNetworkView |
| `evaluationStore` | `getCompetencyRatings(employeeId)` for self vs RH comparison in CompetencyNetworkView |
| `ComparisonTable` | Reused directly in CompetencyNetworkView |
| `assignments.json` | Employee IDs and names reused in org-tree.json fixture |

## Profile-Based Scope Matrix

| Viewer Profile | Matrix Scope | Hierarchy Access | Can Edit Scores |
|---------------|-------------|-----------------|-----------------|
| `jefe` | Direct reports only | `getChildren(viewerId)` | No (read-only) |
| `director` | All descendants | `getDescendants(viewerId)` | No (read-only) |
| `director-general` | Entire org | `getRoot()` → all | No (read-only) |
| `rh` | All employees | Full tree | Yes (`setEntryScores`) |

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| 9×9 grid too dense with many employees | Cell shows count badge; click to expand list; future: virtualize |
| Org tree fixture diverges from assignments.json | Document that org-tree.json is derived from assignments; single source of truth note in fixture header |
| `director-general` not in existing `EvaluationProfile` breaks type checks | Add to union in PR 1; all exhaustive switches get `'director-general'` case |
| WCAG color contrast on quadrant backgrounds | Use DaisyUI `bg-{color}/20` with `text-base-content` for text; test with axe-core |
| Recursive OrgHierarchyTree stack overflow on deep trees | Iterative DFS in store; component caps visual depth at `maxDepth` prop |
