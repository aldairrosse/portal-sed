## ADDED Requirements

### Requirement: Comentarios de objetivo faseados

El sistema SHALL almacenar `goal_comments.phase` como `ENUM(asignacion,avance,cierre) NOT NULL DEFAULT cierre`, con índice en `(goal_id,phase)` y backfill documentado (filas existentes → `cierre`).

#### Scenario: Crear comentario en avance

- **WHEN** un evaluador llama `POST /goals/{id}/comments` con `phase=avance`
- **THEN** recibe 201 y el comentario persiste con `phase=avance`, autor y fecha

#### Scenario: Listar comentarios por fase

- **WHEN** se llama `GET /goals/{id}/comments?phase=cierre`
- **THEN** recibe 200 solo con comentarios de `cierre`, cada uno con autor y fecha

#### Scenario: Phase inválida

- **WHEN** se envía `phase=medio-anio` al backend
- **THEN** recibe 400 (el backend solo acepta el enum canónico)
