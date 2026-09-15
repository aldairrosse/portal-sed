# Dominio — Evaluaciones y objetivos

## Conceptos

- **Ciclo de evaluación:** ventana temporal (fijación → evaluación → cierre).
- **Tipo de objetivo:** desempeño, puesto, otro según catálogo.
- **Objetivo:** indicador, descripción, categoría, metas, peso opcional.
- **Evaluación:** instancia por empleado + evaluador + ciclo + estado.

## Estados (borrador)

- Objetivo: borrador, fijado, en evaluación, cerrado.
- Evaluación empleado: pendiente fijación, pendiente evaluación, completada.

## Reglas a cerrar en OpenSpec

- Quién puede editar después de fijar.
- Notificación email: triggers, plantillas, reintentos.
- Metas numéricas vs cualitativas y validaciones.
- Agregados para RH (sin exponer datos de otros empleados al evaluado).

## Notificaciones

- Cola async; no bloquear request HTTP.
- Idempotencia por `(evaluation_id, event_type)`.

## Skill OpenSpec recomendada

`evaluation-lifecycle` — máquina de estados única documentada antes de handlers.
