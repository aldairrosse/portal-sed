# Design: org-hierarchy-api (C5)

## 1. Overview

**C5 is the backbone of SED.** Every evaluation, goal assignment, and reporting query ultimately resolves through the organizational hierarchy. This change implements the REST API for managing dual organizational trees (corporate and retail), employees, org nodes, evaluator scope computation, and "my evaluatees" queries.

**Key design constraints:**
- Read-heavy workload: tree traversal accounts for ~95% of hierarchy traffic.
- Every API call depends on hierarchy resolution; latency must be sub-50ms for cached reads.
- Structural changes (node moves, employee transfers) are rare but must be atomic and safe.
- Dual tree support: corporate and retail hierarchies coexist; an employee belongs to exactly one tree.

**Decision from proposal:** Tree traversal strategy uses **PostgreSQL `ltree` materialized path** as the primary mechanism, with a **closure table fallback** documented for environments where `ltree` is unavailable.

---

## 2. Package Structure

```
api/internal/
├── handler/org/
│   ├── orgtree_handler.go          # GET /api/v1/org-trees, /:treeId, /:treeId/nodes, /:treeId/export
│   ├── orgnode_handler.go          # GET/POST/PUT/DELETE /api/v1/org-nodes, /:nodeId/move
│   ├── employee_handler.go         # GET /api/v1/employees, /:empId, /:empId/evaluatees, /:empId/manager, /:empId/ancestors, /search, POST /batch
│   ├── evaluator_handler.go        # GET /api/v1/evaluator-scopes
│   ├── routes.go                   # Chi router registration for all org endpoints
│   └── handler_test.go             # Table-driven HTTP tests (all handlers)
├── service/org/
│   ├── orgtree_service.go          # Tree logic: GetTrees, GetTree, GetTreeNodes, ExportTree
│   ├── orgnode_service.go          # Node logic: CRUD, MoveNode with cycle detection
│   ├── employee_service.go         # Employee logic: search, evaluatees, chain of command, batch lookup
│   ├── evaluator_service.go        # Scope computation: ResolveEvaluatorScope, ResolveEvaluator
│   └── service_test.go             # Unit tests with mocked repositories
├── repository/org/
│   ├── orgtree_repo.go             # Ent queries for OrgNode trees (ltree + flat)
│   ├── orgnode_repo.go             # OrgNode CRUD, self-referential edge updates
│   ├── employee_repo.go            # Employee filters, search, direct reports
│   ├── evaluator_repo.go           # EvaluatorScope lookup
│   └── repo_test.go                # Repository tests against testcontainer PostgreSQL
└── pkg/
    ├── tree/
    │   ├── traversal.go            # Tree flattening, nesting, depth filtering
    │   └── closure.go              # Closure table ops (fallback if ltree unavailable)
    ├── search/
    │   └── employee_search.go      # Fuzzy name/email/number search with tsvector
    └── cache/
        └── org_cache.go            # Redis wrappers for tree and evaluatee caching
```

**Principle alignment:** `handler/` is thin — validation, DTO mapping, error response formatting. `service/` contains all business rules (cycle detection, cache invalidation, scope computation). `repository/` contains Ent queries only; no business logic.

---

## 3. Handler Layer

All handlers follow the pattern: parse → validate → call service → map response → set cache headers.

### Endpoint Catalog (14 endpoints)

#### Organization Trees

| # | Method | Path | Request | Response | Errors |
|---|--------|------|---------|----------|--------|
| 1 | `GET` | `/api/v1/org-trees` | Query: `type` (optional: `corporate`\|`retail`) | `OrgTreeListResponse` | — |
| 2 | `GET` | `/api/v1/org-trees/:treeId` | Path: `treeId` (UUID) | `OrgTreeDetailResponse` | `TREE_NOT_FOUND` 404 |
| 3 | `GET` | `/api/v1/org-trees/:treeId/nodes` | Query: `format` (`flat`\|`nested`, default `flat`), `depth` (int, default `all`) | `OrgNodeListResponse` | `TREE_NOT_FOUND` 404 |
| 4 | `GET` | `/api/v1/org-trees/:treeId/export` | Path: `treeId` | `Transfer-Encoding: chunked` JSON stream | `TREE_NOT_FOUND` 404 |

#### Org Nodes

| # | Method | Path | Request | Response | Errors |
|---|--------|------|---------|----------|--------|
| 5 | `GET` | `/api/v1/org-nodes/:nodeId` | Path: `nodeId` (UUID) | `OrgNodeDetailResponse` | `NODE_NOT_FOUND` 404 |
| 6 | `POST` | `/api/v1/org-nodes` | Body: `CreateOrgNodeRequest` | `OrgNodeDetailResponse` | `INVALID_PARENT` 400, `INVALID_TREE_TYPE` 400 |
| 7 | `PUT` | `/api/v1/org-nodes/:nodeId` | Body: `UpdateOrgNodeRequest` | `OrgNodeDetailResponse` | `NODE_NOT_FOUND` 404, `STALE_VERSION` 409 |
| 8 | `DELETE` | `/api/v1/org-nodes/:nodeId` | Path: `nodeId` | `204 No Content` | `NODE_NOT_FOUND` 404, `NODE_HAS_CHILDREN` 409 |
| 9 | `POST` | `/api/v1/org-nodes/:nodeId/move` | Body: `MoveOrgNodeRequest` (`newParentId`) | `OrgNodeDetailResponse` | `NODE_NOT_FOUND` 404, `INVALID_PARENT` 400 |

#### Employees

| # | Method | Path | Request | Response | Errors |
|---|--------|------|---------|----------|--------|
| 10 | `GET` | `/api/v1/employees` | Query: `treeId`, `nodeId`, `profileId`, `isActive`, `q`, `cursor`, `limit` | `EmployeeListResponse` | — |
| 11 | `GET` | `/api/v1/employees/:empId` | Path: `empId` (UUID) | `EmployeeDetailResponse` | `EMPLOYEE_NOT_FOUND` 404 |
| 12 | `GET` | `/api/v1/employees/:empId/evaluatees` | Path: `empId` | `EmployeeListResponse` | `EMPLOYEE_NOT_FOUND` 404 |
| 13 | `GET` | `/api/v1/employees/:empId/manager` | Path: `empId` | `EmployeeDetailResponse` | `EMPLOYEE_NOT_FOUND` 404 |
| 14 | `GET` | `/api/v1/employees/:empId/ancestors` | Path: `empId` | `AncestorChainResponse` | `EMPLOYEE_NOT_FOUND` 404 |
| — | `POST` | `/api/v1/employees/batch` | Body: `BatchEmployeeRequest` (`ids[]`) | `EmployeeListResponse` | — |
| — | `GET` | `/api/v1/employees/search` | Query: `q` (string, min 2 chars), `limit` (default 20, max 50) | `EmployeeListResponse` | — |

#### Evaluator Scopes

| # | Method | Path | Request | Response | Errors |
|---|--------|------|---------|----------|--------|
| — | `GET` | `/api/v1/evaluator-scopes` | Query: `evaluatorId` (UUID), `cycleId` (UUID, optional) | `EvaluatorScopeResponse` | `EMPLOYEE_NOT_FOUND` 404 |
| — | `GET` | `/api/v1/evaluator-scopes/:scopeId` | Path: `scopeId` (UUID) | `EvaluatorScopeResponse` | — |

### Request/Response Schemas

#### `OrgTreeListResponse`
```json
{
  "data": [
    { "id": "uuid", "name": "string", "type": "corporate|retail", "nodeCount": 0 }
  ]
}
```

#### `OrgNodeListResponse` (flat)
```json
{
  "data": [
    {
      "id": "uuid", "parentId": "uuid|null", "name": "string", "type": "corporate|retail",
      "code": "string", "depth": 0, "path": "ltree_path", "employeeCount": 0
    }
  ],
  "meta": { "format": "flat", "total": 0 }
}
```

#### `OrgNodeListResponse` (nested)
```json
{
  "data": {
    "id": "uuid", "name": "string", "children": [ { "id": "uuid", ... } ]
  },
  "meta": { "format": "nested" }
}
```

#### `CreateOrgNodeRequest`
```json
{
  "organizationId": "uuid",
  "parentId": "uuid|null",
  "name": "string (required, 1-255 chars)",
  "type": "corporate|retail",
  "code": "string (unique within org)",
  "metadata": {}
}
```

#### `UpdateOrgNodeRequest`
```json
{
  "name": "string (optional)",
  "code": "string (optional)",
  "metadata": {},
  "version": 0
}
```

#### `MoveOrgNodeRequest`
```json
{ "newParentId": "uuid" }
```

#### `EmployeeListResponse`
```json
{
  "data": [
    {
      "id": "uuid", "firstName": "string", "lastName": "string", "email": "string",
      "employeeNumber": "string", "orgNodeId": "uuid", "managerId": "uuid|null",
      "profileId": "uuid", "isActive": true
    }
  ],
  "meta": { "nextCursor": "string|null", "limit": 50 }
}
```

#### `EmployeeDetailResponse`
```json
{
  "id": "uuid", "firstName": "string", "lastName": "string", "email": "string",
  "employeeNumber": "string", "orgNodeId": "uuid", "managerId": "uuid|null",
  "profileId": "uuid", "isActive": true,
  "orgNode": { "id": "uuid", "name": "string", "path": "ltree" },
  "manager": { "id": "uuid", "firstName": "string", "lastName": "string" } | null
}
```

#### `AncestorChainResponse`
```json
{
  "data": [
    { "id": "uuid", "name": "string", "depth": 0, "relation": "self|direct_manager|director|vp|ceo" }
  ]
}
```

#### `EvaluatorScopeResponse`
```json
{
  "evaluatorId": "uuid",
  "cycleId": "uuid|null",
  "scopeType": "department|team|individual",
  "scopeData": { "orgNodeIds": ["uuid"], "employeeIds": ["uuid"] },
  "evaluateeCount": 0
}
```

### Error Mapping

| Error Code | HTTP | Source | Handler Action |
|------------|------|--------|----------------|
| `TREE_NOT_FOUND` | 404 | Repo: `ent.IsNotFound` on OrgTree | Return 404 with treeId in details |
| `NODE_NOT_FOUND` | 404 | Repo: `ent.IsNotFound` on OrgNode | Return 404 with nodeId in details |
| `EMPLOYEE_NOT_FOUND` | 404 | Repo: `ent.IsNotFound` on Employee | Return 404 with empId in details |
| `NODE_HAS_CHILDREN` | 409 | Service: `CountChildren(nodeId) > 0` | Return 409, refuse delete |
| `INVALID_PARENT` | 400 | Service: cycle detection in `MoveNode` | Return 400, include cycle path in details |
| `STALE_VERSION` | 409 | Service: optimistic lock mismatch | Return 409, include current version in details |
| `INVALID_TREE_TYPE` | 400 | Handler: enum validation | Return 400 |
| `EMPLOYEE_IN_MULTIPLE_TREES` | 409 | Service: unique constraint on `(employee_number, org)` | Return 409 |
| `RATE_LIMIT_EXCEEDED` | 429 | Middleware | Return 429 with `Retry-After` header |

---

## 4. Service Layer — Tree Traversal

### Core Service Methods

```go
// OrgTreeService
type OrgTreeService interface {
    GetTrees(ctx context.Context, filter TreeFilter) ([]OrgTree, error)
    GetTree(ctx context.Context, treeID uuid.UUID) (*OrgTree, error)
    GetTreeNodes(ctx context.Context, treeID uuid.UUID, format NodeFormat, depth int) (*OrgNodeList, error)
    ExportTree(ctx context.Context, treeID uuid.UUID, w io.Writer) error
}

// OrgNodeService
type OrgNodeService interface {
    GetNode(ctx context.Context, nodeID uuid.UUID) (*OrgNode, error)
    CreateNode(ctx context.Context, req CreateOrgNodeRequest) (*OrgNode, error)
    UpdateNode(ctx context.Context, nodeID uuid.UUID, req UpdateOrgNodeRequest) (*OrgNode, error)
    DeleteNode(ctx context.Context, nodeID uuid.UUID) error
    MoveNode(ctx context.Context, nodeID uuid.UUID, newParentID uuid.UUID) (*OrgNode, error)
}

// EmployeeService
type EmployeeService interface {
    ListEmployees(ctx context.Context, filter EmployeeFilter) (*EmployeeList, error)
    GetEmployee(ctx context.Context, empID uuid.UUID) (*Employee, error)
    GetMyEvaluatees(ctx context.Context, evaluatorID uuid.UUID) ([]Employee, error)
    GetManager(ctx context.Context, empID uuid.UUID) (*Employee, error)
    GetChainOfCommand(ctx context.Context, empID uuid.UUID) ([]Ancestor, error)
    BatchLookup(ctx context.Context, ids []uuid.UUID) ([]Employee, error)
    SearchEmployees(ctx context.Context, query string, limit int) ([]Employee, error)
}

// EvaluatorService
type EvaluatorService interface {
    GetEvaluatorScope(ctx context.Context, evaluatorID, cycleID uuid.UUID) (*EvaluatorScope, error)
    ResolveEvaluator(ctx context.Context, evaluateeID uuid.UUID) (*Employee, error)
}
```

### Critical: Tree Traversal Strategy

**Decision: Use PostgreSQL `ltree` materialized path as primary strategy.**

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| **Recursive CTE** | No schema change, standard SQL | O(depth) per query; slow for >1000 nodes or >4 levels | Avoid for production hot paths |
| **`ltree`** | Native GIST/GIN indexes; ancestor/descendant in single index scan; path queries are fast | PostgreSQL-specific; requires extension | **Primary choice** |
| **Closure Table** | Portable across DBs; fast ancestor/descendant queries | Extra table to maintain; writes are O(n) for subtree moves | **Fallback / documented alternative** |

#### `ltree` Implementation

Add to `OrgNode` schema (extends C1 data-model-core):

```go
// Extension to OrgNode schema from data-model-core
fields:
  - path: ltree (materialized path, e.g., "1.3.7.12")
  - version: int (default 0, for optimistic locking)

indices:
  - (path) using GIST — ancestor/descendant queries
  - (path) using btree — exact path lookups, ordering
```

**Path generation rule:** `path = parent.path || '.' || node.id::text` (or use integer surrogate keys for shorter paths if UUIDs become unwieldy; document decision). For readability and debugging, use `node.id::text` — PostgreSQL handles 36-char UUID segments fine in `ltree` up to ~10 levels.

**Service-layer traversal methods:**

```go
// GetMyEvaluatees — direct reports only (single query, no tree needed)
func (s *employeeService) GetMyEvaluatees(ctx context.Context, evaluatorID uuid.UUID) ([]Employee, error) {
    // Cache key: org:evaluatees:{evaluatorID}:v1
    // Query: SELECT * FROM employees WHERE manager_id = $1 AND is_active = true
    // Index: idx_employee_manager_active
}

// GetChainOfCommand — path from employee to root using ltree ancestor query
func (s *employeeService) GetChainOfCommand(ctx context.Context, empID uuid.UUID) ([]Ancestor, error) {
    // 1. Get employee's org_node.path
    // 2. Query: SELECT * FROM org_nodes WHERE path @> $1 ORDER BY path
    //    (all ancestors of this node's path, including self)
    // 3. Join employees to find who sits at each node, or return node chain
}

// GetEvaluatorScope — lookup pre-computed scope or compute on-the-fly
func (s *evaluatorService) GetEvaluatorScope(ctx context.Context, evaluatorID, cycleID uuid.UUID) (*EvaluatorScope, error) {
    // 1. Try Redis: key = org:scope:{evaluatorID}:{cycleID}:v1, TTL 24h
    // 2. If miss: query evaluator_scopes table by (evaluator_id, cycle_id)
    //    Index: idx_evaluatorscope_eval_cycle
    // 3. If scope_type = "department", resolve all employees in scope_data.orgNodeIds
    //    using path @> ANY(...) on org_nodes, then join employees
    // 4. Cache result, set TTL
}

// ResolveEvaluator — who evaluates this person (manager lookup)
func (s *evaluatorService) ResolveEvaluator(ctx context.Context, evaluateeID uuid.UUID) (*Employee, error) {
    // Query: SELECT manager.* FROM employees e JOIN employees manager ON e.manager_id = manager.id
    // WHERE e.id = $1
}
```

#### Cycle Detection (MoveNode)

```go
func (s *orgNodeService) MoveNode(ctx context.Context, nodeID, newParentID uuid.UUID) (*OrgNode, error) {
    // 1. Advisory lock: pg_advisory_lock(hash(nodeID)) to prevent concurrent restructuring
    // 2. Fetch node and new parent
    // 3. Validate: newParentID is not nodeID or any descendant of nodeID
    //    Query: SELECT 1 FROM org_nodes WHERE id = $1 AND path @> (SELECT path FROM org_nodes WHERE id = $2)
    //    If returns row → cycle detected → error INVALID_PARENT
    // 4. Begin tx with DEFERRABLE INITIALLY DEFERRED
    // 5. Update node's parent_id
    // 6. Update node's path and all descendants' paths
    //    UPDATE org_nodes SET path = $1 || subpath(path, nlevel(old_path)) WHERE path <@ old_path
    // 7. Invalidate Redis: org:tree:*, org:evaluatees:*
    // 8. Commit
}
```

---

## 5. Repository Layer

### Ent Patterns for Self-Referential Relationships

```go
// orgnode_repo.go

type OrgNodeRepo struct {
    client *ent.Client
}

func (r *OrgNodeRepo) GetNodeWithChildren(ctx context.Context, nodeID uuid.UUID) (*ent.OrgNode, error) {
    return r.client.OrgNode.Query().
        Where(orgnode.ID(nodeID)).
        WithParent().
        WithChildren(func(q *ent.OrgNodeQuery) {
            q.Order(ent.Asc(orgnode.FieldName))
        }).
        Only(ctx)
}

func (r *OrgNodeRepo) GetDescendants(ctx context.Context, path string) ([]*ent.OrgNode, error) {
    // ltree query via custom predicate (Ent supports raw SQL predicates)
    return r.client.OrgNode.Query().
        Where(func(s *sql.Selector) {
            s.Where(sql.ExprP("path <@ $1", path))
        }).
        Order(ent.Asc(orgnode.FieldPath)).
        All(ctx)
}

func (r *OrgNodeRepo) GetAncestors(ctx context.Context, path string) ([]*ent.OrgNode, error) {
    return r.client.OrgNode.Query().
        Where(func(s *sql.Selector) {
            s.Where(sql.ExprP("path @> $1", path))
        }).
        Order(ent.Asc(orgnode.FieldPath)).
        All(ctx)
}

func (r *OrgNodeRepo) CountChildren(ctx context.Context, nodeID uuid.UUID) (int, error) {
    return r.client.OrgNode.Query().
        Where(orgnode.ParentID(nodeID)).
        Count(ctx)
}

func (r *OrgNodeRepo) UpdatePathAndDescendants(ctx context.Context, tx *ent.Tx, oldPath, newPath string) error {
    // Update all nodes whose path starts with oldPath
    _, err := tx.OrgNode.Update().
        Where(func(s *sql.Selector) {
            s.Where(sql.ExprP("path <@ $1", oldPath))
        }).
        SetPath(sql.Expr("$1 || subpath(path, nlevel($2))", newPath, oldPath)).
        Save(ctx)
    return err
}
```

### Employee Repository

```go
// employee_repo.go

func (r *EmployeeRepo) ListByManager(ctx context.Context, managerID uuid.UUID, activeOnly bool) ([]*ent.Employee, error) {
    q := r.client.Employee.Query().Where(employee.ManagerID(managerID))
    if activeOnly {
        q = q.Where(employee.IsActive(true))
    }
    return q.Order(ent.Asc(employee.FieldLastName), ent.Asc(employee.FieldFirstName)).All(ctx)
}

func (r *EmployeeRepo) Search(ctx context.Context, query string, limit int) ([]*ent.Employee, error) {
    // Use PostgreSQL tsvector or ILIKE with trigram index
    // CREATE INDEX idx_employee_search ON employees USING gin (to_tsvector('spanish', first_name || ' ' || last_name || ' ' || email));
    return r.client.Employee.Query().
        Where(func(s *sql.Selector) {
            s.Where(sql.ExprP(
                "to_tsvector('spanish', first_name || ' ' || last_name || ' ' || email) @@ plainto_tsquery('spanish', $1)",
                query,
            ))
        }).
        Limit(limit).
        All(ctx)
}

func (r *EmployeeRepo) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*ent.Employee, error) {
    return r.client.Employee.Query().
        Where(employee.IDIn(ids...)).
        All(ctx)
}
```

### Efficient Ancestor/Descendant Queries

**With `ltree`:**
- Ancestors of node with path `1.3.7.12`: `WHERE path @> '1.3.7.12'` (GIST index)
- Descendants of node with path `1.3`: `WHERE path <@ '1.3'` (GIST index)
- Depth filter: `WHERE nlevel(path) <= maxDepth`

**With closure table (fallback):**
```sql
-- Table: org_node_closure (ancestor_id, descendant_id, depth)
CREATE UNIQUE INDEX idx_closure_ancestor_desc ON org_node_closure(ancestor_id, descendant_id);
CREATE INDEX idx_closure_descendant ON org_node_closure(descendant_id);

-- Ancestors:
SELECT ancestor_id FROM org_node_closure WHERE descendant_id = $1 ORDER BY depth DESC;
-- Descendants:
SELECT descendant_id FROM org_node_closure WHERE ancestor_id = $1;
```

---

## 6. Caching Strategy

### Redis Cache Design

**Cache keys (all prefixed with environment, e.g., `sed:prod:`):**

| Key Pattern | Value | TTL | Invalidation Trigger |
|-------------|-------|-----|----------------------|
| `org:tree:{treeId}:v1` | Serialized full tree (flat nodes) | 1h | Any node create/update/move/delete in tree |
| `org:tree:{treeId}:etag` | ETag string (`tree-{treeId}-{lastModified}`) | 1h | Same as above |
| `org:evaluatees:{empId}:v1` | Array of direct-report employee IDs | 24h | Employee manager_id change, employee activation/deactivation |
| `org:scope:{evaluatorId}:{cycleId}:v1` | EvaluatorScope JSON | 24h | EvaluatorScope table change, employee transfer |
| `org:ancestors:{empId}:v1` | Array of ancestor node IDs | 12h | Employee org_node_id change, tree restructuring |
| `org:employee:{empId}:v1` | Single employee JSON | 1h | Employee update |
| `org:search:{hash(q)}:v1` | Search results (employee IDs) | 5m | Employee name/email change (rare, short TTL acceptable) |

### Cache Invalidation

```go
// InvalidateTreeCache removes all cached entries for a given tree.
func (c *OrgCache) InvalidateTreeCache(ctx context.Context, treeID uuid.UUID) error {
    pipe := c.redis.Pipeline()
    pipe.Del(ctx, fmt.Sprintf("org:tree:%s:v1", treeID))
    pipe.Del(ctx, fmt.Sprintf("org:tree:%s:etag", treeID))
    // Pattern delete for evaluatees and ancestors would require SCAN; instead,
    // use a cache generation counter per tree: org:tree:{treeId}:gen
    // Include generation in evaluatee/ancestor keys; increment gen on tree change.
    pipe.Incr(ctx, fmt.Sprintf("org:tree:%s:gen", treeID))
    _, err := pipe.Exec(ctx)
    return err
}
```

**Generation-based invalidation:** Instead of deleting individual keys, increment a generation counter per tree. All dependent keys embed the generation number. This turns invalidation into a single `INCR` instead of `SCAN` + multiple `DEL`.

### ETag on Tree Structure

```go
func (h *OrgTreeHandler) GetTreeNodes(w http.ResponseWriter, r *http.Request) {
    treeID := chi.URLParam(r, "treeId")
    etag, _ := h.cache.GetETag(r.Context(), treeID)
    if match := r.Header.Get("If-None-Match"); match == etag {
        w.WriteHeader(http.StatusNotModified)
        return
    }

    nodes, err := h.service.GetTreeNodes(r.Context(), treeID, ...)
    // ...
    w.Header().Set("ETag", etag)
    w.Header().Set("Cache-Control", "max-age=3600, must-revalidate")
}
```

### Cache Warming on Startup

```go
// In main.go or server initialization
func warmOrgCache(ctx context.Context, svc OrgTreeService, cache *OrgCache) error {
    trees, err := svc.GetTrees(ctx, TreeFilter{})
    if err != nil {
        return err
    }
    for _, tree := range trees {
        _, _ = svc.GetTreeNodes(ctx, tree.ID, NodeFormatFlat, 0) // triggers cache fill
    }
    return nil
}
```

---

## 7. OpenAPI 3.1 Spec

```yaml
openapi: 3.1.0
info:
  title: SED Org Hierarchy API
  version: 1.0.0
  description: Organizational hierarchy, employees, and evaluator scope API for SED.

servers:
  - url: /api/v1

paths:
  /org-trees:
    get:
      operationId: listOrgTrees
      parameters:
        - name: type
          in: query
          schema:
            type: string
            enum: [corporate, retail]
      responses:
        '200':
          description: List of org trees
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/OrgTree'

  /org-trees/{treeId}:
    get:
      operationId: getOrgTree
      parameters:
        - name: treeId
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Org tree detail
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/OrgTree'
        '404':
          $ref: '#/components/responses/TreeNotFound'

  /org-trees/{treeId}/nodes:
    get:
      operationId: getOrgTreeNodes
      parameters:
        - name: treeId
          in: path
          required: true
          schema: { type: string, format: uuid }
        - name: format
          in: query
          schema:
            type: string
            enum: [flat, nested]
            default: flat
        - name: depth
          in: query
          schema:
            type: integer
            default: -1
      responses:
        '200':
          description: Tree nodes
          content:
            application/json:
              schema:
                oneOf:
                  - $ref: '#/components/schemas/OrgNodeFlatList'
                  - $ref: '#/components/schemas/OrgNodeNested'

  /org-trees/{treeId}/export:
    get:
      operationId: exportOrgTree
      parameters:
        - name: treeId
          in: path
          required: true
          schema: { type: string, format: uuid }
      responses:
        '200':
          description: Streaming JSON export
          headers:
            Transfer-Encoding:
              schema: { type: string }
          content:
            application/json:
              schema:
                type: object
                properties:
                  nodes:
                    type: array
                    items:
                      $ref: '#/components/schemas/OrgNode'

  /org-nodes:
    post:
      operationId: createOrgNode
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrgNodeRequest'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/OrgNode'
        '400':
          $ref: '#/components/responses/BadRequest'

  /org-nodes/{nodeId}:
    get:
      operationId: getOrgNode
      parameters:
        - name: nodeId
          in: path
          required: true
          schema: { type: string, format: uuid }
      responses:
        '200':
          description: Node detail
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/OrgNode'
        '404':
          $ref: '#/components/responses/NodeNotFound'
    put:
      operationId: updateOrgNode
      parameters:
        - name: nodeId
          in: path
          required: true
          schema: { type: string, format: uuid }
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateOrgNodeRequest'
      responses:
        '200':
          description: Updated
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/OrgNode'
        '409':
          $ref: '#/components/responses/Conflict'
    delete:
      operationId: deleteOrgNode
      parameters:
        - name: nodeId
          in: path
          required: true
          schema: { type: string, format: uuid }
      responses:
        '204':
          description: Deleted
        '409':
          $ref: '#/components/responses/NodeHasChildren'

  /org-nodes/{nodeId}/move:
    post:
      operationId: moveOrgNode
      parameters:
        - name: nodeId
          in: path
          required: true
          schema: { type: string, format: uuid }
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/MoveOrgNodeRequest'
      responses:
        '200':
          description: Moved
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/OrgNode'
        '400':
          $ref: '#/components/responses/BadRequest'

  /employees:
    get:
      operationId: listEmployees
      parameters:
        - name: treeId
          in: query
          schema: { type: string, format: uuid }
        - name: nodeId
          in: query
          schema: { type: string, format: uuid }
        - name: profileId
          in: query
          schema: { type: string, format: uuid }
        - name: isActive
          in: query
          schema: { type: boolean }
        - name: q
          in: query
          schema: { type: string }
        - name: cursor
          in: query
          schema: { type: string }
        - name: limit
          in: query
          schema: { type: integer, default: 50, maximum: 200 }
      responses:
        '200':
          description: Employee list
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/Employee'
                  meta:
                    $ref: '#/components/schemas/PaginationMeta'

  /employees/{empId}:
    get:
      operationId: getEmployee
      parameters:
        - name: empId
          in: path
          required: true
          schema: { type: string, format: uuid }
      responses:
        '200':
          description: Employee detail
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/EmployeeDetail'
        '404':
          $ref: '#/components/responses/EmployeeNotFound'

  /employees/{empId}/evaluatees:
    get:
      operationId: getMyEvaluatees
      parameters:
        - name: empId
          in: path
          required: true
          schema: { type: string, format: uuid }
      responses:
        '200':
          description: Direct reports
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/Employee'
        '404':
          $ref: '#/components/responses/EmployeeNotFound'

  /employees/{empId}/manager:
    get:
      operationId: getManager
      parameters:
        - name: empId
          in: path
          required: true
          schema: { type: string, format: uuid }
      responses:
        '200':
          description: Manager detail
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/Employee'
        '404':
          $ref: '#/components/responses/EmployeeNotFound'

  /employees/{empId}/ancestors:
    get:
      operationId: getChainOfCommand
      parameters:
        - name: empId
          in: path
          required: true
          schema: { type: string, format: uuid }
      responses:
        '200':
          description: Chain of command
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/Ancestor'
        '404':
          $ref: '#/components/responses/EmployeeNotFound'

  /employees/batch:
    post:
      operationId: batchLookupEmployees
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                ids:
                  type: array
                  items: { type: string, format: uuid }
                  maxItems: 100
      responses:
        '200':
          description: Batch results
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/Employee'

  /employees/search:
    get:
      operationId: searchEmployees
      parameters:
        - name: q
          in: query
          required: true
          schema: { type: string, minLength: 2 }
        - name: limit
          in: query
          schema: { type: integer, default: 20, maximum: 50 }
      responses:
        '200':
          description: Search results
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/Employee'

  /evaluator-scopes:
    get:
      operationId: getEvaluatorScope
      parameters:
        - name: evaluatorId
          in: query
          required: true
          schema: { type: string, format: uuid }
        - name: cycleId
          in: query
          schema: { type: string, format: uuid }
      responses:
        '200':
          description: Evaluator scope
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/EvaluatorScope'
        '404':
          $ref: '#/components/responses/EmployeeNotFound'

components:
  schemas:
    OrgTree:
      type: object
      properties:
        id: { type: string, format: uuid }
        name: { type: string }
        type: { type: string, enum: [corporate, retail] }
        nodeCount: { type: integer }

    OrgNode:
      type: object
      properties:
        id: { type: string, format: uuid }
        parentId: { type: string, format: uuid, nullable: true }
        name: { type: string }
        type: { type: string, enum: [corporate, retail] }
        code: { type: string }
        depth: { type: integer }
        path: { type: string }
        employeeCount: { type: integer }
        metadata: { type: object, additionalProperties: true }

    OrgNodeFlatList:
      type: object
      properties:
        data:
          type: array
          items:
            $ref: '#/components/schemas/OrgNode'
        meta:
          type: object
          properties:
            format: { type: string, enum: [flat] }
            total: { type: integer }

    OrgNodeNested:
      type: object
      properties:
        data:
          $ref: '#/components/schemas/OrgNode'
        meta:
          type: object
          properties:
            format: { type: string, enum: [nested] }

    CreateOrgNodeRequest:
      type: object
      required: [organizationId, name, type, code]
      properties:
        organizationId: { type: string, format: uuid }
        parentId: { type: string, format: uuid, nullable: true }
        name: { type: string, minLength: 1, maxLength: 255 }
        type: { type: string, enum: [corporate, retail] }
        code: { type: string, minLength: 1, maxLength: 100 }
        metadata: { type: object, additionalProperties: true }

    UpdateOrgNodeRequest:
      type: object
      properties:
        name: { type: string, minLength: 1, maxLength: 255 }
        code: { type: string, minLength: 1, maxLength: 100 }
        metadata: { type: object, additionalProperties: true }
        version: { type: integer }

    MoveOrgNodeRequest:
      type: object
      required: [newParentId]
      properties:
        newParentId: { type: string, format: uuid }

    Employee:
      type: object
      properties:
        id: { type: string, format: uuid }
        firstName: { type: string }
        lastName: { type: string }
        email: { type: string, format: email }
        employeeNumber: { type: string }
        orgNodeId: { type: string, format: uuid }
        managerId: { type: string, format: uuid, nullable: true }
        profileId: { type: string, format: uuid }
        isActive: { type: boolean }

    EmployeeDetail:
      allOf:
        - $ref: '#/components/schemas/Employee'
        - type: object
          properties:
            orgNode:
              type: object
              properties:
                id: { type: string, format: uuid }
                name: { type: string }
                path: { type: string }
            manager:
              type: object
              nullable: true
              properties:
                id: { type: string, format: uuid }
                firstName: { type: string }
                lastName: { type: string }

    PaginationMeta:
      type: object
      properties:
        nextCursor: { type: string, nullable: true }
        limit: { type: integer }

    Ancestor:
      type: object
      properties:
        id: { type: string, format: uuid }
        name: { type: string }
        depth: { type: integer }
        relation: { type: string, enum: [self, direct_manager, director, vp, ceo] }

    EvaluatorScope:
      type: object
      properties:
        evaluatorId: { type: string, format: uuid }
        cycleId: { type: string, format: uuid, nullable: true }
        scopeType: { type: string, enum: [department, team, individual] }
        scopeData:
          type: object
          properties:
            orgNodeIds:
              type: array
              items: { type: string, format: uuid }
            employeeIds:
              type: array
              items: { type: string, format: uuid }
        evaluateeCount: { type: integer }

    Error:
      type: object
      properties:
        code: { type: string }
        message: { type: string }
        details:
          type: array
          items: { type: string }
        traceId: { type: string }

  responses:
    TreeNotFound:
      description: Tree not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            code: TREE_NOT_FOUND
            message: Organizational tree not found
            details: ["treeId: 550e8400-e29b-41d4-a716-446655440000"]
    NodeNotFound:
      description: Node not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    EmployeeNotFound:
      description: Employee not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    BadRequest:
      description: Invalid request
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    Conflict:
      description: Conflict (optimistic lock or business rule)
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    NodeHasChildren:
      description: Cannot delete node with children
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
```

---

## 8. Concurrency Details

### Advisory Locks on Tree Restructuring

PostgreSQL advisory locks (session-level, non-transactional) prevent concurrent tree mutations from corrupting the materialized path.

```go
func (r *OrgNodeRepo) AcquireTreeLock(ctx context.Context, treeID uuid.UUID) (func(), error) {
    // Use a 64-bit hash of treeID as the lock ID
    lockID := hashUUID(treeID)
    _, err := r.client.ExecContext(ctx, "SELECT pg_advisory_lock($1)", lockID)
    if err != nil {
        return nil, err
    }
    release := func() {
        _, _ = r.client.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", lockID)
    }
    return release, nil
}
```

**Usage in MoveNode service:**
1. Acquire advisory lock for the tree.
2. Defer release.
3. Begin Ent transaction.
4. Perform cycle check.
5. Update paths.
6. Invalidate cache (within tx or after commit — cache invalidation must not roll back DB changes).

### Employee Transfer Transaction Pattern

```go
func (s *employeeService) TransferEmployee(ctx context.Context, empID, newOrgNodeID, newManagerID uuid.UUID) error {
    tx, err := s.client.Tx(ctx)
    if err != nil {
        return err
    }
    defer func() {
        if r := recover(); r != nil {
            _ = tx.Rollback()
            panic(r)
        }
    }()

    emp, err := tx.Employee.UpdateOneID(empID).
        SetOrgNodeID(newOrgNodeID).
        SetManagerID(newManagerID).
        Save(ctx)
    if err != nil {
        _ = tx.Rollback()
        return err
    }

    if err := tx.Commit(); err != nil {
        return err
    }

    // Cache invalidation AFTER successful commit
    s.cache.InvalidateEmployee(ctx, empID)
    s.cache.InvalidateEvaluatees(ctx, newManagerID)
    if emp.ManagerID != uuid.Nil {
        s.cache.InvalidateEvaluatees(ctx, emp.ManagerID)
    }
    return nil
}
```

### Optimistic Locking on OrgNode Updates

```go
func (r *OrgNodeRepo) UpdateWithVersion(ctx context.Context, nodeID uuid.UUID, version int, mutator func(*ent.OrgNodeUpdateOne)) (*ent.OrgNode, error) {
    node, err := r.client.OrgNode.Query().Where(orgnode.ID(nodeID)).Only(ctx)
    if err != nil {
        return nil, err
    }
    if node.Version != version {
        return nil, ErrStaleVersion
    }

    upd := r.client.OrgNode.UpdateOne(node)
    mutator(upd)
    upd.SetVersion(version + 1)
    return upd.Save(ctx)
}
```

**Error mapping:** `ErrStaleVersion` → HTTP 409 with code `STALE_VERSION`.

---

## 9. Performance Optimization

### Materialized Path (`ltree`)

```sql
-- Enable extension
CREATE EXTENSION IF NOT EXISTS ltree;

-- Add path column to org_nodes (migration)
ALTER TABLE org_nodes ADD COLUMN path ltree;
ALTER TABLE org_nodes ADD COLUMN version int NOT NULL DEFAULT 0;

-- Indexes
CREATE INDEX idx_orgnode_path_gist ON org_nodes USING GIST (path);
CREATE INDEX idx_orgnode_path_btree ON org_nodes USING btree (path);

-- Query patterns:
-- Ancestors (including self):
--   SELECT * FROM org_nodes WHERE path @> '1.3.7.12' ORDER BY path;
-- Descendants (including self):
--   SELECT * FROM org_nodes WHERE path <@ '1.3' ORDER BY path;
-- Depth-limited:
--   SELECT * FROM org_nodes WHERE path <@ '1.3' AND nlevel(path) <= 3;
```

### Closure Table (Alternative)

If `ltree` is unavailable (e.g., managed DB without extension support):

```sql
CREATE TABLE org_node_closure (
    ancestor_id uuid REFERENCES org_nodes(id) ON DELETE CASCADE,
    descendant_id uuid REFERENCES org_nodes(id) ON DELETE CASCADE,
    depth int NOT NULL,
    PRIMARY KEY (ancestor_id, descendant_id)
);

CREATE INDEX idx_closure_descendant ON org_node_closure(descendant_id);

-- Ancestors:
--   SELECT ancestor_id FROM org_node_closure WHERE descendant_id = $1 ORDER BY depth DESC;
-- Descendants:
--   SELECT descendant_id FROM org_node_closure WHERE ancestor_id = $1;
-- Insert subtree: insert closures for all ancestors of new parent + all descendants of moved node
```

### Denormalized Ancestor Chain for Hot Paths

For frequently queried chains of command (executives, HR dashboards), pre-compute and cache in Redis rather than querying the DB every time.

### Index Recommendations

| Index | Table | Columns | Type | Purpose |
|-------|-------|---------|------|---------|
| `idx_orgnode_path_gist` | `org_nodes` | `path` | GIST | Ancestor/descendant ltree queries |
| `idx_orgnode_path_btree` | `org_nodes` | `path` | btree | Path ordering, exact match |
| `idx_orgnode_org_parent` | `org_nodes` | `(organization_id, parent_id)` | btree | Children listing |
| `idx_orgnode_org_type` | `org_nodes` | `(organization_id, type)` | btree | Filter by tree type |
| `idx_employee_manager_active` | `employees` | `(manager_id, is_active)` | btree | My evaluatees |
| `idx_employee_org_profile` | `employees` | `(org_node_id, profile_id)` | btree | Filter by node + profile |
| `idx_employee_search` | `employees` | `to_tsvector('spanish', first_name \|\| ' ' \|\| last_name \|\| ' ' \|\| email)` | GIN | Full-text search |
| `idx_employee_number` | `employees` | `(employee_number)` | btree | Exact lookup |
| `idx_evaluatorscope_eval_cycle` | `evaluator_scopes` | `(evaluator_id, cycle_id)` | btree | Scope lookup |
| `idx_evaluatorscope_data_gin` | `evaluator_scopes` | `scope_data` | GIN | Query inside scope_data JSONB |

### Read Replicas

All read endpoints (`GET`) should use a read-only Ent client pointing to PostgreSQL replicas. Write endpoints (`POST`, `PUT`, `DELETE`) use the primary.

```go
func (r *OrgNodeRepo) readClient() *ent.Client {
    if r.readReplica != nil {
        return r.readReplica
    }
    return r.client
}
```

---

## 10. Testing Strategy

### Test Pyramid

| Layer | Tool | Coverage Target | Key Tests |
|-------|------|-----------------|-----------|
| Handler | `net/http/httptest` + Chi | 80% | Table-driven HTTP tests for all 14 endpoints; error code mapping |
| Service | Go test + mocked repos | 85% | Cycle detection, cache invalidation logic, scope computation |
| Repository | Testcontainers (PostgreSQL) | 80% | Ent queries against real DB; ltree correctness |
| Integration | Testcontainers + Redis | 60% | End-to-end: create tree → move node → query evaluatees |
| Performance | `go test -bench` + `k6` (future) | — | Deep tree (1000+ nodes), concurrent moves |

### Tree Traversal Correctness Tests

```go
func TestGetChainOfCommand(t *testing.T) {
    ctx := context.Background()
    // Build tree: CEO(1) → VP(2) → Director(3) → Manager(4) → Employee(5)
    // Expected chain for Employee(5): [5, 4, 3, 2, 1]
    chain, err := svc.GetChainOfCommand(ctx, emp5.ID)
    require.NoError(t, err)
    require.Len(t, chain, 5)
    assert.Equal(t, "self", chain[0].Relation)
    assert.Equal(t, "ceo", chain[4].Relation)
}

func TestCycleDetection(t *testing.T) {
    ctx := context.Background()
    // Try to move node A under node B, where B is already a descendant of A
    err := svc.MoveNode(ctx, nodeA.ID, nodeB.ID)
    assert.ErrorIs(t, err, ErrInvalidParent)
}
```

### Deep Tree Performance Tests

```go
func BenchmarkGetTreeNodes_1000Nodes(b *testing.B) {
    ctx := context.Background()
    treeID := seedTree(b, 1000, 4) // 1000 nodes, 4 levels deep
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = svc.GetTreeNodes(ctx, treeID, NodeFormatFlat, 0)
    }
}
// Target: < 50ms per call for flat format, < 100ms for nested.
```

### Concurrent Tree Modification Tests

```go
func TestConcurrentNodeMoves(t *testing.T) {
    ctx := context.Background()
    // Two goroutines attempt to move different subtrees simultaneously
    // Verify: no orphaned nodes, all paths remain valid ltree strings,
    // advisory locks serialize conflicting moves.
}
```

### Test Data Fixtures

Create a reusable fixture package `internal/testfixtures/` that seeds:
- A corporate tree (CEO → 2 VPs → 4 Directors → 8 Managers → 32 Employees)
- A retail tree (Regional → Store → Department → Associate)
- Inactive employees and nodes for edge cases

### Performance Acceptance Criteria

| Scenario | Target | Measurement |
|----------|--------|-------------|
| `GET /org-trees/:id/nodes` (10k nodes, flat) | < 50ms p99 | Benchmark + load test |
| `GET /employees/:id/evaluatees` (cached) | < 30ms p99 | Benchmark |
| `GET /employees/:id/evaluatees` (uncached) | < 100ms p99 | Benchmark |
| `GET /employees` list (2000 req/s) | < 1% error rate | Load test |
| Tree ETag 304 rate | > 80% of tree reads | Production metric |
| Batch lookup (100 IDs) | < 50ms p99 | Benchmark |
| Employee search | < 50ms p99 | Benchmark |

---

## Appendix A: Schema Extension from C1

The C1 `data-model-core` schema is extended with two fields for C5:

**`OrgNode` additions:**
- `path` (`ltree`): materialized path for fast traversal.
- `version` (`int`, default 0): optimistic locking counter.

No other schema changes are required. All other entities (`Employee`, `EvaluatorScope`, `Organization`) remain exactly as defined in C1.

## Appendix B: Migration Notes

1. Add `ltree` extension in a separate migration before schema changes.
2. Backfill `path` for existing nodes using recursive CTE (one-time migration).
3. Create GIST and btree indexes on `path`.
4. If using closure table fallback, create `org_node_closure` and populate it.
5. Add `version` column with default 0; no backfill needed (existing rows start at version 0).
6. All migrations are **versioned** and **repeatable** — no manual edits after apply.

## Appendix C: Decision Log

| Decision | Rationale |
|----------|-----------|
| `ltree` over recursive CTE | Sub-50ms traversal at 10k nodes; GIST index handles ancestor/descendant in one scan. |
| `ltree` over closure table | Less write amplification on node moves; single column vs. extra table. |
| Optimistic locking on OrgNode | Rare writes, high read concurrency; advisory locks for structural ops, version for content updates. |
| Generation-based cache invalidation | Avoids Redis `SCAN` + multiple `DEL`; single `INCR` invalidates all tree-dependent keys. |
| Read replica for all GETs | 95% read workload; primary only for writes and cache misses. |
| ETag + 304 on tree reads | Reduces DB load by 80% for stable org structures (common case). |

---

*End of design document.*
