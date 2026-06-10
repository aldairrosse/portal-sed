package seed

import (
	"context"
	"log"
	"time"

	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/cycle"
	"github.com/sed-evaluacion-desempeno/api/internal/phasedefinition"
	"github.com/sed-evaluacion-desempeno/api/internal/phasetransition"
)

// SeedCycle creates Cycle 2026, PhaseDefinitions, and PhaseTransitions.
func SeedCycle(ctx context.Context, client *internal.Client) error {
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

	orgID := seedID("org-sed")
	cycleID := seedID("cycle-2026")

	// 1. Cycle 2026
	startedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := tx.Cycle.Create().
		SetID(cycleID).
		SetYear(2026).
		SetCurrentPhase(cycle.CurrentPhaseAsignacion).
		SetStartedAt(startedAt).
		SetOrganizationID(orgID).
		Exec(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}

	// 2. PhaseDefinitions
	phaseDefs := []struct {
		id    string
		phase phasedefinition.Phase
		label string
		order int
	}{
		{id: "phase-def-asignacion", phase: phasedefinition.PhaseAsignacion, label: "Asignación de objetivos", order: 1},
		{id: "phase-def-avance", phase: phasedefinition.PhaseAvance, label: "Seguimiento de avance", order: 2},
		{id: "phase-def-cierre", phase: phasedefinition.PhaseCierre, label: "Cierre de evaluación", order: 3},
	}

	phaseDefIDs := make(map[string]string) // phase name -> def ID
	for _, pd := range phaseDefs {
		pdID := seedID(pd.id)

		allowedActors := []string{"employee", "manager", "rh"}
		allowedActions := []string{"view"}
		blockedActions := []string{}

		if pd.phase == phasedefinition.PhaseAsignacion {
			allowedActions = []string{"view", "set_goals", "edit_goals"}
		} else if pd.phase == phasedefinition.PhaseAvance {
			allowedActions = []string{"view", "update_progress"}
		} else if pd.phase == phasedefinition.PhaseCierre {
			allowedActions = []string{"view", "evaluate", "close"}
		}

		if err := tx.PhaseDefinition.Create().
			SetID(pdID).
			SetPhase(pd.phase).
			SetLabel(pd.label).
			SetOrder(pd.order).
			SetAllowedActors(allowedActors).
			SetAllowedActions(allowedActions).
			SetBlockedActions(blockedActions).
			SetCycleID(cycleID).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
		phaseDefIDs[string(pd.phase)] = pd.id
	}

	// 3. PhaseTransitions
	transitions := []struct {
		id        string
		fromPhase string
		toPhase   string
		trigger   phasetransition.Trigger
		fromDefID string
		toDefID   string
	}{
		{
			id: "trans-asignacion-avance",
			fromPhase: "asignacion",
			toPhase:   "avance",
			trigger:   phasetransition.TriggerManualRh,
			fromDefID: "phase-def-asignacion",
			toDefID:   "phase-def-avance",
		},
		{
			id: "trans-avance-cierre",
			fromPhase: "avance",
			toPhase:   "cierre",
			trigger:   phasetransition.TriggerManualRh,
			fromDefID: "phase-def-avance",
			toDefID:   "phase-def-cierre",
		},
	}

	for _, tr := range transitions {
		trID := seedID(tr.id)
		fromDefUUID := seedID(tr.fromDefID)
		toDefUUID := seedID(tr.toDefID)

		if err := tx.PhaseTransition.Create().
			SetID(trID).
			SetFromPhase(phasetransition.FromPhase(tr.fromPhase)).
			SetToPhase(phasetransition.ToPhase(tr.toPhase)).
			SetTrigger(tr.trigger).
			SetCycleID(cycleID).
			SetFromPhaseID(fromDefUUID).
			SetToPhaseID(toDefUUID).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("[seed] cycle: created 1 cycle, %d phase definitions, %d transitions", len(phaseDefs), len(transitions))
	return nil
}
