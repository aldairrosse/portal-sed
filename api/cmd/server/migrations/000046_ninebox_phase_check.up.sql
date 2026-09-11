-- +goose Up
-- +goose StatementBegin
-- sed-cierre-avances-9box: NineBoxPhase = [avance, cierre] only.
-- 1) Idempotent backfill medio-anio -> avance (rerun of 000043; no-op when clean).
UPDATE phase_definitions SET phase = 'avance', updated_at = NOW() WHERE phase = 'medio-anio';

-- 2) Merge duplicate 'avance' phase_definitions per cycle (legacy cycles had
-- avance + medio-anio rows, now two 'avance' rows). Keeper = earliest created.
-- 2a) Move entries from loser-phase matrices into the keeper-phase matrix when
-- both exist for the same (cycle, evaluator); drop duplicated evaluatees first
-- so idx_nine_box_entries_matrix_eval (matrix_id, evaluatee_id) cannot collide.
DELETE FROM nine_box_entries e
USING
      (SELECT cycle_id, MIN(created_at) AS keep_created FROM phase_definitions
        WHERE phase = 'avance' GROUP BY cycle_id HAVING COUNT(*) > 1) dup
JOIN phase_definitions pd_loser ON pd_loser.cycle_id = dup.cycle_id AND pd_loser.phase = 'avance'
  AND pd_loser.created_at > dup.keep_created
JOIN nine_box_matrixes m_loser ON m_loser.phase_id = pd_loser.id
JOIN nine_box_matrixes m_keep ON m_keep.cycle_id = m_loser.cycle_id
  AND m_keep.evaluator_id = m_loser.evaluator_id
JOIN phase_definitions pd_keep ON pd_keep.id = m_keep.phase_id
  AND pd_keep.cycle_id = dup.cycle_id AND pd_keep.phase = 'avance'
  AND pd_keep.created_at = dup.keep_created
WHERE e.matrix_id = m_loser.id
  AND EXISTS (SELECT 1 FROM nine_box_entries s
              WHERE s.matrix_id = m_keep.id AND s.evaluatee_id = e.evaluatee_id);

UPDATE nine_box_entries e
SET matrix_id = m_keep.id
FROM
      (SELECT cycle_id, MIN(created_at) AS keep_created FROM phase_definitions
        WHERE phase = 'avance' GROUP BY cycle_id HAVING COUNT(*) > 1) dup
JOIN phase_definitions pd_loser ON pd_loser.cycle_id = dup.cycle_id AND pd_loser.phase = 'avance'
  AND pd_loser.created_at > dup.keep_created
JOIN nine_box_matrixes m_loser ON m_loser.phase_id = pd_loser.id
JOIN nine_box_matrixes m_keep ON m_keep.cycle_id = m_loser.cycle_id
  AND m_keep.evaluator_id = m_loser.evaluator_id
JOIN phase_definitions pd_keep ON pd_keep.id = m_keep.phase_id
  AND pd_keep.cycle_id = dup.cycle_id AND pd_keep.phase = 'avance'
  AND pd_keep.created_at = dup.keep_created
WHERE e.matrix_id = m_loser.id;

DELETE FROM nine_box_matrixes m_loser
USING
      (SELECT cycle_id, MIN(created_at) AS keep_created FROM phase_definitions
        WHERE phase = 'avance' GROUP BY cycle_id HAVING COUNT(*) > 1) dup
JOIN phase_definitions pd_loser ON pd_loser.cycle_id = dup.cycle_id AND pd_loser.phase = 'avance'
  AND pd_loser.created_at > dup.keep_created
JOIN phase_definitions pd_keep ON pd_keep.cycle_id = dup.cycle_id AND pd_keep.phase = 'avance'
  AND pd_keep.created_at = dup.keep_created
JOIN nine_box_matrixes m_keep ON m_keep.phase_id = pd_keep.id
WHERE m_loser.phase_id = pd_loser.id
  AND m_loser.cycle_id = m_keep.cycle_id
  AND m_loser.evaluator_id = m_keep.evaluator_id;

-- 2b) Point remaining loser-phase matrices at the keeper phase_id, then drop
-- the loser phase_definitions rows.
WITH keepers AS (
    SELECT DISTINCT ON (cycle_id) id AS keep_id, cycle_id, created_at
    FROM phase_definitions WHERE phase = 'avance' ORDER BY cycle_id, created_at, id
)
UPDATE nine_box_matrixes m
SET phase_id = k.keep_id
FROM phase_definitions pd, keepers k
WHERE m.phase_id = pd.id AND pd.cycle_id = k.cycle_id
  AND pd.phase = 'avance' AND pd.id <> k.keep_id;

DELETE FROM phase_definitions pd
USING (SELECT cycle_id, MIN(created_at) AS keep_created FROM phase_definitions
        WHERE phase = 'avance' GROUP BY cycle_id HAVING COUNT(*) > 1) dup
WHERE pd.cycle_id = dup.cycle_id AND pd.phase = 'avance'
  AND pd.created_at > dup.keep_created
  AND NOT EXISTS (SELECT 1 FROM nine_box_matrixes m WHERE m.phase_id = pd.id);

-- 3) DELETE matrices in asignacion via join phase_definitions (entries first,
-- matrices cascade via Ent OnDelete but explicit delete keeps raw-SQL DBs safe).
DELETE FROM nine_box_entries e
USING nine_box_matrixes m JOIN phase_definitions pd ON pd.id = m.phase_id
WHERE e.matrix_id = m.id AND pd.phase = 'asignacion';

DELETE FROM nine_box_matrixes m
USING phase_definitions pd
WHERE m.phase_id = pd.id AND pd.phase = 'asignacion';

-- 4) Complement unique index 000036 (re-ensure after phase_id remap).
CREATE UNIQUE INDEX IF NOT EXISTS idx_nine_box_matrixes_cycle_eval_phase
    ON nine_box_matrixes (cycle_id, evaluator_id, phase_id);

-- NOTE: a cross-table CHECK (nine_box_matrixes.phase_id -> phase_definitions.phase
-- IN ('avance','cierre')) cannot be a plain CHECK constraint over an FK; it would
-- require a trigger. Enforcement lives in NineBoxService.ResolvePhaseID
-- (application-level rejection of non-avance/cierre phases).
-- +goose StatementEnd
