-- +goose Up
-- +goose StatementBegin
-- goose no crea phase_id (solo el auto-migrate de Ent en main.go lo hace,
-- DESPUÉS de goose). En BD limpias la tabla está vacía aquí, así que
-- NOT NULL sin default es seguro; en BD existentes la columna ya existe
-- y este ALTER es no-op.
ALTER TABLE nine_box_matrixes ADD COLUMN IF NOT EXISTS phase_id uuid NOT NULL;

-- Dedupe nine_box_matrixes: keep the row with the highest id per
-- (cycle_id, evaluator_id, phase_id) and re-assign nine_box_entries
-- to the survivor BEFORE deleting the duplicates (FK ON DELETE CASCADE).
-- phase_id is NOT NULL in this table.
-- PostgreSQL has no max(uuid), so the survivor (keep_id = highest id) is
-- picked with row_number() OVER (ORDER BY id DESC) instead of an aggregate.
-- CTEs are statement-scoped in PostgreSQL, so ranked/dups are repeated in
-- every statement that needs them.

-- 1) For each evaluatee, keep a single entry per duplicated group and never
--    lose the survivor's data. An entry on a LOSER matrix m is deleted when
--    the same evaluatee already exists on a "better" matrix of the same
--    group: the survivor (keep_id, highest id) or any loser with a SMALLER id
--    (deterministic order). With 3+ duplicates where the survivor lacks an
--    evaluatee shared by several losers, only the oldest loser's entry
--    survives, so the UPDATE below cannot collide on
--    idx_nine_box_entries_matrix_eval (matrix_id, evaluatee_id).
WITH ranked AS (
    SELECT id, cycle_id, evaluator_id, phase_id,
           row_number() OVER (
               PARTITION BY cycle_id, evaluator_id, phase_id
               ORDER BY id DESC
           ) AS rn,
           count(*) OVER (
               PARTITION BY cycle_id, evaluator_id, phase_id
           ) AS cnt
    FROM nine_box_matrixes
),
dups AS (
    SELECT cycle_id, evaluator_id, phase_id, id AS keep_id
    FROM ranked
    WHERE rn = 1 AND cnt > 1
)
DELETE FROM nine_box_entries e
USING nine_box_matrixes m, dups d
WHERE e.matrix_id = m.id
  AND m.id <> d.keep_id
  AND m.cycle_id = d.cycle_id
  AND m.evaluator_id = d.evaluator_id
  AND m.phase_id = d.phase_id
  AND EXISTS (
      SELECT 1
      FROM nine_box_entries s
      JOIN nine_box_matrixes m2 ON m2.id = s.matrix_id
      WHERE m2.cycle_id = d.cycle_id
        AND m2.evaluator_id = d.evaluator_id
        AND m2.phase_id = d.phase_id
        AND (m2.id = d.keep_id OR m2.id < m.id)
        AND s.evaluatee_id = e.evaluatee_id
  );

-- 2) Re-assign remaining entries of loser matrices to the survivor.
WITH ranked AS (
    SELECT id, cycle_id, evaluator_id, phase_id,
           row_number() OVER (
               PARTITION BY cycle_id, evaluator_id, phase_id
               ORDER BY id DESC
           ) AS rn,
           count(*) OVER (
               PARTITION BY cycle_id, evaluator_id, phase_id
           ) AS cnt
    FROM nine_box_matrixes
),
dups AS (
    SELECT cycle_id, evaluator_id, phase_id, id AS keep_id
    FROM ranked
    WHERE rn = 1 AND cnt > 1
)
UPDATE nine_box_entries e
SET matrix_id = d.keep_id
FROM nine_box_matrixes m, dups d
WHERE e.matrix_id = m.id
  AND e.matrix_id <> d.keep_id
  AND m.cycle_id = d.cycle_id
  AND m.evaluator_id = d.evaluator_id
  AND m.phase_id = d.phase_id;

-- 3) Delete loser matrices (their entries were reassigned or dropped above).
WITH ranked AS (
    SELECT id, cycle_id, evaluator_id, phase_id,
           row_number() OVER (
               PARTITION BY cycle_id, evaluator_id, phase_id
               ORDER BY id DESC
           ) AS rn,
           count(*) OVER (
               PARTITION BY cycle_id, evaluator_id, phase_id
           ) AS cnt
    FROM nine_box_matrixes
),
dups AS (
    SELECT cycle_id, evaluator_id, phase_id, id AS keep_id
    FROM ranked
    WHERE rn = 1 AND cnt > 1
)
DELETE FROM nine_box_matrixes m
USING dups d
WHERE m.id <> d.keep_id
  AND m.cycle_id = d.cycle_id
  AND m.evaluator_id = d.evaluator_id
  AND m.phase_id = d.phase_id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_nine_box_matrixes_cycle_eval_phase
    ON nine_box_matrixes (cycle_id, evaluator_id, phase_id);
-- +goose StatementEnd
