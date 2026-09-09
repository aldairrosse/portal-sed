# Delta para fix-competency-network-data

## employee-competency-ratings (NUEVO)

### Propósito

Endpoint y store para consultar calificaciones de competencias de un empleado específico por ciclo. Alimenta `CompetencyNetworkView` en la ruta `/evaluacion/9x9/competencias/{employeeId}`.

### Requerimiento: Consulta de competencias evaluadas por empleado

El sistema DEBE exponer `GET /api/v1/evaluations/employee/{employeeId}?cycle_id={cycleId}` y retornar las competencias con su calificación, pilar, nivel de aceptación y criterios de escala organizados en la estructura que `CompetencyNetworkView` espera.

#### Escenario: Consulta exitosa

- GIVEN empleado `E123` con 3 competencias evaluadas en el ciclo `C-2025`
- WHEN se llama `GET /evaluations/employee/E123?cycle_id=C-2025`
- THEN retorna HTTP 200 con cuerpo:
  - `employee: { id, name }`
  - `cycle: { id, phase }`
  - `pillars: [{ id, name, competencies: [{ id, name, rating, acceptanceLevel, scaleCriteria: [{ level, description }] }] }]`
- AND las competencias están agrupadas por su pilar (`pillarId` en `competency-framework`)

#### Escenario: Empleado sin evaluaciones en el ciclo

- GIVEN empleado `E999` sin registros en `evaluation_competencies` para ciclo `C-2025`
- WHEN se consulta `GET /evaluations/employee/E999?cycle_id=C-2025`
- THEN retorna HTTP 200 con `pillars: []`
- AND `CompetencyNetworkView` muestra estado vacío ("No hay evaluaciones registradas")

#### Escenario: Falta parámetro cycle_id

- GIVEN cualquier empleado
- WHEN se llama `GET /evaluations/employee/{employeeId}` sin `cycle_id`
- THEN retorna HTTP 400 con `{ code: "MISSING_PARAM", message: "cycle_id es requerido" }`

#### Escenario: Empleado no accesible por el requester

- GIVEN requester sin rol `rh` ni relación jerárquica con empleado `E123`
- WHEN consulta `GET /evaluations/employee/E123?cycle_id=C-2025`
- THEN retorna HTTP 403

### Requerimiento: Store evaluationStore.load() con employeeId opcional

El store `evaluationStore` DEBE aceptar un parámetro opcional `employeeId` en `load()`. Si se provee, consulta `GET /evaluations/employee/{employeeId}?cycle_id=...`. Si se omite, mantiene el comportamiento actual (datos del usuario logueado). El `cycle_id` se obtiene del store de ciclo activo.

#### Escenario: Carga por employeeId externo

- GIVEN store con `activeCycleId = "C-2025"` y `employeeId = "E123"` como argumento
- WHEN se invoca `evaluationStore.load("E123")`
- THEN consulta `GET /evaluations/employee/E123?cycle_id=C-2025`
- AND popula las mismas propiedades reactivas que usa `CompetencyNetworkView`

#### Escenario: Retrocompatibilidad — carga sin employeeId

- GIVEN store sin argumento `employeeId`
- WHEN se invoca `evaluationStore.load()`
- THEN mantiene la lógica actual (endpoint / lógica de usuario logueado)
- AND no rompe ninguna pantalla existente

---

## evaluation-lifecycle (MODIFICADO)

### Requerimiento: La ruta de competencias debe invocar carga de datos al montar

La ruta `/evaluacion/9x9/competencias/[employeeId]` DEBE disparar `evaluationStore.load(employeeId)` al montar el componente de página.

(Previamente: la ruta renderizaba `CompetencyNetworkView` sin invocar `evaluationStore.load()`, resultando en datos vacíos.)

#### Escenario: Carga de datos al montar la ruta

- GIVEN ruta `/evaluacion/9x9/competencias/E123`
- WHEN el componente `+page.svelte` se monta
- THEN `+page.ts` invoca `evaluationStore.load("E123")` en `onMount`
- AND `cycle_id` se obtiene del store de ciclo activo (`cycleStore.activeId`)
- AND `CompetencyNetworkView` recibe datos reactivos y renderiza la red de competencias

#### Escenario: Skeleton visible durante carga

- GIVEN ruta de competencias montada y `load()` en progreso
- WHEN los datos aún no se han resuelto
- THEN se renderiza `CompetencyNetworkSkeleton` (componente nuevo, patrón existente de skeleton)
- AND al resolverse la carga, el skeleton se reemplaza por `CompetencyNetworkView`

#### Escenario: Error en carga muestra estado de error

- GIVEN ruta de competencias montada
- WHEN `load("E123")` falla (red, 403, 500)
- THEN se muestra estado de error con mensaje contextual
- AND NO se muestra el skeleton indefinidamente
