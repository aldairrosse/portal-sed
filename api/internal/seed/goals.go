package seed

import (
	"context"
	"log"

	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/goal"
	"github.com/sed-evaluacion-desempeno/api/internal/kpi"
)

// SeedGoals creates GoalCategories, Goals, KPIs, GoalKpiLinks, and GoalAssignments.
func SeedGoals(ctx context.Context, client *internal.Client) error {
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

	// 1. GoalCategories (4 categories, linked to emp-dg-01 as owner)
	catDefs := []struct {
		id          string
		name        string
		description string
		weight      float64
	}{
		{id: "cat-ventas-finanzas", name: "Ventas y resultados financieros", description: "Objetivos relacionados con ingresos, rentabilidad y gestión financiera", weight: 40},
		{id: "cat-clientes-calidad", name: "Clientes y calidad", description: "Objetivos enfocados en satisfacción del cliente y calidad del servicio", weight: 30},
		{id: "cat-personas-equipo", name: "Personas y equipo", description: "Objetivos de desarrollo del talento, clima y liderazgo", weight: 20},
		{id: "cat-operaciones-procesos", name: "Operaciones y procesos", description: "Objetivos de eficiencia operativa, mejora continua y procesos", weight: 10},
	}
	for _, c := range catDefs {
		cid := seedID(c.id)
		if err := tx.GoalCategory.Create().
			SetID(cid).
			SetName(c.name).
			SetDescription(c.description).
			SetWeight(c.weight).
			SetEmployeeID(seedID("emp-dg-01")).
			SetCreatedBy(seedID("system")).
			SetUpdatedBy(seedID("system")).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 2. Goals (10 goals)
	goalDefs := []struct {
		id           string
		name         string
		description  string
		categoryID   string
		weight       float64
		unit         goal.Unit
		targetValue  float64
		currentValue float64
		state        goal.State
	}{
		{id: "goal-alcanzar-ventas", name: "Alcanzar meta de ventas mensuales", description: "Cumplir con la cuota de ventas asignada cada mes del período", categoryID: "cat-ventas-finanzas", weight: 50, unit: goal.UnitMoneda, targetValue: 1000000, currentValue: 650000, state: goal.StateEnSeguimiento},
		{id: "goal-mejorar-margen", name: "Mejorar margen operativo", description: "Incrementar el margen de utilidad operativa respecto al año anterior", categoryID: "cat-ventas-finanzas", weight: 30, unit: goal.UnitPorcentaje, targetValue: 25, currentValue: 18, state: goal.StateEnSeguimiento},
		{id: "goal-reducir-costos", name: "Reducir costos operativos", description: "Identificar y ejecutar oportunidades de ahorro en gastos operativos", categoryID: "cat-ventas-finanzas", weight: 20, unit: goal.UnitNumero, targetValue: 15, currentValue: 0, state: goal.StateFijada},
		{id: "goal-mejorar-satisfaccion", name: "Mejorar satisfacción del cliente", description: "Alcanzar el puntaje objetivo en encuestas de satisfacción", categoryID: "cat-clientes-calidad", weight: 40, unit: goal.UnitPorcentaje, targetValue: 90, currentValue: 82, state: goal.StateEnSeguimiento},
		{id: "goal-reducir-quejas", name: "Reducir quejas y reclamaciones", description: "Disminuir el número de quejas recibidas en un 20% respecto al período anterior", categoryID: "cat-clientes-calidad", weight: 30, unit: goal.UnitPorcentaje, targetValue: 20, currentValue: 35, state: goal.StateEnSeguimiento},
		{id: "goal-fidelizacion", name: "Implementar programa de fidelización", description: "Diseñar y lanzar un programa de fidelización para clientes frecuentes", categoryID: "cat-clientes-calidad", weight: 30, unit: goal.UnitPorcentaje, targetValue: 80, currentValue: 10, state: goal.StateEnSeguimiento},
		{id: "goal-reducir-rotacion", name: "Reducir rotación de personal", description: "Disminuir la rotación voluntaria implementando planes de retención", categoryID: "cat-personas-equipo", weight: 50, unit: goal.UnitPorcentaje, targetValue: 10, currentValue: 8, state: goal.StateEnSeguimiento},
		{id: "goal-capacitaciones", name: "Completar capacitaciones del equipo", description: "Asegurar que cada miembro complete al menos 40 horas de capacitación", categoryID: "cat-personas-equipo", weight: 50, unit: goal.UnitNumero, targetValue: 40, currentValue: 20, state: goal.StateEnSeguimiento},
		{id: "goal-optimizar-inventario", name: "Optimizar rotación de inventario", description: "Mejorar el índice de rotación de inventario en los puntos de venta", categoryID: "cat-operaciones-procesos", weight: 60, unit: goal.UnitNumero, targetValue: 4.5, currentValue: 0, state: goal.StateFijada},
		{id: "goal-reducir-ausentismo", name: "Reducir ausentismo", description: "Implementar medidas para reducir el ausentismo por debajo del 3%", categoryID: "cat-operaciones-procesos", weight: 40, unit: goal.UnitPorcentaje, targetValue: 3, currentValue: 100, state: goal.StateEnSeguimiento},
	}
	for _, g := range goalDefs {
		gid := seedID(g.id)

		// Use the current time as created_at/updated_at reference
		b := tx.Goal.Create().
			SetID(gid).
			SetName(g.name).
			SetDescription(g.description).
			SetUnit(g.unit).
			SetWeight(g.weight).
			SetTargetValue(g.targetValue).
			SetCurrentValue(g.currentValue).
			SetState(g.state).
			SetCategoryID(seedID(g.categoryID))

		// Set audit fields (created_by, updated_by) to the DG employee
		if err := b.
			SetCreatedBy(seedID("emp-dg-01")).
			SetUpdatedBy(seedID("emp-dg-01")).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 3. KPIs (6 KPIs)
	kpiDefs := []struct {
		id          string
		name        string
		description string
		unit        kpi.Unit
	}{
		{id: "kpi-ventas-mensuales", name: "Ventas mensuales", description: "Ingresos totales por ventas en el mes", unit: kpi.UnitMoneda},
		{id: "kpi-satisfaccion-cliente", name: "Satisfacción del cliente", description: "Porcentaje de clientes satisfechos según encuesta NPS", unit: kpi.UnitPorcentaje},
		{id: "kpi-productividad", name: "Productividad del equipo", description: "Porcentaje de metas del equipo alcanzadas en el período", unit: kpi.UnitPorcentaje},
		{id: "kpi-rotacion-inventario", name: "Rotación de inventario", description: "Veces que se renueva el inventario en el período", unit: kpi.UnitNumero},
		{id: "kpi-ausentismo", name: "Ausentismo", description: "Porcentaje de horas no trabajadas sobre horas programadas", unit: kpi.UnitPorcentaje},
		{id: "kpi-margen-operativo", name: "Margen operativo", description: "Margen de utilidad operativa como porcentaje de ingresos", unit: kpi.UnitPorcentaje},
	}
	for _, k := range kpiDefs {
		kid := seedID(k.id)
		if err := tx.KPI.Create().
			SetID(kid).
			SetName(k.name).
			SetDescription(k.description).
			SetUnit(k.unit).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 4. GoalKpiLinks (11 links)
	linkDefs := []struct {
		goalID string
		kpiID  string
	}{
		{goalID: "goal-alcanzar-ventas", kpiID: "kpi-ventas-mensuales"},
		{goalID: "goal-alcanzar-ventas", kpiID: "kpi-margen-operativo"},
		{goalID: "goal-mejorar-margen", kpiID: "kpi-margen-operativo"},
		{goalID: "goal-reducir-costos", kpiID: "kpi-rotacion-inventario"},
		{goalID: "goal-mejorar-satisfaccion", kpiID: "kpi-satisfaccion-cliente"},
		{goalID: "goal-reducir-quejas", kpiID: "kpi-satisfaccion-cliente"},
		{goalID: "goal-fidelizacion", kpiID: "kpi-satisfaccion-cliente"},
		{goalID: "goal-reducir-rotacion", kpiID: "kpi-ausentismo"},
		{goalID: "goal-capacitaciones", kpiID: "kpi-productividad"},
		{goalID: "goal-optimizar-inventario", kpiID: "kpi-rotacion-inventario"},
		{goalID: "goal-reducir-ausentismo", kpiID: "kpi-ausentismo"},
	}
	for _, l := range linkDefs {
		if err := tx.GoalKpiLink.Create().
			SetGoalID(seedID(l.goalID)).
			SetKpiID(seedID(l.kpiID)).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 5. GoalAssignments (10 assignments from assignments.json)
	assignDefs := []struct {
		id         string
		employeeID string
		cycleID    string
	}{
		{id: "asign-colaborador-01", employeeID: "emp-colaborador-01"},
		{id: "asign-colaborador-02", employeeID: "emp-colaborador-02"},
		{id: "asign-vendedor-01", employeeID: "emp-vendedor-01"},
		{id: "asign-jefe-01", employeeID: "emp-jefe-01"},
		{id: "asign-gerente-tienda-01", employeeID: "emp-gerente-tienda-01"},
		{id: "asign-divisional-01", employeeID: "emp-divisional-01"},
		{id: "asign-regional-01", employeeID: "emp-regional-01"},
		{id: "asign-director-01", employeeID: "emp-director-01"},
		{id: "asign-rh-01", employeeID: "emp-rh-01"},
		{id: "asign-dg-01", employeeID: "emp-dg-01"},
	}
	// For employee IDs that don't exist in the org tree (from assignments.json), skip them
	existingEmpIDs := make(map[string]bool)
	for _, e := range employeeDefs() {
		existingEmpIDs[e.id] = true
	}

	for _, a := range assignDefs {
		aid := seedID(a.id)

		if !existingEmpIDs[a.employeeID] {
			log.Printf("[seed] goals: skipping assignment %s (employee %s not seeded)", a.id, a.employeeID)
			continue
		}

		if err := tx.GoalAssignment.Create().
			SetID(aid).
			SetEmployeeID(seedID(a.employeeID)).
			SetCycleID(cycleID).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("[seed] goals: created %d categories, %d goals, %d KPIs, %d links, %d assignments",
		len(catDefs), len(goalDefs), len(kpiDefs), len(linkDefs), len(assignDefs))
	return nil
}
