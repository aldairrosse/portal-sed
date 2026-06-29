## ADDED Requirements

### Requirement: Area metrics endpoint in OpenAPI

The system SHALL expose `GET /org-nodes/{nodeId}/area-metrics` in the OpenAPI spec `org-hierarchy.yaml`. The endpoint SHALL accept `nodeId` (path) and `cycleId` (query, optional). The response SHALL include aggregated metrics for direct employees of the selected node.

#### Scenario: Area metrics for mid-year phase

- **WHEN** frontend calls `GET /org-nodes/{nodeId}/area-metrics?cycleId={cycleId}` where cycle phase is `medio-anio`
- **THEN** response 200 includes:
  ```json
  {
    "nodeId": "uuid",
    "employeeCount": 4,
    "employeesWithGoals": 3,
    "avgProgress": 67.5,
    "completedGoals": 8,
    "pendingGoals": 12,
    "avgRating": null,
    "ratingsCount": 0
  }
  ```
- **AND** `avgRating` is null
- **AND** `ratingsCount` is 0

#### Scenario: Area metrics for end-of-year phase

- **WHEN** frontend calls `GET /org-nodes/{nodeId}/area-metrics?cycleId={cycleId}` where cycle phase is `fin-anio`
- **THEN** response 200 includes:
  ```json
  {
    "nodeId": "uuid",
    "employeeCount": 5,
    "employeesWithGoals": 0,
    "avgProgress": null,
    "completedGoals": 0,
    "pendingGoals": 0,
    "avgRating": 4.2,
    "ratingsCount": 5
  }
  ```
- **AND** `avgProgress` is null
- **AND** `completedGoals` and `pendingGoals` are 0

#### Scenario: Area metrics without cycleId

- **WHEN** frontend calls `GET /org-nodes/{nodeId}/area-metrics` without cycleId
- **THEN** response 200 returns metrics computed across all available data
- **AND** both `avgProgress` and `avgRating` may be non-null if data exists

#### Scenario: Area metrics with employees array

- **WHEN** frontend calls `GET /org-nodes/{nodeId}/area-metrics?cycleId={cycleId}`
- **THEN** response includes `employees` array with light employee projections:
  ```json
  {
    "employees": [
      { "id": "uuid", "firstName": "Juan", "lastName": "Pérez", "profileId": "uuid" }
    ]
  }
  ```
- **AND** employees list only includes direct employees of the node (not descendants)

### Requirement: TypeScript types for org hierarchy

The frontend SHALL generate TypeScript types from the OpenAPI spec `org-hierarchy.yaml` using `openapi-typescript`. The generated types SHALL be placed at `web/src/lib/api/schemas/org-hierarchy.d.ts`.

#### Scenario: Types generated successfully

- **WHEN** developer runs `openapi-typescript api/openapi/org-hierarchy.yaml -o web/src/lib/api/schemas/org-hierarchy.d.ts`
- **THEN** types file is created with all schemas from org-hierarchy.yaml
- **AND** types include `AreaMetrics`, `EvaluatorScope`, `OrgNode`, `Employee`
- **AND** `AreaMetrics` schema includes all fields: `nodeId`, `employeeCount`, `employeesWithGoals`, `avgProgress`, `completedGoals`, `pendingGoals`, `avgRating`, `ratingsCount`, `employees`

#### Scenario: Client uses org hierarchy types

- **WHEN** `web/src/lib/api/client.ts` imports `OrgHierarchyPaths` from `./schemas/org-hierarchy.d.ts`
- **THEN** `AppPaths` type includes all org hierarchy endpoints
- **AND** `client.GET('/org-trees')` is fully typed
- **AND** `client.GET('/org-nodes/{nodeId}/area-metrics')` is fully typed
- **AND** `client.GET('/evaluator-scopes')` is fully typed
- **AND** `client.GET('/employees/{empId}/evaluatees')` is fully typed

### Requirement: OpenAPI schema for AreaMetrics

The OpenAPI spec SHALL define the `AreaMetrics` schema in `components.schemas`:

```yaml
AreaMetrics:
  type: object
  properties:
    nodeId:
      type: string
      format: uuid
    employeeCount:
      type: integer
    employeesWithGoals:
      type: integer
    avgProgress:
      type: number
      format: float
      nullable: true
    completedGoals:
      type: integer
    pendingGoals:
      type: integer
    avgRating:
      type: number
      format: float
      nullable: true
    ratingsCount:
      type: integer
    employees:
      type: array
      items:
        $ref: '#/components/schemas/AreaMetricsEmployee'

AreaMetricsEmployee:
  type: object
  properties:
    id:
      type: string
      format: uuid
    firstName:
      type: string
    lastName:
      type: string
    profileId:
      type: string
      format: uuid
```

#### Scenario: Schema matches handler response

- **WHEN** backend handler `GetAreaMetrics` returns `dto.AreaMetricsResponse`
- **THEN** the JSON serialization matches the `AreaMetrics` schema exactly
- **AND** all required fields are present and non-optional in the response

### Requirement: Path item for area-metrics endpoint

The OpenAPI spec SHALL add a path item under `/org-nodes/{nodeId}/area-metrics`:

```yaml
/org-nodes/{nodeId}/area-metrics:
  get:
    operationId: getAreaMetrics
    summary: Get area metrics
    description: Returns aggregated metrics for the direct employees of an org node.
    parameters:
      - name: nodeId
        in: path
        required: true
        schema:
          type: string
          format: uuid
      - name: cycleId
        in: query
        required: false
        schema:
          type: string
          format: uuid
          description: Optional cycle to scope metrics to
    responses:
      '200':
        description: Successful response
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/AreaMetrics'
      '404':
        $ref: '#/components/responses/NodeNotFound'
      '400':
        $ref: '#/components/responses/BadRequest'
```

### Requirement: Endpoint error responses

The `GET /org-nodes/{nodeId}/area-metrics` endpoint SHALL return the following error responses:

#### Scenario: Node not found

- **WHEN** `nodeId` does not match any existing org node
- **THEN** response 404 with:
  ```json
  {
    "error": {
      "code": "NODE_NOT_FOUND",
      "message": "Org node not found",
      "trace_id": "abc12345"
    }
  }
  ```

#### Scenario: Invalid UUID format

- **WHEN** `nodeId` is not a valid UUID
- **THEN** response 400 with:
  ```json
  {
    "error": {
      "code": "INVALID_REQUEST",
      "message": "nodeId path parameter must be a valid UUID",
      "trace_id": "abc12345"
    }
  }
  ```

#### Scenario: Invalid cycleId format

- **WHEN** `cycleId` query parameter is provided but not a valid UUID
- **THEN** response 400 with invalid request error

#### Scenario: Service unavailable

- **WHEN** metrics service returns an unexpected error
- **THEN** response 500 with generic error (no internal details exposed)

## Acceptance Criteria

1. `org-hierarchy.yaml` includes `AreaMetrics` and `AreaMetricsEmployee` schemas
2. `org-hierarchy.yaml` includes path item for `GET /org-nodes/{nodeId}/area-metrics`
3. `openapi-typescript` generates `org-hierarchy.d.ts` with `AreaMetrics` type
4. `client.ts` includes `OrgHierarchyPaths` in `AppPaths` type
5. All existing endpoints in `org-hierarchy.yaml` remain unchanged
6. Generated types compile without errors (`pnpm run check` passes)
7. Handler response matches OpenAPI schema (validated by contract test)
