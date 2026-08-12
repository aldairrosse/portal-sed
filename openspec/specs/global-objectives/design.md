# Design: Objetivos Globales + Metas Compartidas

## Arquitectura general

```
┌─────────────────────────────────────────────────────────────┐
│  Frontend (Svelte 5)                                        │
│                                                              │
│  /objetivos/globales     /objetivos/compartidas              │
│  ┌─────────────────┐    ┌─────────────────────┐             │
│  │ Acordeones      │    │ Acordeones           │             │
│  │ Cualitativos    │    │ Cualitativos         │             │
│  │ Cuantitativos   │    │ Cuantitativos        │             │
│  └─────────────────┘    └─────────────────────┘             │
│                                                              │
│  /objetivos/asignacion (existente, extendido)                │
│  ┌─────────────────────────────────────────────┐             │
│  │ Metas globales (solo lectura)               │             │
│  │ Metas compartidas (solo lectura)            │             │
│  │ Metas personales (editable)                 │             │
│  └─────────────────────────────────────────────┘             │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  API (Go + Chi)                                             │
│                                                              │
│  /api/v1/goals/global      /api/v1/goals/shared              │
│  ┌─────────────────┐      ┌─────────────────────┐           │
│  │ CRUD + assign   │      │ CRUD + members       │           │
│  │ rules + template│      │ progress             │           │
│  └─────────────────┘      └─────────────────────┘           │
│                                                              │
│  /api/v1/goals (existente, extendido)                        │
│  ┌─────────────────────────────────────────────┐             │
│  │ CRUD personal + tipo + goal_kind            │             │
│  └─────────────────────────────────────────────┘             │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  BD (PostgreSQL)                                            │
│                                                              │
│  goals (extendido: type, goal_kind)                          │
│  goal_categories (extendido: type)                           │
│  goal_templates (nuevo)                                      │
│  goal_template_kpi_links (nuevo)                             │
│  global_goal_assignments (nuevo)                             │
│  global_goal_rules (nuevo)                                   │
│  shared_goal_groups (nuevo)                                  │
│  shared_goal_members (nuevo)                                 │
└─────────────────────────────────────────────────────────────┘
```

## Schema BD propuesto

### Goal (extendido)

```sql
-- Nuevos campos
ALTER TABLE goals ADD COLUMN type goal_type DEFAULT 'personal';
ALTER TABLE goals ADD COLUMN goal_kind goal_kind_type;

-- Tipos
CREATE TYPE goal_type AS ENUM ('personal', 'global', 'shared');
CREATE TYPE goal_kind_type AS ENUM ('qualitative', 'quantitative');
```

### GoalTemplate (nuevo)

```sql
CREATE TABLE goal_templates (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR NOT NULL,
  description TEXT,
  unit goal_unit NOT NULL,
  direction direction_type DEFAULT 'ascendente',
  target_value FLOAT NOT NULL,
  goal_kind goal_kind_type NOT NULL,
  created_by UUID NOT NULL REFERENCES employees(id),
  is_public BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### GoalTemplateKpiLink (nuevo)

```sql
CREATE TABLE goal_template_kpi_links (
  template_id UUID NOT NULL REFERENCES goal_templates(id) ON DELETE CASCADE,
  kpi_id UUID NOT NULL REFERENCES kpis(id) ON DELETE CASCADE,
  PRIMARY KEY (template_id, kpi_id)
);
```

### GlobalGoalAssignment (nuevo)

```sql
CREATE TABLE global_goal_assignments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
  employee_id UUID NOT NULL REFERENCES employees(id),
  weight FLOAT NOT NULL CHECK (weight >= 0 AND weight <= 100),
  target_value FLOAT NOT NULL,
  baseline_value FLOAT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(goal_id, employee_id)
);
```

### GlobalGoalRule (nuevo)

```sql
CREATE TABLE global_goal_rules (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
  rule_type rule_type NOT NULL,
  department_id UUID REFERENCES org_nodes(id),
  min_direct_reports INT,
  default_weight FLOAT NOT NULL CHECK (default_weight >= 0 AND default_weight <= 100),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(goal_id, rule_type)
);

CREATE TYPE rule_type AS ENUM ('department', 'min_direct_reports');
```

### SharedGoalGroup (nuevo)

```sql
CREATE TABLE shared_goal_groups (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
  created_by UUID NOT NULL REFERENCES employees(id),
  name VARCHAR NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### SharedGoalMember (nuevo)

```sql
CREATE TABLE shared_goal_members (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id UUID NOT NULL REFERENCES shared_goal_groups(id) ON DELETE CASCADE,
  employee_id UUID NOT NULL REFERENCES employees(id),
  weight FLOAT NOT NULL CHECK (weight >= 0 AND weight <= 100),
  target_value FLOAT NOT NULL,
  baseline_value FLOAT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(group_id, employee_id)
);
```

## API Endpoints

### Global Objectives

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/goals/global` | Crear meta global con asignaciones/reglas |
| GET | `/api/v1/goals/global` | Listar metas globales por ciclo |
| PUT | `/api/v1/goals/global/:id` | Actualizar meta global |
| DELETE | `/api/v1/goals/global/:id` | Eliminar meta global |
| POST | `/api/v1/goals/global/:id/execute-rules` | Ejecutar reglas de asignación |
| GET | `/api/v1/goals/templates` | Listar plantillas |
| POST | `/api/v1/goals/templates` | Crear plantilla |
| POST | `/api/v1/goals/templates/:id/use` | Crear meta desde plantilla |

### Shared Goals

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/goals/shared` | Crear meta compartida con miembros |
| GET | `/api/v1/goals/shared` | Listar metas compartidas (creador o miembro) |
| PUT | `/api/v1/goals/shared/:id` | Actualizar meta compartida |
| DELETE | `/api/v1/goals/shared/:id` | Eliminar meta compartida |
| POST | `/api/v1/goals/shared/:id/members` | Agregar miembro |
| DELETE | `/api/v1/goals/shared/:id/members/:employeeId` | Remover miembro |
| PUT | `/api/v1/goals/shared/:id/progress/:employeeId` | Registrar avance |

## Permisos

| Acción | RH | Jefe/Director | Empleado |
|--------|----|---------------|----------|
| Crear meta global | ✅ | ❌ | ❌ |
| Editar meta global | ✅ | ❌ | ❌ |
| Ver meta global | ✅ | ✅ | ✅ (solo lectura) |
| Crear meta compartida | ❌ | ✅ | ❌ |
| Editar meta compartida | ❌ | ✅ (solo creador) | ❌ |
| Ver meta compartida | ✅ | ✅ | ✅ (solo lectura) |
| Registrar avance global | ✅ | ❌ | ❌ |
| Registrar avance compartida | ❌ | ✅ (solo creador) | ❌ |

## Validación de pesos

```
Total empleado = 
  Sum(metas personales) + 
  Sum(metas globales) + 
  Sum(metas compartidas) 
  ≤ 100%

Regla: las metas globales y compartidas son obligatorias,
las personales complementan.
```
