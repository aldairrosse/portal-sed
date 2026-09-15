# Datos y ORM

## Motor

PostgreSQL 15+.

## ORM

- **Preferido: Ent** — esquema como código, migraciones, tipos fuertes, buen fit con Go a escala.
- **Alternativa: GORM** — velocidad inicial; vigilar N+1 y migraciones.

## Índices (diseñar en spec antes de código)

Ejemplos de claves de acceso:

- Evaluaciones por `(organization_id, cycle_id, employee_id)`.
- Objetivos por `(evaluation_id, type)`.
- Catálogo por `(organization_id, category_id)`.

Toda query de listado en spec debe nombrar el índice esperado.

## Migraciones

- Herramienta: Atlas con Ent, o `golang-migrate` si GORM.
- Nunca editar migración ya aplicada en prod.

## Integridad

- FKs donde el dominio lo exija.
- Soft delete solo si spec de auditoría lo requiere.
- Campos JSONB para metadatos flexibles, no para datos query-heavy sin índice GIN planificado.

## Auditoría

Tabla o columnas: `created_at`, `updated_at`, `created_by`, `updated_by` en entidades de evaluación y objetivos.
