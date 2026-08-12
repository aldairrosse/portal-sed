# Tasks: Objetivos Globales + Metas Compartidas

## Fase 1: Schema BD

### 1.1 Extender Goal schema
- [ ] Agregar campo `type` (enum: personal, global, shared)
- [ ] Agregar campo `goal_kind` (enum: qualitative, quantitative)
- [ ] Crear migración SQL
- [ ] Actualizar Ent schema

### 1.2 Crear GoalTemplate schema
- [ ] Definir campos: name, description, unit, direction, targetValue, goalKind, createdBy, isPublic
- [ ] Crear migración SQL
- [ ] Crear Ent schema

### 1.3 Crear GoalTemplateKpiLink schema
- [ ] Definir relación N:M template-kpi
- [ ] Crear migración SQL
- [ ] Crear Ent schema

### 1.4 Crear GlobalGoalAssignment schema
- [ ] Definir campos: goalId, employeeId, weight, targetValue, baselineValue
- [ ] Crear migración SQL
- [ ] Crear Ent schema

### 1.5 Crear GlobalGoalRule schema
- [ ] Definir campos: goalId, ruleType, departmentId, minDirectReports, defaultWeight
- [ ] Crear migración SQL
- [ ] Crear Ent schema

### 1.6 Crear SharedGoalGroup schema
- [ ] Definir campos: goalId, createdBy, name, description
- [ ] Crear migración SQL
- [ ] Crear Ent schema

### 1.7 Crear SharedGoalMember schema
- [ ] Definir campos: groupId, employeeId, weight, targetValue, baselineValue
- [ ] Crear migración SQL
- [ ] Crear Ent schema

## Fase 2: Backend API

### 2.1 Repositorio Global Goals
- [ ] Crear `global_repo.go` con CRUD
- [ ] Implementar asignación masiva
- [ ] Implementar ejecución de reglas
- [ ] Implementar queries con filtros

### 2.2 Servicio Global Goals
- [ ] Crear `global_service.go`
- [ ] Implementar validación de permisos RH
- [ ] Implementar validación de pesos
- [ ] Implementar lógica de reglas

### 2.3 Handler Global Goals
- [ ] Crear `global_handler.go`
- [ ] Implementar endpoints REST
- [ ] Implementar validación de entrada
- [ ] Implementar respuesta de error

### 2.4 Repositorio Shared Goals
- [ ] Crear `shared_repo.go` con CRUD
- [ ] Implementar gestión de miembros
- [ ] Implementar registro de avance
- [ ] Implementar queries con filtros

### 2.5 Servicio Shared Goals
- [ ] Crear `shared_service.go`
- [ ] Implementar validación de permisos de creador
- [ ] Implementar validación de pesos
- [ ] Implementar lógica de miembros

### 2.6 Handler Shared Goals
- [ ] Crear `shared_handler.go`
- [ ] Implementar endpoints REST
- [ ] Implementar validación de entrada
- [ ] Implementar respuesta de error

### 2.7 Repositorio Templates
- [ ] Crear operaciones CRUD para plantillas
- [ ] Implementar vinculación KPIs
- [ ] Implementar uso de plantilla

### 2.8 Integración con Goals existente
- [ ] Modificar handler de goals para soportar tipo
- [ ] Modificar servicio de goals para filtrar por tipo
- [ ] Actualizar respuestas de API con tipo y goal_kind

## Fase 3: Frontend

### 3.1 Configuración de menú
- [ ] Agregar ítem "Objetivos globales" en menuConfig.ts
- [ ] Agregar ítem "Metas compartidas" en menuConfig.ts
- [ ] Configurar permisos por perfil

### 3.2 Ruta Objetivos Globales
- [ ] Crear `/objetivos/globales/+page.svelte`
- [ ] Implementar acordeones cualitativos/cuantitativos
- [ ] Implementar formulario de creación
- [ ] Implementar panel de asignación
- [ ] Implementar gestión de plantillas

### 3.3 Ruta Metas Compartidas
- [ ] Crear `/objetivos/compartidas/+page.svelte`
- [ ] Implementar acordeones cualitativos/cuantitativos
- [ ] Implementar formulario de creación
- [ ] Implementar panel de miembros
- [ ] Implementar gestión de miembros existentes

### 3.4 Extender Asignación existente
- [ ] Modificar `/objetivos/asignacion` para mostrar metas globales
- [ ] Modificar `/objetivos/asignacion` para mostrar metas compartidas
- [ ] Implementar modo solo lectura para metas globales/compartidas
- [ ] Agregar badges "Global" y "Compartida"

### 3.5 Componentes reutilizables
- [ ] Crear componente `AccordionSection` para separar cualitativos/cuantitativos
- [ ] Crear componente `GoalTypeBadge` para mostrar tipo de meta
- [ ] Crear componente `WeightIndicatorGlobal` para pesos variables
- [ ] Crear componente `AssignmentPanel` para asignación masiva
- [ ] Crear componente `MemberPanel` para gestión de miembros

### 3.6 Store
- [ ] Extender `goalsStore.svelte.ts` para metas globales
- [ ] Extender `goalsStore.svelte.ts` para metas compartidas
- [ ] Implementar getters para filtrar por tipo
- [ ] Implementar mutaciones para CRUD

## Fase 4: Tests

### 4.1 Tests Backend
- [ ] Tests de integración para API global goals
- [ ] Tests de integración para API shared goals
- [ ] Tests de validación de permisos
- [ ] Tests de validación de pesos

### 4.2 Tests Frontend
- [ ] Tests de componente para acordeones
- [ ] Tests de componente para formularios
- [ ] Tests de store para metas globales/compartidas
- [ ] Tests de navegación

### 4.3 Tests E2E
- [ ] Test de flujo RH crear meta global
- [ ] Test de flujo jefe crear meta compartida
- [ ] Test de visualización solo lectura
- [ ] Test de validación de pesos

## Fase 5: Documentación

### 5.1 OpenAPI
- [ ] Documentar endpoints de global goals
- [ ] Documentar endpoints de shared goals
- [ ] Documentar endpoints de templates

### 5.2 AGENTS.md
- [ ] Actualizar módulos de producto
- [ ] Documentar nuevas rutas
- [ ] Documentar permisos

## Orden de implementación sugerido

1. Schema BD (1.1-1.7)
2. Repositorios Backend (2.1, 2.4, 2.7)
3. Servicios Backend (2.2, 2.5)
4. Handlers Backend (2.3, 2.6)
5. Integración Backend (2.8)
6. Configuración menú (3.1)
7. Componentes reutilizables (3.5)
8. Store (3.6)
9. Ruta Objetivos Globales (3.2)
10. Ruta Metas Compartidas (3.3)
11. Extender Asignación (3.4)
12. Tests Backend (4.1)
13. Tests Frontend (4.2)
14. Tests E2E (4.3)
15. Documentación (5.1, 5.2)
