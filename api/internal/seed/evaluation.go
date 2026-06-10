package seed

import (
	"context"
	"log"

	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/evaluation"
)

// SeedEvaluation creates Evaluation records, EvaluationCompetency (self + RH ratings),
// and EvaluationGoal (closures).
func SeedEvaluation(ctx context.Context, client *internal.Client) error {
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

	cycleID := seedID("cycle-2026")

	// 1. Evaluations: one per employee that appears in self-evaluations or rh-evaluations or goal-closures
	evalEmployeeIDs := []string{
		"emp-colaborador-01",
		"emp-vendedor-01",
		"emp-jefe-01",
	}

	for _, empID := range evalEmployeeIDs {
		evalID := seedID("eval-" + empID)

		if err := tx.Evaluation.Create().
			SetID(evalID).
			SetPhase(evaluation.PhaseAvance).
			SetState(evaluation.StatePendienteAvance).
			SetEmployeeID(seedID(empID)).
			SetCycleID(cycleID).
			SetCreatedBy(seedID("system")).
			SetUpdatedBy(seedID("system")).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 2. EvaluationCompetency - self-ratings (from self-evaluations.json)
	// Fixture data maps employeeId + competencyId + selfRating/selfComment
	selfRatings := []struct {
		evalID       string
		employeeID   string
		competencyID string
		profileID    string
		rating       int
		comments     string
	}{
		{evalID: "eval-emp-colaborador-01", employeeID: "emp-colaborador-01", competencyID: "comp-comunicacion", profileID: "colaborador", rating: 4, comments: "Considero que mi comunicación ha sido clara y oportuna"},
		{evalID: "eval-emp-colaborador-01", employeeID: "emp-colaborador-01", competencyID: "comp-desarrollo-equipo", profileID: "colaborador", rating: 3, comments: "He participado en actividades de equipo"},
		{evalID: "eval-emp-colaborador-01", employeeID: "emp-colaborador-01", competencyID: "comp-toma-decisiones", profileID: "colaborador", rating: 4, comments: ""},
		{evalID: "eval-emp-colaborador-01", employeeID: "emp-colaborador-01", competencyID: "comp-dominio-herramientas", profileID: "colaborador", rating: 4, comments: "Domino las herramientas del puesto"},
		{evalID: "eval-emp-colaborador-01", employeeID: "emp-colaborador-01", competencyID: "comp-resolucion-problemas", profileID: "colaborador", rating: 3, comments: ""},
		{evalID: "eval-emp-colaborador-01", employeeID: "emp-colaborador-01", competencyID: "comp-colaboracion", profileID: "colaborador", rating: 4, comments: "Buena colaboración con el equipo"},
		{evalID: "eval-emp-colaborador-01", employeeID: "emp-colaborador-01", competencyID: "comp-adaptabilidad", profileID: "colaborador", rating: 0, comments: ""},
		{evalID: "eval-emp-colaborador-01", employeeID: "emp-colaborador-01", competencyID: "comp-responsabilidad", profileID: "colaborador", rating: 0, comments: ""},
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-comunicacion", profileID: "vendedor", rating: 5, comments: "Excelente comunicación con clientes"},
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-desarrollo-equipo", profileID: "vendedor", rating: 4, comments: ""},
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-toma-decisiones", profileID: "vendedor", rating: 4, comments: "Decisiones rápidas y efectivas"},
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-dominio-herramientas", profileID: "vendedor", rating: 4, comments: ""},
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-resolucion-problemas", profileID: "vendedor", rating: 5, comments: "Resolución efectiva de problemas"},
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-colaboracion", profileID: "vendedor", rating: 4, comments: ""},
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-adaptabilidad", profileID: "vendedor", rating: 3, comments: ""},
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-responsabilidad", profileID: "vendedor", rating: 5, comments: "Siempre cumplo con mis compromisos"},
		{evalID: "eval-emp-jefe-01", employeeID: "emp-jefe-01", competencyID: "comp-comunicacion", profileID: "jefe", rating: 0, comments: ""},
	}
	for _, sr := range selfRatings {
		ecID := seedID("ec-" + sr.evalID + "-" + sr.competencyID)

		b := tx.EvaluationCompetency.Create().
			SetID(ecID).
			SetEvaluationID(seedID(sr.evalID)).
			SetCompetencyID(seedID(sr.competencyID)).
			SetProfileID(seedID("profile-" + sr.profileID))

		if sr.rating > 0 {
			b = b.SetRating(sr.rating)
		} else {
			b = b.SetRating(1) // default minimum rating when fixture has no rating
		}
		if sr.comments != "" {
			b = b.SetComments(sr.comments)
		}

		if err := b.Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 3. RH ratings (from rh-evaluations.json)
	rhRatings := []struct {
		evalID       string
		employeeID   string
		competencyID string
		profileID    string
		rating       int
		comments     string
	}{
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-comunicacion", profileID: "vendedor", rating: 4, comments: "Comunicación efectiva con clientes, aunque puede mejorar en reportes escritos"},
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-desarrollo-equipo", profileID: "vendedor", rating: 3, comments: ""},
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-toma-decisiones", profileID: "vendedor", rating: 4, comments: ""},
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-dominio-herramientas", profileID: "vendedor", rating: 3, comments: "Requiere capacitación en CRM"},
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-resolucion-problemas", profileID: "vendedor", rating: 4, comments: ""},
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-colaboracion", profileID: "vendedor", rating: 4, comments: ""},
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-adaptabilidad", profileID: "vendedor", rating: 3, comments: ""},
		{evalID: "eval-emp-vendedor-01", employeeID: "emp-vendedor-01", competencyID: "comp-responsabilidad", profileID: "vendedor", rating: 5, comments: ""},
	}
	for _, rr := range rhRatings {
		ecID := seedID("ec-rh-" + rr.employeeID + "-" + rr.competencyID)

		b := tx.EvaluationCompetency.Create().
			SetID(ecID).
			SetEvaluationID(seedID(rr.evalID)).
			SetCompetencyID(seedID(rr.competencyID)).
			SetProfileID(seedID("profile-" + rr.profileID)).
			SetRating(rr.rating)

		if rr.comments != "" {
			b = b.SetComments(rr.comments)
		}

		if err := b.Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 4. EvaluationGoal (goal closures from goal-closures.json)
	goalClosures := []struct {
		employeeID    string
		goalID        string
		finalProgress int
		selfAssessment string
	}{
		{employeeID: "emp-colaborador-01", goalID: "goal-mejorar-satisfaccion", finalProgress: 85, selfAssessment: "Se logró mejorar la satisfacción del cliente significativamente"},
		{employeeID: "emp-colaborador-01", goalID: "goal-reducir-quejas", finalProgress: 70, selfAssessment: "Se redujeron las quejas aunque no se alcanzó la meta completa"},
		{employeeID: "emp-vendedor-01", goalID: "goal-alcanzar-ventas", finalProgress: 95, selfAssessment: "Superé la meta de ventas en el último trimestre"},
		{employeeID: "emp-jefe-01", goalID: "goal-alcanzar-ventas", finalProgress: 0, selfAssessment: ""},
		{employeeID: "emp-jefe-01", goalID: "goal-mejorar-margen", finalProgress: 0, selfAssessment: ""},
		{employeeID: "emp-jefe-01", goalID: "goal-reducir-costos", finalProgress: 0, selfAssessment: ""},
	}
	for _, gc := range goalClosures {
		evalID := seedID("eval-" + gc.employeeID)
		egID := seedID("eg-" + gc.employeeID + "-" + gc.goalID)

		// Convert 0-100 finalProgress to 1-5 rating
		var rating *int
		if gc.finalProgress > 0 {
			r := progressToRating(gc.finalProgress)
			rating = &r
		}

		b := tx.EvaluationGoal.Create().
			SetID(egID).
			SetEvaluationID(evalID).
			SetGoalID(seedID(gc.goalID))

		if rating != nil {
			b = b.SetFinalRating(*rating)
		}
		if gc.selfAssessment != "" {
			b = b.SetFinalComments(gc.selfAssessment)
		}

		if err := b.Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("[seed] evaluation: created %d evaluations, %d competency ratings, %d goal closures",
		len(evalEmployeeIDs), len(selfRatings)+len(rhRatings), len(goalClosures))
	return nil
}

// progressToRating converts a 0-100 progress value to a 1-5 rating scale.
func progressToRating(progress int) int {
	switch {
	case progress >= 81:
		return 5
	case progress >= 61:
		return 4
	case progress >= 41:
		return 3
	case progress >= 21:
		return 2
	default:
		return 1
	}
}
