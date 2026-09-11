## 1. Spec y contrato

- [ ] 1.1 Publicar deltas (evaluations, ninebox-computed, competency-framework) y verificar `rtk proxy openspec validate --all --strict --json` en 0 failed.
- [ ] 1.2 Actualizar OpenAPI (`PUT goal-state`, snapshots, comentarios con autor/fecha) y verificar tipos TS generados sin diff pendiente.

## 2. Implementación futura (fuera de este artefacto, solo checklist)

- [ ] 2.1 Migración `evaluation_goals` (+`avance_progress`, `cierre_progress` NULL) y verificar migración sube/baja en limpio.
- [ ] 2.2 Sincronización `current_value`/snapshot por fase y cierre con valor directo (sin `final_rating`) y verificar tests de tabla del servicio en verde.
- [ ] 2.3 Ponderado 0.8/0.2 ignorando ausentes + escalado `direction` en 9-box y verificar matriz dorada coincide entre `RecomputeMatrix` y `ComputeMatrixView`.
