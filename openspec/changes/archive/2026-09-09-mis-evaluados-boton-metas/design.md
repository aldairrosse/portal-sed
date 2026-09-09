# Design: Mis-Evaluados Status, Validación de Pesos y Navegación de Metas por Época

## Architecture Overview

```
                                  ┌─────────────────────────────────────────┐
                                  │   web/src/routes/mis-evaluados          │
                                  │   (Tabla de Evaluados con Status)       │
                                  └────────────────────┬────────────────────┘
                                                       │
                                        ¿Época / Status?
                                       /                \
                       Inicio de año / Borrador       Medio-Fin de año / Enviada
                                     /                    \
                                    ▼                      ▼
┌─────────────────────────────────────────┐  ┌─────────────────────────────────────────┐
│ web/src/routes/objetivos/asignacion     │  │ web/src/routes/mis-evaluados/[id]/metas │
│ (?empId={id})                           │  │ (Vista de consulta y evaluación metas) │
│ - Formulación y asignación de metas     │  │ - Seguimiento de avances y evaluación   │
│ - Validación Double 100% (Pesos)        │  │ - Lectura de categorías/metas enviadas  │
└────────────────────┬────────────────────┘  └─────────────────────────────────────────┘
                     │ (Guardar / Enviar)
                     ▼
┌──────────────────────────────────────────────────────────────────────────────────────┐
│ Backend: POST /api/v1/employees/{empId}/assignments                                 │
│ 1. ValidateDoubleWeighting(ctx, empID)                                               │
│ 2. If Valid: status='enviada', submitted_at=now()                                    │
│ 3. If Invalid & Submit requested: Error 400 (WeightInvalid)                         │
│ 4. Persist in goal_assignments via Ent                                               │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 1. Database & Ent Schema Changes

### Migration SQL (`api/cmd/server/migrations/000039_add_goal_assignment_status_and_submitted_at.up.sql`)
```sql
-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'goal_assignment_status') THEN
        CREATE TYPE goal_assignment_status AS ENUM ('borrador', 'enviada');
    END IF;
END $$;

ALTER TABLE goal_assignments
    ADD COLUMN IF NOT EXISTS status goal_assignment_status NOT NULL DEFAULT 'borrador',
    ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE goal_assignments
    DROP COLUMN IF EXISTS submitted_at,
    DROP COLUMN IF EXISTS status;

DROP TYPE IF EXISTS goal_assignment_status;
-- +goose StatementEnd
```

### Ent Schema (`api/internal/schema/goalassignment.go`)
```go
func (GoalAssignment) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).
            Default(uuid.New).
            StorageKey("id"),
        field.UUID("employee_id", uuid.UUID{}),
        field.UUID("cycle_id", uuid.UUID{}),
        field.Enum("status").
            Values("borrador", "enviada").
            Default("borrador").
            SchemaType(map[string]string{
                dialect.Postgres: "goal_assignment_status",
            }),
        field.Time("submitted_at").
            Optional().
            Nillable(),
    }
}
```

---

## 2. Backend DTO, Repository & Service Layer

### DTO (`api/internal/dto/goal/goal_dto.go`)
```go
type AssignmentResponse struct {
    ID          string                 `json:"id"`
    EmployeeID  string                 `json:"employee_id"`
    CycleID     string                 `json:"cycle_id"`
    Status      string                 `json:"status"`
    SubmittedAt *string                `json:"submitted_at,omitempty"`
    Categories  []CategoryResponse     `json:"categories"`
    GlobalGoals []AssignedGoalResponse `json:"global_goals,omitempty"`
    SharedGoals []AssignedGoalResponse `json:"shared_goals,omitempty"`
    CreatedAt   string                 `json:"created_at"`
}
```

### Repository (`api/internal/repository/goal/assignment_repo.go`)
```go
type AssignmentRow struct {
    ID          uuid.UUID  `json:"id"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    EmployeeID  uuid.UUID  `json:"employee_id"`
    CycleID     uuid.UUID  `json:"cycle_id"`
    Status      string     `json:"status"`
    SubmittedAt *time.Time `json:"submitted_at,omitempty"`
}

// CreateOrSubmitAssignment inserts or updates the assignment with status validation
func (r *AssignmentRepo) CreateOrSubmitAssignment(ctx context.Context, empID, cycleID uuid.UUID, status string, submittedAt *time.Time) (*AssignmentRow, error)
```

### Handler (`api/internal/handler/goal/goal_handler.go`)
En `CreateAssignment`:
1. Resuelve `empID` y `cycleID`.
2. Ejecuta `h.weightSvc.ValidateDoubleWeighting(r.Context(), empID)`.
3. Si los pesos son válidos (100% categorías y 100% metas por categoría):
   - Establece `status = "enviada"` y `submitted_at = time.Now()`.
4. Si los pesos no son válidos pero se solicita creación base:
   - Establece `status = "borrador"` y `submitted_at = nil`.
5. Persiste mediante el repositorio y retorna el `AssignmentResponse` con su respectivo `status` y `submitted_at`.

---

## 3. OpenAPI Contract Updates (`api/openapi/goals-api.yaml`)

```yaml
AssignmentResponse:
  type: object
  properties:
    id:
      type: string
      format: uuid
    employee_id:
      type: string
      format: uuid
    cycle_id:
      type: string
      format: uuid
    status:
      type: string
      enum: [borrador, enviada]
      description: Estado actual de la asignación de metas
    submitted_at:
      type: string
      format: date-time
      nullable: true
      description: Fecha y hora en que la asignación fue enviada formalmente
    categories:
      type: array
      items:
        $ref: '#/components/schemas/CategoryResponse'
    global_goals:
      type: array
      items:
        $ref: '#/components/schemas/AssignedGoalResponse'
    shared_goals:
      type: array
      items:
        $ref: '#/components/schemas/AssignedGoalResponse'
    created_at:
      type: string
      format: date-time
```

---

## 4. Frontend Architecture & Routing

### Store & Tipos (`web/src/lib/types/goal.ts` y `goalsStore.svelte.ts`)
1. `EmployeeAssignment`:
   ```typescript
   export interface EmployeeAssignment {
       id: string;
       employeeId: string;
       employeeName: string;
       employeeNumber?: string;
       profileId: EvaluationProfile;
       managerId: string | null;
       goalIds: string[];
       status?: 'borrador' | 'enviada';
       submittedAt?: string | null;
       createdAt: string;
       updatedAt: string;
   }
   ```
2. Eliminación del auto-POST en `_doLoad()`:
   - Se elimina la creación automática de asignaciones vacías en `_doLoad()` al consultar subordinados para evitar crear borradores no deseados.
3. Normalización en `normalizeApiData`:
   - Mapear `status` y `submitted_at` desde la respuesta de la API.

### Vista de Asignación (`web/src/routes/objetivos/asignacion/+page.svelte`)
- Importa `page` desde `$app/stores` (o `$app/state`).
- Efecto reactivo para leer `empId` del query parameter:
  ```typescript
  $effect(() => {
      const urlEmpId = $page.url.searchParams.get('empId');
      if (urlEmpId && urlEmpId !== selectedEmployeeId) {
          selectedEmployeeId = urlEmpId;
          loadForEmployee(urlEmpId);
      }
  });
  ```

### Tabla de Evaluados (`web/src/lib/components/evaluation/EmployeeEvaluationTable.svelte`)
- Alineación de encabezados:
  - Empleado | Perfil | Estado Metas | Autoevaluación | Evaluación | Estado Final | Acciones
- Badge de Estado de Metas:
  - `Enviada` (badge-success) si `status === 'enviada'`.
  - `Borrador` (badge-warning) si `status === 'borrador'` o sin enviar.
- Botón **Metas**:
  ```svelte
  {@const isDraftOrBeginning = currentPhase === 'inicio-anio' || rowStatus === 'borrador'}
  <a
      href={isDraftOrBeginning ? `/objetivos/asignacion?empId=${row.id}` : `/mis-evaluados/${row.id}/metas`}
      class="btn btn-outline btn-xs gap-1"
      title="Ver/Editar Metas"
  >
      <Target class="w-3.5 h-3.5" />
      Metas
  </a>
  ```

### Nueva Ruta (`web/src/routes/mis-evaluados/[id]/metas/+page.svelte`)
- Carga datos del colaborador (`/employees/{id}`) y sus metas asignadas (`loadForEmployee(id)`).
- Muestra el resumen de objetivos personales, globales y compartidos, avances y KPIs vinculados en modo lectura/revisión para el evaluador.
- Incluye botón de regreso a `/mis-evaluados`.
