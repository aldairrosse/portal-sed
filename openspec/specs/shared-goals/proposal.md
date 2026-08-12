# Proposal: Metas Compartidas (Jefes/Directores)

## Intent

Permitir a jefes/directores crear metas para su grupo directo, donde solo el creador puede editar y registrar avance. Los miembros ven las metas en modo solo lectura con ponderación variable.

## Scope

### In scope
- Creación de metas compartidas por jefes/directores
- Grupo de destinatarios (evaluados directos)
- Pesos variables por miembro
- Permisos exclusivos del creador
- Acordeones cualitativos/cuantitativos
- API REST para CRUD y gestión de miembros
- UI en `/objetivos/compartidas`

### Out of scope
- Metas globales (scope de `global-objectives`)
- Evaluación de metas (scope de A5)
- Notificaciones
- Historial de cambios
- Edición por miembros

## Approach

1. Extender schema `Goal` con campos `type` y `goal_kind`
2. Crear tablas `SharedGoalGroup`, `SharedGoalMember`
3. Implementar API REST con permisos de creador
4. Crear UI con acordeones y gestión de miembros
5. Validar integridad de pesos

## Risks

- **Permisos**: garantizar que solo el creador pueda editar
- **Peso variable**: validar que la suma total por empleado ≤100% con metas personales/globales
- **Gestión de miembros**: agregar/remover miembros después de crear la meta
