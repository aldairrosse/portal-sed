package seed

import (
	"context"
	"log"

	"github.com/sed-evaluacion-desempeno/api/internal"
)

// pillarDef describes a competency pillar.
type pillarDef struct {
	id          string
	name        string
	description string
}

// competencyDef describes a competency within a pillar.
type competencyDef struct {
	id          string
	name        string
	description string
	pillarID    string
}

// scaleCritDef describes a scale criterion for a competency x pillar x level.
type scaleCritDef struct {
	id           string
	competencyID string
	pillarID     string
	level        int
	description  string
}

// acceptanceDef describes a competency acceptance level for a profile.
type acceptanceDef struct {
	competencyID string
	profileID    string
	level        int
}

// SeedCompetency creates LevelDefinitions, Pillars, Competencies, ScaleCriteria,
// and CompetencyAcceptanceLevels.
func SeedCompetency(ctx context.Context, client *internal.Client) error {
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

	// 1. LevelDefinitions (global 1-5 scale)
	levels := []struct {
		level       int
		label       string
		description string
	}{
		{level: 1, label: "No aceptable", description: "No cumple con el estándar mínimo esperado"},
		{level: 2, label: "En desarrollo", description: "Muestra avances pero requiere mejora consistente"},
		{level: 3, label: "Cumple expectativas", description: "Cumple con lo esperado para su rol"},
		{level: 4, label: "Sobre expectativas", description: "Supera consistentemente lo esperado"},
		{level: 5, label: "Excepcional", description: "Es referente y modelo a seguir"},
	}
	for _, l := range levels {
		if err := tx.LevelDefinition.Create().
			SetLevel(l.level).
			SetLabel(l.label).
			SetDescription(l.description).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 2. Pillars
	pillars := []pillarDef{
		{id: "pilar-liderazgo", name: "Liderazgo", description: "Capacidad de guiar, motivar y desarrollar equipos de alto rendimiento"},
		{id: "pilar-tecnico", name: "Técnico", description: "Conocimientos y habilidades técnicas necesarias para el rol"},
		{id: "pilar-comportamental", name: "Comportamental", description: "Actitudes y comportamientos esperados en el entorno laboral"},
	}
	for _, p := range pillars {
		pid := seedID(p.id)
		if err := tx.Pillar.Create().
			SetID(pid).
			SetName(p.name).
			SetDescription(p.description).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 3. Competencies
	competencies := []competencyDef{
		{id: "comp-comunicacion", name: "Comunicación efectiva", description: "Transmite ideas con claridad y escucha activamente", pillarID: "pilar-liderazgo"},
		{id: "comp-desarrollo-equipo", name: "Desarrollo de equipo", description: "Fomenta el crecimiento profesional de los miembros del equipo", pillarID: "pilar-liderazgo"},
		{id: "comp-toma-decisiones", name: "Toma de decisiones", description: "Analiza opciones y decide con criterio y oportunidad", pillarID: "pilar-liderazgo"},
		{id: "comp-dominio-herramientas", name: "Dominio de herramientas", description: "Maneja con solvencia las herramientas y sistemas del puesto", pillarID: "pilar-tecnico"},
		{id: "comp-resolucion-problemas", name: "Resolución de problemas", description: "Identifica causas raíz y propone soluciones efectivas", pillarID: "pilar-tecnico"},
		{id: "comp-colaboracion", name: "Colaboración", description: "Trabaja en equipo y contribuye a un ambiente positivo", pillarID: "pilar-comportamental"},
		{id: "comp-adaptabilidad", name: "Adaptabilidad", description: "Se ajusta a cambios con flexibilidad y actitud constructiva", pillarID: "pilar-comportamental"},
		{id: "comp-responsabilidad", name: "Responsabilidad", description: "Cumple compromisos y asume las consecuencias de sus actos", pillarID: "pilar-comportamental"},
	}
	for _, c := range competencies {
		cid := seedID(c.id)
		if err := tx.Competency.Create().
			SetID(cid).
			SetName(c.name).
			SetDescription(c.description).
			SetPillarID(seedID(c.pillarID)).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 4. ScaleCriteria (40 entries: 8 competencies × 5 levels)
	scaleCriteria := []scaleCritDef{
		// Comunicación (pilar-liderazgo)
		{id: "sc-comp-comunicacion-pilar-liderazgo-1", competencyID: "comp-comunicacion", pillarID: "pilar-liderazgo", level: 1, description: "Se expresa de forma confusa y no escucha a los demás"},
		{id: "sc-comp-comunicacion-pilar-liderazgo-2", competencyID: "comp-comunicacion", pillarID: "pilar-liderazgo", level: 2, description: "Se expresa con poca claridad y escucha de forma selectiva"},
		{id: "sc-comp-comunicacion-pilar-liderazgo-3", competencyID: "comp-comunicacion", pillarID: "pilar-liderazgo", level: 3, description: "Se expresa con claridad y escucha activamente en la mayoría de las situaciones"},
		{id: "sc-comp-comunicacion-pilar-liderazgo-4", competencyID: "comp-comunicacion", pillarID: "pilar-liderazgo", level: 4, description: "Comunica ideas de forma efectiva y adapta su mensaje según la audiencia"},
		{id: "sc-comp-comunicacion-pilar-liderazgo-5", competencyID: "comp-comunicacion", pillarID: "pilar-liderazgo", level: 5, description: "Inspira y persuade mediante una comunicación excepcional en toda la organización"},

		// Desarrollo de equipo (pilar-liderazgo)
		{id: "sc-comp-desarrollo-equipo-pilar-liderazgo-1", competencyID: "comp-desarrollo-equipo", pillarID: "pilar-liderazgo", level: 1, description: "No muestra interés en el desarrollo de otros"},
		{id: "sc-comp-desarrollo-equipo-pilar-liderazgo-2", competencyID: "comp-desarrollo-equipo", pillarID: "pilar-liderazgo", level: 2, description: "Ocasionalmente brinda retroalimentación a sus pares"},
		{id: "sc-comp-desarrollo-equipo-pilar-liderazgo-3", competencyID: "comp-desarrollo-equipo", pillarID: "pilar-liderazgo", level: 3, description: "Retroalimenta y apoya el crecimiento de los miembros del equipo"},
		{id: "sc-comp-desarrollo-equipo-pilar-liderazgo-4", competencyID: "comp-desarrollo-equipo", pillarID: "pilar-liderazgo", level: 4, description: "Busca activamente oportunidades de desarrollo y delega con propósito"},
		{id: "sc-comp-desarrollo-equipo-pilar-liderazgo-5", competencyID: "comp-desarrollo-equipo", pillarID: "pilar-liderazgo", level: 5, description: "Crea una cultura de aprendizaje continuo y forma líderes en la organización"},

		// Toma de decisiones (pilar-liderazgo)
		{id: "sc-comp-toma-decisiones-pilar-liderazgo-1", competencyID: "comp-toma-decisiones", pillarID: "pilar-liderazgo", level: 1, description: "Evita tomar decisiones o las pospone sin justificación"},
		{id: "sc-comp-toma-decisiones-pilar-liderazgo-2", competencyID: "comp-toma-decisiones", pillarID: "pilar-liderazgo", level: 2, description: "Toma decisiones impulsivas sin analizar opciones"},
		{id: "sc-comp-toma-decisiones-pilar-liderazgo-3", competencyID: "comp-toma-decisiones", pillarID: "pilar-liderazgo", level: 3, description: "Analiza opciones y toma decisiones informadas en los plazos esperados"},
		{id: "sc-comp-toma-decisiones-pilar-liderazgo-4", competencyID: "comp-toma-decisiones", pillarID: "pilar-liderazgo", level: 4, description: "Evalúa riesgos y beneficios con profundidad y decide con agilidad"},
		{id: "sc-comp-toma-decisiones-pilar-liderazgo-5", competencyID: "comp-toma-decisiones", pillarID: "pilar-liderazgo", level: 5, description: "Toma decisiones estratégicas de alto impacto con información incompleta y acierta consistentemente"},

		// Dominio de herramientas (pilar-tecnico)
		{id: "sc-comp-dominio-herramientas-pilar-tecnico-1", competencyID: "comp-dominio-herramientas", pillarID: "pilar-tecnico", level: 1, description: "Desconoce las herramientas básicas del puesto y requiere supervisión constante"},
		{id: "sc-comp-dominio-herramientas-pilar-tecnico-2", competencyID: "comp-dominio-herramientas", pillarID: "pilar-tecnico", level: 2, description: "Usa herramientas básicas con ayuda frecuente y comete errores"},
		{id: "sc-comp-dominio-herramientas-pilar-tecnico-3", competencyID: "comp-dominio-herramientas", pillarID: "pilar-tecnico", level: 3, description: "Maneja las herramientas del puesto de forma autónoma y eficiente"},
		{id: "sc-comp-dominio-herramientas-pilar-tecnico-4", competencyID: "comp-dominio-herramientas", pillarID: "pilar-tecnico", level: 4, description: "Domina las herramientas y encuentra formas de optimizar su uso"},
		{id: "sc-comp-dominio-herramientas-pilar-tecnico-5", competencyID: "comp-dominio-herramientas", pillarID: "pilar-tecnico", level: 5, description: "Es referente técnico y propone mejoras que elevan la productividad del equipo"},

		// Resolución de problemas (pilar-tecnico)
		{id: "sc-comp-resolucion-problemas-pilar-tecnico-1", competencyID: "comp-resolucion-problemas", pillarID: "pilar-tecnico", level: 1, description: "No identifica problemas ni propone soluciones por sí mismo"},
		{id: "sc-comp-resolucion-problemas-pilar-tecnico-2", competencyID: "comp-resolucion-problemas", pillarID: "pilar-tecnico", level: 2, description: "Identifica problemas pero depende de otros para resolverlos"},
		{id: "sc-comp-resolucion-problemas-pilar-tecnico-3", competencyID: "comp-resolucion-problemas", pillarID: "pilar-tecnico", level: 3, description: "Resuelve problemas habituales de forma autónoma y eficaz"},
		{id: "sc-comp-resolucion-problemas-pilar-tecnico-4", competencyID: "comp-resolucion-problemas", pillarID: "pilar-tecnico", level: 4, description: "Resuelve problemas complejos con análisis estructurado y soluciones creativas"},
		{id: "sc-comp-resolucion-problemas-pilar-tecnico-5", competencyID: "comp-resolucion-problemas", pillarID: "pilar-tecnico", level: 5, description: "Anticipa problemas y diseña soluciones sistémicas que benefician a toda la organización"},

		// Colaboración (pilar-comportamental)
		{id: "sc-comp-colaboracion-pilar-comportamental-1", competencyID: "comp-colaboracion", pillarID: "pilar-comportamental", level: 1, description: "Trabaja de forma aislada y no comparte información con el equipo"},
		{id: "sc-comp-colaboracion-pilar-comportamental-2", competencyID: "comp-colaboracion", pillarID: "pilar-comportamental", level: 2, description: "Colabora solo cuando se le solicita y de forma limitada"},
		{id: "sc-comp-colaboracion-pilar-comportamental-3", competencyID: "comp-colaboracion", pillarID: "pilar-comportamental", level: 3, description: "Colabora activamente y comparte información de forma oportuna"},
		{id: "sc-comp-colaboracion-pilar-comportamental-4", competencyID: "comp-colaboracion", pillarID: "pilar-comportamental", level: 4, description: "Fomenta la colaboración y genera sinergia en el equipo"},
		{id: "sc-comp-colaboracion-pilar-comportamental-5", competencyID: "comp-colaboracion", pillarID: "pilar-comportamental", level: 5, description: "Construye redes de colaboración que trascienden su equipo y área"},

		// Adaptabilidad (pilar-comportamental)
		{id: "sc-comp-adaptabilidad-pilar-comportamental-1", competencyID: "comp-adaptabilidad", pillarID: "pilar-comportamental", level: 1, description: "Se resiste al cambio y se frustra ante situaciones nuevas"},
		{id: "sc-comp-adaptabilidad-pilar-comportamental-2", competencyID: "comp-adaptabilidad", pillarID: "pilar-comportamental", level: 2, description: "Acepta el cambio con reticencia y tarda en adaptarse"},
		{id: "sc-comp-adaptabilidad-pilar-comportamental-3", competencyID: "comp-adaptabilidad", pillarID: "pilar-comportamental", level: 3, description: "Se adapta a los cambios con una actitud positiva y aprende rápido"},
		{id: "sc-comp-adaptabilidad-pilar-comportamental-4", competencyID: "comp-adaptabilidad", pillarID: "pilar-comportamental", level: 4, description: "Abraza el cambio y ayuda a otros a navegar la transición"},
		{id: "sc-comp-adaptabilidad-pilar-comportamental-5", competencyID: "comp-adaptabilidad", pillarID: "pilar-comportamental", level: 5, description: "Lidera el cambio y convierte la incertidumbre en oportunidad para el equipo"},

		// Responsabilidad (pilar-comportamental)
		{id: "sc-comp-responsabilidad-pilar-comportamental-1", competencyID: "comp-responsabilidad", pillarID: "pilar-comportamental", level: 1, description: "No cumple plazos ni asume la responsabilidad de sus acciones"},
		{id: "sc-comp-responsabilidad-pilar-comportamental-2", competencyID: "comp-responsabilidad", pillarID: "pilar-comportamental", level: 2, description: "Cumple de forma inconsistente y busca excusas ante errores"},
		{id: "sc-comp-responsabilidad-pilar-comportamental-3", competencyID: "comp-responsabilidad", pillarID: "pilar-comportamental", level: 3, description: "Cumple sus compromisos y asume sus resultados con honestidad"},
		{id: "sc-comp-responsabilidad-pilar-comportamental-4", competencyID: "comp-responsabilidad", pillarID: "pilar-comportamental", level: 4, description: "Supera expectativas en sus compromisos y promueve la rendición de cuentas"},
		{id: "sc-comp-responsabilidad-pilar-comportamental-5", competencyID: "comp-responsabilidad", pillarID: "pilar-comportamental", level: 5, description: "Es modelo de integridad y responsabilidad, inspirando a otros con su ejemplo"},
	}
	for _, sc := range scaleCriteria {
		scid := seedID(sc.id)
		if err := tx.ScaleCriterion.Create().
			SetID(scid).
			SetLevel(sc.level).
			SetDescription(sc.description).
			SetCompetencyID(seedID(sc.competencyID)).
			SetPillarID(seedID(sc.pillarID)).
			Exec(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 5. CompetencyAcceptanceLevels (8 competencies × 8 profiles = 64 entries, all level 3)
	// Profiles with acceptance levels: colaborador, jefe, vendedor, gerente-tienda, divisional, regional, director, rh
	profiles := []string{"colaborador", "jefe", "vendedor", "gerente-tienda", "divisional", "regional", "director", "rh"}
	allCompetencies := []string{"comp-comunicacion", "comp-desarrollo-equipo", "comp-toma-decisiones", "comp-dominio-herramientas", "comp-resolucion-problemas", "comp-colaboracion", "comp-adaptabilidad", "comp-responsabilidad"}

	count := 0
	for _, comp := range allCompetencies {
		for _, prof := range profiles {
			acceptID := seedID("accept-" + comp + "-" + prof)
			if err := tx.CompetencyAcceptanceLevel.Create().
				SetID(acceptID).
				SetLevel(3).
				SetCompetencyID(seedID(comp)).
				SetProfileID(seedID("profile-" + prof)).
				Exec(ctx); err != nil {
				_ = tx.Rollback()
				return err
			}
			count++
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("[seed] competency: created 5 level definitions, 3 pillars, %d competencies, %d scale criteria, %d acceptance levels",
		len(competencies), len(scaleCriteria), count)
	return nil
}
