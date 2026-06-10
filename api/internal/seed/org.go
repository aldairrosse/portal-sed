package seed

import (
	"context"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
	"github.com/sed-evaluacion-desempeno/api/internal/orgnode"
)

// Deterministic namespace for seed UUIDs.
var ns = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

func seedID(s string) uuid.UUID {
	return uuid.NewSHA1(ns, []byte(s))
}

// orgNodeDef describes a node in the org tree.
type orgNodeDef struct {
	id        string
	name      string
	profileID string // evaluation profile name
	parentID  string // empty string for root
	nodeType  string // "corporate" or "retail"
}

// orgTree returns the org hierarchy as flat lists ordered parent-first.
func orgTree() []orgNodeDef {
	return []orgNodeDef{
		{id: "emp-dg-01", name: "Carlos Mendoza", profileID: "director-general", parentID: "", nodeType: "corporate"},
		{id: "emp-director-01", name: "Carmen Jiménez Castro", profileID: "director", parentID: "emp-dg-01", nodeType: "corporate"},
		{id: "emp-jefe-01", name: "Carlos Rodríguez Pérez", profileID: "jefe", parentID: "emp-director-01", nodeType: "corporate"},
		{id: "emp-colaborador-01", name: "María López García", profileID: "colaborador", parentID: "emp-jefe-01", nodeType: "corporate"},
		{id: "emp-colaborador-02", name: "Ana López", profileID: "colaborador", parentID: "emp-jefe-01", nodeType: "corporate"},
		{id: "emp-vendedor-01", name: "Ana Martínez Simón", profileID: "vendedor", parentID: "emp-jefe-01", nodeType: "retail"},
		{id: "emp-jefe-02", name: "Roberto Díaz", profileID: "jefe", parentID: "emp-director-01", nodeType: "corporate"},
		{id: "emp-colaborador-03", name: "Sofía Martínez", profileID: "colaborador", parentID: "emp-jefe-02", nodeType: "corporate"},
		{id: "emp-vendedor-02", name: "Pedro Sánchez", profileID: "vendedor", parentID: "emp-jefe-02", nodeType: "retail"},
		{id: "emp-colaborador-director-01", name: "Lucía Hernández", profileID: "colaborador", parentID: "emp-director-01", nodeType: "corporate"},
		{id: "emp-director-02", name: "Miguel Ángel Reyes", profileID: "director", parentID: "emp-dg-01", nodeType: "corporate"},
		{id: "emp-jefe-03", name: "Fernando Castro", profileID: "jefe", parentID: "emp-director-02", nodeType: "corporate"},
		{id: "emp-colaborador-04", name: "Valeria Torres", profileID: "colaborador", parentID: "emp-jefe-03", nodeType: "corporate"},
		{id: "emp-colaborador-05", name: "Diego Morales", profileID: "colaborador", parentID: "emp-jefe-03", nodeType: "corporate"},
		{id: "emp-colaborador-06", name: "Camila Vargas", profileID: "colaborador", parentID: "emp-jefe-03", nodeType: "corporate"},
		{id: "emp-jefe-04", name: "Isabel Navarro", profileID: "jefe", parentID: "emp-director-02", nodeType: "corporate"},
		{id: "emp-colaborador-07", name: "Andrés Gutiérrez", profileID: "colaborador", parentID: "emp-jefe-04", nodeType: "corporate"},
		{id: "emp-colaborador-08", name: "Paula Ríos", profileID: "colaborador", parentID: "emp-jefe-04", nodeType: "corporate"},
		{id: "emp-director-03", name: "Alejandra Romero", profileID: "director", parentID: "emp-dg-01", nodeType: "corporate"},
		{id: "emp-jefe-05", name: "Ricardo Peña", profileID: "jefe", parentID: "emp-director-03", nodeType: "corporate"},
		{id: "emp-colaborador-09", name: "Mariana Flores", profileID: "colaborador", parentID: "emp-jefe-05", nodeType: "corporate"},
		{id: "emp-colaborador-10", name: "Oscar Medina", profileID: "colaborador", parentID: "emp-jefe-05", nodeType: "corporate"},
		{id: "emp-jefe-06", name: "Claudia Silva", profileID: "jefe", parentID: "emp-director-03", nodeType: "corporate"},
		{id: "emp-colaborador-11", name: "Tomás Guerrero", profileID: "colaborador", parentID: "emp-jefe-06", nodeType: "corporate"},
		{id: "emp-colaborador-12", name: "Natalia Cruz", profileID: "colaborador", parentID: "emp-jefe-06", nodeType: "corporate"},
		{id: "emp-colaborador-13", name: "Javier Ramos", profileID: "colaborador", parentID: "emp-jefe-06", nodeType: "corporate"},
		{id: "emp-colaborador-director-02", name: "Elena Vega", profileID: "colaborador", parentID: "emp-director-03", nodeType: "corporate"},
		{id: "emp-jefe-directo-dg-01", name: "Roberto Fuentes", profileID: "jefe", parentID: "emp-dg-01", nodeType: "corporate"},
		{id: "emp-colaborador-dg-01", name: "Patricia León", profileID: "colaborador", parentID: "emp-jefe-directo-dg-01", nodeType: "corporate"},
		{id: "emp-colaborador-dg-02", name: "Manuel Ortega", profileID: "colaborador", parentID: "emp-jefe-directo-dg-01", nodeType: "corporate"},
		{id: "emp-rh-01", name: "Laura Moreno Peña", profileID: "rh", parentID: "emp-dg-01", nodeType: "corporate"},
	}
}

// nodeCode derives a short code from a node ID, e.g. "emp-dg-01" -> "DG-01".
func nodeCode(id string) string {
	parts := strings.SplitN(id, "-", 3)
	if len(parts) < 3 {
		return id
	}
	// parts[1] is like "dg", parts[2] like "01"
	// For nodes where parts[1] is composite like "colaborador-director", keep it
	return strings.ToUpper(parts[1]) + "-" + parts[2]
}

// profileDef holds data for an EvaluationProfile.
type profileDef struct {
	name        string
	description string
}

func profileDefs() []profileDef {
	return []profileDef{
		{name: "colaborador", description: "Colaborador de la organización"},
		{name: "jefe", description: "Jefe de equipo o departamento"},
		{name: "vendedor", description: "Vendedor en tienda o canal"},
		{name: "gerente-tienda", description: "Gerente de tienda retail"},
		{name: "divisional", description: "Divisional o jefe de área"},
		{name: "regional", description: "Regional de zona"},
		{name: "director", description: "Director de área"},
		{name: "director-general", description: "Director general de la organización"},
		{name: "rh", description: "Recursos humanos"},
	}
}

// SeedOrg creates Organization, OrgNodes, and EvaluationProfiles.
func SeedOrg(ctx context.Context, client *internal.Client) error {
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

	// 1. Organization
	orgID := seedID("org-sed")
	if err := tx.Organization.Create().
		SetID(orgID).
		SetName("SED Organización").
		SetSlug("sed-org").
		Exec(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}

	// 2. EvaluationProfiles (needed before OrgNodes that reference them via employees)
	profileIDs := make(map[string]uuid.UUID)
	for _, p := range profileDefs() {
		pid := seedID("profile-" + p.name)
		profileIDs[p.name] = pid
		if err := tx.EvaluationProfile.Create().
			SetID(pid).
			SetName(p.name).
			SetDescription(p.description).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 3. OrgNodes (parent references by ID, so we insert in parent-first order)
	nodes := orgTree()
	nodeIDs := make(map[string]uuid.UUID, len(nodes))
	for _, n := range nodes {
		nid := seedID(n.id)
		nodeIDs[n.id] = nid

		b := tx.OrgNode.Create().
			SetID(nid).
			SetName(n.name).
			SetType(orgnode.Type(n.nodeType)).
			SetCode(nodeCode(n.id)).
			SetOrganizationID(orgID)

		if n.parentID != "" {
			if parentUUID, ok := nodeIDs[n.parentID]; ok {
				b = b.SetParentID(parentUUID)
			}
		}
		if err := b.Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("[seed] org: created 1 organization, %d org nodes, %d profiles", len(nodes), len(profileDefs()))
	return nil
}
