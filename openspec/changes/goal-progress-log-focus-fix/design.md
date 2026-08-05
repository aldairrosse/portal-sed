# Design: goal-progress-log-focus-fix

## 1. Arquitectura general

```
web/                                          api/
GoalRow.svelte                                ┌─────────────────────────────────┐
├─ input (bind:value local)                   │ handler/goal/goal_handler.go    │
│  └─ onchange (blur) ──► onUpdateProgress    │  UpdateGoalProgress (PATCH)     │
│                          │                  └──────────────┬──────────────────┘
│                          ▼                                 ▼
│               goalsStore.updateGoalProgress     service/goal/progress_service.go
│               (optimista, sin await,            ├─ phaseCheck.CanUpdateProgress
│                sin reload)                      ├─ GetGoal → previousValue
│                          │                      └─ ownership check (cat.EmployeeID)
│                          ▼                                  │
│               PATCH /goals/{goalId}/progress    repository/goal/goal_repo.go
│               (background, fire & forget)       UpdateGoalCurrentValueWithLog
│               error → rollback + alerta         ├─ BeginTx
│                                                 ├─ UPDATE goals SET current_value,
│                                                 │   updated_at, version = version+1
│                                                 ├─ INSERT goal_progress_log
│                                                 │   (goal_id, employee_id,
│                                                 │    previous_value, new_value)
│                                                 └─ Commit / Rollback
```

Dos frentes independientes:

1. **Frontend** — fix de foco y guardado en background (no toca la API, no depende de C8 para el fix; el patrón queda listo para el wiring).
2. **Backend** — tabla `goal_progress_log` con escritura atómica junto al UPDATE de `current_value`.

## 2. Frontend: GoalRow.svelte

### Estado

```ts
// Antes (bug):
let progressValue = $state(goal.progress ?? 0);
$effect(() => { progressValue = goal.progress ?? 0; });   // ← resync por prop: churn + focus loss

// Después:
let progressValue = $state(goal.progress ?? 0);            // seed al montar
let progressFocused = $state(false);

function handleProgressFocus()  { progressFocused = true; }
function handleProgressBlur()   {
  progressFocused = false;
  const val = progressValue;
  if (!isNaN(val) && val !== goal.progress) onUpdateProgress?.(goal.id, val);
}
// Resync externo solo cuando NO hay foco (opcional, para cambio de empleado):
// $effect(() => { if (!progressFocused) progressValue = goal.progress ?? 0; });
```

### Markup

```svelte
<input
  type="number"
  class="input input-bordered input-xs w-20"
  bind:value={progressValue}
  onfocus={handleProgressFocus}
  onchange={handleProgressBlur}   <!-- onchange → blur; se elimina oninput -->
  ...
/>
```

Cambios clave:

- `oninput` (cada tecla) → `onchange` (blur). Una persistencia por sesión de edición.
- `value={progressValue}` + `oninput` → `bind:value={progressValue}` (estado local autoritativo).
- Eliminado el `$effect` incondicional que sobreescribía el valor y forzaba re-render con cada prop nuevo (el `{...g}` del store recreaba el objeto en cada tecla).
- El guardado compara con `goal.progress` para no persistir si no hubo cambio.

### goalsStore.svelte.ts — updateGoalProgress (optimista, background)

```ts
export async function updateGoalProgress(goalId: string, progress: number): Promise<void> {
  const prev = goals.find((g) => g.id === goalId)?.progress;
  // Optimista: aplicar inmediatamente sin tocar el nodo del input
  goals = goals.map((g) =>
    g.id === goalId ? { ...g, progress, progressUpdatedAt: new Date().toISOString() } : g
  );
  try {
    await persistProgress(goalId, progress);   // PATCH vía client (C8); hoy: no-op en DEV
  } catch {
    // Rollback local, sin refetch
    if (prev !== undefined) {
      goals = goals.map((g) => (g.id === goalId ? { ...g, progress: prev } : g));
    }
    lastProgressError = `No se pudo guardar el avance de "${goals.find((g) => g.id === goalId)?.name}".`;
  }
}
```

- **Nunca** `reload()` / `invalidateAll()` tras guardar progreso.
- `lastProgressError` expuesto para que la página muestre una alerta (`errorMsg` existente en `+page.svelte`).
- `persistProgress` es la frontera con el wiring de C8: hoy `no-op` en DEV (fixtures), en producción `client.PATCH('/goals/{goalId}/progress', ...)`.

## 3. Backend: goal_progress_log

### 3.1 Ent schema — `api/internal/schema/goalprogresslog.go`

```go
type GoalProgressLog struct{ ent.Schema }

func (GoalProgressLog) Mixin() []ent.Mixin { return []ent.Mixin{AuditMixin{}} } // id + created_at + created_by

func (GoalProgressLog) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).StorageKey("id"),
        field.UUID("goal_id", uuid.UUID{}),
        field.UUID("employee_id", uuid.UUID{}),
        field.Float("previous_value"),
        field.Float("new_value"),
    }
}

func (GoalProgressLog) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("goal", Goal.Type).Ref("progress_logs").Unique().Required().Field("goal_id"),
    }
}

func (GoalProgressLog) Index() []ent.Index {
    return []ent.Index{
        index.Fields("goal_id"),
        index.Fields("employee_id"),
    }
}
```

- `created_at`/`created_by` vienen del `AuditMixin` existente (igual que `Goal`).
- En `Goal.Edges()` se agrega `edge.To("progress_logs", GoalProgressLog.Type)`.

### 3.2 Migración — `api/internal/migrate/schema.go`

Nueva entrada siguiendo el patrón de `GoalsTable`:

```go
GoalProgressLogsTable = &schema.Table{
    Name:    "goal_progress_logs",
    Symbol:  "goal_progress_logs_goals_progress_logs",
    Columns: []*schema.Column{ /* id, goal_id, employee_id, previous_value, new_value, created_at, created_by */ },
    Indexes: []*schema.Index{
        {Name: "idx_goal_progress_logs_goal_id", Columns: [*]{GoalIDColumn}},
        {Name: "idx_goal_progress_logs_employee_id", Columns: [*]{EmployeeIDColumn}},
    },
    Relations: []*schema.Relation{ /* BelongsTo goals */ },
}
```

### 3.3 Repo — `api/internal/repository/goal/goal_repo.go`

```go
// UpdateGoalCurrentValueWithLog atomically updates current_value and inserts a progress log row.
func (r *GoalRepo) UpdateGoalCurrentValueWithLog(
    ctx context.Context, goalID uuid.UUID, employeeID uuid.UUID,
    currentValue, previousValue float64,
) (*GoalRow, error) {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil { return nil, err }
    defer tx.Rollback() // no-op tras Commit

    res, err := tx.ExecContext(ctx,
        `UPDATE goals SET current_value = $1, updated_at = $2, version = version + 1 WHERE id = $3`,
        currentValue, time.Now(), goalID)
    if err != nil { return nil, err }
    affected, _ := res.RowsAffected()
    if affected == 0 { return nil, pkgerrors.ErrGoalNotFound }

    _, err = tx.ExecContext(ctx,
        `INSERT INTO goal_progress_logs (goal_id, employee_id, previous_value, new_value, created_at)
         VALUES ($1, $2, $3, $4, $5)`,
        goalID, employeeID, previousValue, currentValue, time.Now())
    if err != nil { return nil, err }

    if err := tx.Commit(); err != nil { return nil, err }

    g, err := r.client.Goal.Query().Where(goal.ID(goalID)).Only(ctx) // mismo re-fetch que hoy
    if err != nil { return nil, err }
    return r.goalToRow(ctx, g)
}
```

### 3.4 Servicio — `api/internal/service/goal/progress_service.go`

```go
func (s *ProgressService) UpdateGoalProgress(ctx context.Context, empID, goalID uuid.UUID, req dtogoal.UpdateProgressRequest) (*repogoal.GoalRow, error) {
    if err := s.phaseCheck.CanUpdateProgress(ctx, empID.String()); err != nil { return nil, err }
    if req.CurrentValue < 0 { return nil, pkgerrors.ErrInvalidRequest }

    existing, err := s.goalRepo.GetGoal(ctx, goalID)   // ya existe: ownership + previousValue
    if err != nil { return nil, err }
    cat, err := s.catRepo.GetCategory(ctx, existing.CategoryID)
    if err != nil { return nil, err }
    if cat.EmployeeID != empID { return nil, pkgerrors.ErrGoalNotFound }

    if equalFloat(existing.CurrentValue, req.CurrentValue) { return existing, nil } // sin cambio → sin log

    return s.goalRepo.UpdateGoalCurrentValueWithLog(ctx, goalID, empID, req.CurrentValue, existing.CurrentValue)
}
```

- `equalFloat` usa la tolerancia EPSILON (0.01) que el store ya emplea.
- `ProgressServicer` (interfaces.go) no cambia de firma; solo la implementación.

### 3.5 OpenAPI — `api/openapi/goals-api.yaml`

`updateGoalProgress` (línea ~174): actualizar el `summary` y añadir una nota de auditoría:

```yaml
summary: Update goal progress (currentValue) — each change persists a goal_progress_log row (previous_value → new_value, recorded_by = session employee)
```

Sin cambios en `UpdateProgressRequest` ni en las respuestas.

## 4. Manejo de errores

| Caso | Frontend | Backend |
|---|---|---|
| Phase no permite progreso (`403`) | Rollback + alerta | `PhaseRestricted` (sin cambios) |
| Meta no existe / no es del empleado (`404`) | Rollback + alerta | `ErrGoalNotFound` (sin cambios) |
| Mismo valor | No persiste (guard en blur) | No log, no `version++` |
| Fallo de BD entre UPDATE e INSERT | — | Rollback de transacción → nada cambia |
| Red/5xx | Rollback local, sin refetch | — |

## 5. Verificación

- **Backend:** `go test ./api/...` — nuevo test en `progress_service_test.go`/`goal_handler_test.go`: PATCH cambia `current_value`, crea fila de log con `previous_value`/`new_value`/`employee_id`, no crea fila si el valor no cambió, no crea fila en fase que no sea avance.
- **Frontend:** `pnpm run check` (svelte-check) + test de componente GoalRow (Vitest + Testing Library): escribir no pierde foco; blur persiste; error de persistencia restaura valor.
- **Migración:** aplicar contra BD local (docker-compose) y verificar `goal_progress_logs` + índices.
