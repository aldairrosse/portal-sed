# Proposal: Objetivos Globales (RH)

## Intent

Permitir a RH crear metas organizacionales y asignarlas masivamente a empleados, departamentos o grupos con reglas de negocio flexibles. Las metas globales son de solo lectura para empleados y soportan ponderación variable por destinatario.

## Scope

### In scope
- Creación de metas globales por RH (cualitativas/cuantitativas)
- Asignación masiva: empleados específicos, departamento, reportes directos
- Pesos variables por destinatario
- Formularios reutilizables (plantillas)
- Acordeones cualitativos/cuantitativos
- API REST para CRUD
- UI en `/objetivos/globales`

### Out of scope
- Metas compartidas (scope de `shared-goals`)
- Evaluación de metas (scope de A5)
- Notificaciones
- Historial de cambios

## Approach

1. Extender schema `Goal` con campos `type` y `goal_kind`
2. Crear tablas `GoalTemplate`, `GlobalGoalAssignment`, `GlobalGoalRule`
3. Implementar API REST con permisos RH-only
4. Crear UI con acordeones y herramientas de asignación
5. Validar integridad de pesos

## Risks

- **Complejidad de asignación masiva**: reglas por departamento/reportes pueden ser complejas de implementar
- **Peso variable**: validar que la suma total por empleado ≤100% con metas personales
- **Rendimiento**: consultas con múltiples JOINs para asignaciones masivas
