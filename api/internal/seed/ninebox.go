package seed

import (
	"context"
	"log"

	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/nineboxscale"
)

// SeedNineBox creates NineBoxScales, NineBoxQuadrants, NineBoxMatrices, and NineBoxEntries.
func SeedNineBox(ctx context.Context, client *internal.Client) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// 1. NineBoxScales (9 levels × 2 axes = 18 entries)
	scaleLabels := []struct {
		level       int
		label       string
		description string
	}{
		{level: 1, label: "Muy bajo", description: "No cumple expectativas"},
		{level: 2, label: "Bajo", description: "Requiere mejora significativa"},
		{level: 3, label: "Por debajo", description: "Debajo del estándar esperado"},
		{level: 4, label: "Aceptable", description: "Cumple mínimamente"},
		{level: 5, label: "Moderado", description: "Cumple parcialmente"},
		{level: 6, label: "Bueno", description: "Cumple consistentemente"},
		{level: 7, label: "Notable", description: "Supera expectativas"},
		{level: 8, label: "Sobresaliente", description: "Supera consistentemente"},
		{level: 9, label: "Excepcional", description: "Referente en la organización"},
	}
	for _, sl := range scaleLabels {
		for _, axis := range []string{"performance", "potential"} {
			sid := seedID("nbs-" + axis + "-" + string(rune('0'+sl.level)))
			// Use a simple sequential approach for level IDs
			// Generate unique IDs using axis + level
			if err := tx.NineBoxScale.Create().
				SetID(sid).
				SetAxis(nineboxscale.Axis(axis)).
				SetLevel(sl.level).
				SetLabel(sl.label).
				SetDescription(sl.description).
				Exec(ctx); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}

	// 2. NineBoxQuadrants (7 quadrants from quadrant-definitions.json)
	quadrants := []struct {
		id                   string
		quadrant             int
		label                string
		description          string
		color                string
		actionRecommendation string
	}{
		{id: "star", quadrant: 1, label: "Estrella", description: "Alto desempeño, alto potencial", color: "bg-success/20", actionRecommendation: "Retener y desarrollar como futuro líder"},
		{id: "growth", quadrant: 2, label: "Crecimiento", description: "Alto desempeño, potencial medio", color: "bg-info/20", actionRecommendation: "Impulsar desarrollo para alcanzar máximo potencial"},
		{id: "high-potential", quadrant: 3, label: "Alto Potencial", description: "Desempeño medio, alto potencial", color: "bg-accent/20", actionRecommendation: "Asignar mentor y retos progresivos"},
		{id: "core-player", quadrant: 4, label: "Jugador Clave", description: "Desempeño medio, potencial medio", color: "bg-primary/10", actionRecommendation: "Mantener motivación y ofrecer desarrollo constante"},
		{id: "risk", quadrant: 5, label: "Riesgo", description: "Bajo desempeño, alto potencial", color: "bg-warning/20", actionRecommendation: "Investigar causas y crear plan de mejora"},
		{id: "effective", quadrant: 6, label: "Efectivo", description: "Bajo desempeño, potencial medio / Desempeño medio, bajo potencial", color: "bg-secondary/20", actionRecommendation: "Reforzar áreas de oportunidad con capacitación"},
		{id: "underperformer", quadrant: 7, label: "Bajo Rendimiento", description: "Bajo desempeño, bajo potencial", color: "bg-error/20", actionRecommendation: "Evaluar plan de mejora o separación"},
	}
	for _, q := range quadrants {
		qid := seedID("nbq-" + q.id)
		if err := tx.NineBoxQuadrant.Create().
			SetID(qid).
			SetQuadrant(q.quadrant).
			SetLabel(q.label).
			SetDescription(q.description).
			SetColor(q.color).
			SetActionRecommendation(q.actionRecommendation).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 3. NineBoxMatrices (one per evaluator that has entries)
	// From matrix-entries.json, the entries reference evaluators.
	// We create one matrix per evaluator for the cycle.
	evaluators := []string{"emp-dg-01"}
	for _, evaluatorID := range evaluators {
		matrixID := seedID("nbm-cycle-2026-" + evaluatorID)

		if err := tx.NineBoxMatrix.Create().
			SetID(matrixID).
			SetCycleID(seedID("cycle-2026")).
			SetEvaluatorID(seedID(evaluatorID)).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 4. NineBoxEntries (from matrix-entries.json)
	type entryDef struct {
		id               string
		evaluateeID      string
		performanceScore int
		potentialScore   int
		quadrant         int
		comments         string
	}

	entries := []entryDef{
		{id: "nb-dg-01", evaluateeID: "emp-dg-01", performanceScore: 9, potentialScore: 8, quadrant: 1, comments: ""},
		{id: "nb-director-01", evaluateeID: "emp-director-01", performanceScore: 8, potentialScore: 7, quadrant: 1, comments: ""},
		{id: "nb-jefe-01", evaluateeID: "emp-jefe-01", performanceScore: 7, potentialScore: 6, quadrant: 2, comments: ""},
		{id: "nb-colaborador-01", evaluateeID: "emp-colaborador-01", performanceScore: 5, potentialScore: 6, quadrant: 4, comments: ""},
		{id: "nb-vendedor-01", evaluateeID: "emp-vendedor-01", performanceScore: 8, potentialScore: 5, quadrant: 2, comments: ""},
		{id: "nb-rh-01", evaluateeID: "emp-rh-01", performanceScore: 3, potentialScore: 7, quadrant: 5, comments: ""},
	}

	matrixID := seedID("nbm-cycle-2026-emp-dg-01")

	for _, e := range entries {
		eid := seedID(e.id)
		if err := tx.NineBoxEntry.Create().
			SetID(eid).
			SetPerformanceScore(e.performanceScore).
			SetPotentialScore(e.potentialScore).
			SetQuadrant(e.quadrant).
			SetComments(e.comments).
			SetMatrixID(matrixID).
			SetEvaluateeID(seedID(e.evaluateeID)).
			SetCreatedBy(seedID("system")).
			SetUpdatedBy(seedID("system")).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("[seed] ninebox: created 18 scale entries, %d quadrants, 1 matrix, %d entries",
		len(quadrants), len(entries))
	return nil
}
