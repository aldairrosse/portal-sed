package seed

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/evaluatorscope"
)

// employeeDef holds data for creating an Employee.
type employeeDef struct {
	id          string
	orgNodeID   string
	managerID   string // empty string for top-level
	profileID   string
	firstName   string
	lastName    string
	email       string
	employeeNum string
}

func employeeDefs() []employeeDef {
	return []employeeDef{
		{id: "emp-dg-01", orgNodeID: "emp-dg-01", managerID: "", profileID: "director-general", firstName: "Carlos", lastName: "Mendoza", email: "carlos.mendoza@empresa.com", employeeNum: "EMP-001"},
		{id: "emp-director-01", orgNodeID: "emp-director-01", managerID: "emp-dg-01", profileID: "director", firstName: "Carmen", lastName: "Jiménez Castro", email: "carmen.jimenez@empresa.com", employeeNum: "EMP-002"},
		{id: "emp-jefe-01", orgNodeID: "emp-jefe-01", managerID: "emp-director-01", profileID: "jefe", firstName: "Carlos", lastName: "Rodríguez Pérez", email: "carlos.rodriguez@empresa.com", employeeNum: "EMP-003"},
		{id: "emp-colaborador-01", orgNodeID: "emp-colaborador-01", managerID: "emp-jefe-01", profileID: "colaborador", firstName: "Frankil Aldair", lastName: "Perez", email: "maria.lopez@empresa.com", employeeNum: "EMP-004"},
		{id: "emp-colaborador-02", orgNodeID: "emp-colaborador-02", managerID: "emp-jefe-01", profileID: "colaborador", firstName: "Ana", lastName: "López", email: "ana.lopez@empresa.com", employeeNum: "EMP-005"},
		{id: "emp-vendedor-01", orgNodeID: "emp-vendedor-01", managerID: "emp-jefe-01", profileID: "vendedor", firstName: "Ana", lastName: "Martínez Simón", email: "ana.martinez@empresa.com", employeeNum: "EMP-006"},
		{id: "emp-jefe-02", orgNodeID: "emp-jefe-02", managerID: "emp-director-01", profileID: "jefe", firstName: "Roberto", lastName: "Díaz", email: "roberto.diaz@empresa.com", employeeNum: "EMP-007"},
		{id: "emp-colaborador-03", orgNodeID: "emp-colaborador-03", managerID: "emp-jefe-02", profileID: "colaborador", firstName: "Sofía", lastName: "Martínez", email: "sofia.martinez@empresa.com", employeeNum: "EMP-008"},
		{id: "emp-vendedor-02", orgNodeID: "emp-vendedor-02", managerID: "emp-jefe-02", profileID: "vendedor", firstName: "Pedro", lastName: "Sánchez", email: "pedro.sanchez@empresa.com", employeeNum: "EMP-009"},
		{id: "emp-colaborador-director-01", orgNodeID: "emp-colaborador-director-01", managerID: "emp-director-01", profileID: "colaborador", firstName: "Lucía", lastName: "Hernández", email: "lucia.hernandez@empresa.com", employeeNum: "EMP-010"},
		{id: "emp-director-02", orgNodeID: "emp-director-02", managerID: "emp-dg-01", profileID: "director", firstName: "Miguel Ángel", lastName: "Reyes", email: "miguel.reyes@empresa.com", employeeNum: "EMP-011"},
		{id: "emp-jefe-03", orgNodeID: "emp-jefe-03", managerID: "emp-director-02", profileID: "jefe", firstName: "Fernando", lastName: "Castro", email: "fernando.castro@empresa.com", employeeNum: "EMP-012"},
		{id: "emp-colaborador-04", orgNodeID: "emp-colaborador-04", managerID: "emp-jefe-03", profileID: "colaborador", firstName: "Valeria", lastName: "Torres", email: "valeria.torres@empresa.com", employeeNum: "EMP-013"},
		{id: "emp-colaborador-05", orgNodeID: "emp-colaborador-05", managerID: "emp-jefe-03", profileID: "colaborador", firstName: "Diego", lastName: "Morales", email: "diego.morales@empresa.com", employeeNum: "EMP-014"},
		{id: "emp-colaborador-06", orgNodeID: "emp-colaborador-06", managerID: "emp-jefe-03", profileID: "colaborador", firstName: "Camila", lastName: "Vargas", email: "camila.vargas@empresa.com", employeeNum: "EMP-015"},
		{id: "emp-jefe-04", orgNodeID: "emp-jefe-04", managerID: "emp-director-02", profileID: "jefe", firstName: "Isabel", lastName: "Navarro", email: "isabel.navarro@empresa.com", employeeNum: "EMP-016"},
		{id: "emp-colaborador-07", orgNodeID: "emp-colaborador-07", managerID: "emp-jefe-04", profileID: "colaborador", firstName: "Andrés", lastName: "Gutiérrez", email: "andres.gutierrez@empresa.com", employeeNum: "EMP-017"},
		{id: "emp-colaborador-08", orgNodeID: "emp-colaborador-08", managerID: "emp-jefe-04", profileID: "colaborador", firstName: "Paula", lastName: "Ríos", email: "paula.rios@empresa.com", employeeNum: "EMP-018"},
		{id: "emp-director-03", orgNodeID: "emp-director-03", managerID: "emp-dg-01", profileID: "director", firstName: "Alejandra", lastName: "Romero", email: "alejandra.romero@empresa.com", employeeNum: "EMP-019"},
		{id: "emp-jefe-05", orgNodeID: "emp-jefe-05", managerID: "emp-director-03", profileID: "jefe", firstName: "Ricardo", lastName: "Peña", email: "ricardo.pena@empresa.com", employeeNum: "EMP-020"},
		{id: "emp-colaborador-09", orgNodeID: "emp-colaborador-09", managerID: "emp-jefe-05", profileID: "colaborador", firstName: "Mariana", lastName: "Flores", email: "mariana.flores@empresa.com", employeeNum: "EMP-021"},
		{id: "emp-colaborador-10", orgNodeID: "emp-colaborador-10", managerID: "emp-jefe-05", profileID: "colaborador", firstName: "Oscar", lastName: "Medina", email: "oscar.medina@empresa.com", employeeNum: "EMP-022"},
		{id: "emp-jefe-06", orgNodeID: "emp-jefe-06", managerID: "emp-director-03", profileID: "jefe", firstName: "Claudia", lastName: "Silva", email: "claudia.silva@empresa.com", employeeNum: "EMP-023"},
		{id: "emp-colaborador-11", orgNodeID: "emp-colaborador-11", managerID: "emp-jefe-06", profileID: "colaborador", firstName: "Tomás", lastName: "Guerrero", email: "tomas.guerrero@empresa.com", employeeNum: "EMP-024"},
		{id: "emp-colaborador-12", orgNodeID: "emp-colaborador-12", managerID: "emp-jefe-06", profileID: "colaborador", firstName: "Natalia", lastName: "Cruz", email: "natalia.cruz@empresa.com", employeeNum: "EMP-025"},
		{id: "emp-colaborador-13", orgNodeID: "emp-colaborador-13", managerID: "emp-jefe-06", profileID: "colaborador", firstName: "Javier", lastName: "Ramos", email: "javier.ramos@empresa.com", employeeNum: "EMP-026"},
		{id: "emp-colaborador-director-02", orgNodeID: "emp-colaborador-director-02", managerID: "emp-director-03", profileID: "colaborador", firstName: "Elena", lastName: "Vega", email: "elena.vega@empresa.com", employeeNum: "EMP-027"},
		{id: "emp-jefe-directo-dg-01", orgNodeID: "emp-jefe-directo-dg-01", managerID: "emp-dg-01", profileID: "jefe", firstName: "Roberto", lastName: "Fuentes", email: "roberto.fuentes@empresa.com", employeeNum: "EMP-028"},
		{id: "emp-colaborador-dg-01", orgNodeID: "emp-colaborador-dg-01", managerID: "emp-jefe-directo-dg-01", profileID: "colaborador", firstName: "Patricia", lastName: "León", email: "patricia.leon@empresa.com", employeeNum: "EMP-029"},
		{id: "emp-colaborador-dg-02", orgNodeID: "emp-colaborador-dg-02", managerID: "emp-jefe-directo-dg-01", profileID: "colaborador", firstName: "Manuel", lastName: "Ortega", email: "manuel.ortega@empresa.com", employeeNum: "EMP-030"},
		{id: "emp-rh-01", orgNodeID: "emp-rh-01", managerID: "emp-dg-01", profileID: "rh", firstName: "Laura", lastName: "Moreno Peña", email: "laura.moreno@empresa.com", employeeNum: "EMP-031"},
	}
}

// SeedEmployees creates Employee records and EvaluatorScope records.
func SeedEmployees(ctx context.Context, client *internal.Client) error {
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

	emps := employeeDefs()
	empUUIDs := make(map[string]uuid.UUID, len(emps))

	for _, e := range emps {
		eid := seedID(e.id)
		empUUIDs[e.id] = eid

		b := tx.Employee.Create().
			SetID(eid).
			SetFirstName(e.firstName).
			SetLastName(e.lastName).
			SetEmployeeNumber(e.employeeNum).
			SetEmail(e.email).
			SetIsActive(true).
			SetOrgNodeID(seedID(e.orgNodeID)).
			SetProfileID(seedID("profile-" + e.profileID)).
			SetCreatedBy(seedID("system")).
			SetUpdatedBy(seedID("system"))

		if e.managerID != "" {
			if mid, ok := empUUIDs[e.managerID]; ok {
				b = b.SetManagerID(mid)
			}
		}

		if err := b.Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// Create EvaluatorScope records for managers.
	// Use the manager relationship: each employee who is a manager of someone gets a scope.
	managerScopes := make(map[string]bool)
	for _, e := range emps {
		if e.managerID != "" {
			managerScopes[e.managerID] = true
		}
	}

	for mid := range managerScopes {
		scopeID := seedID("scope-" + mid)
		evaluatorUUID := empUUIDs[mid]

		b := tx.EvaluatorScope.Create().
			SetID(scopeID).
			SetScopeType(evaluatorscope.ScopeTypeTeam).
			SetEvaluatorID(evaluatorUUID)

		// Attach scope_data with a simple JSON object
		scopeData := map[string]interface{}{
			"managed_by": mid,
		}
		b = b.SetScopeData(scopeData)

		if err := b.Exec(ctx); err != nil {
			// Non-fatal: scope creation should not block employees
			log.Printf("[seed] employees: failed to create scope for %s: %v", mid, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("[seed] employees: created %d employees, %d evaluator scopes", len(emps), len(managerScopes))
	return nil
}


