package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"github.com/sed-evaluacion-desempeno/api/internal/seed"
	syncsvc "github.com/sed-evaluacion-desempeno/api/internal/service/sync"
)

// One-off manual runner: replica el cron de server/main.go (sync.NewService(db).Run)
// sin scheduler. Propaga los logs de mobonet_sync (deshabilitado/reactivado/upsert).
//
// Puestos: por defecto NO se actualizan en existentes (solo status is_active
// + datos personales); los INSERT siempre asignan puesto. Para sincronizar
// puestos en existentes:
//
//	SYNC_PUESTOS=1 /opt/api/sync
//	/opt/api/sync --sync-puestos=1   (o -p=1; el flag sobreescribe el env)
//
// Exit codes (convención del proyecto: server/main.go usa log.Fatal → exit 1):
//   - DATABASE_URL vacío o error de DB/Run → log.Fatal (exit 1)
//   - MOBONET_URL, MOBONET_KEY y SEED_DB_URL todos vacíos → log.Fatal (exit 1), no se ejecuta upsert.
//     Se aborta en el runner para no tocar Service.Run (su mock mode queda solo como
//     fallback defensivo del cron).
func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	// SYNC_PUESTOS=1 sincroniza puestos (job_title/profile_id) en existentes;
	// default 0 (solo status + datos personales). El flag lo sobreescribe.
	syncPuestos := flag.Bool("sync-puestos", strings.TrimSpace(os.Getenv("SYNC_PUESTOS")) == "1",
		"sincroniza puestos (job_title/profile_id) en empleados existentes [SYNC_PUESTOS]")
	flag.BoolVar(syncPuestos, "p", *syncPuestos, "alias de --sync-puestos")
	flag.Parse()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("[sync] DATABASE_URL is required")
	}

	// Guard EN EL RUNNER (no modificar Run salvo justificación): todo vacío → abortar.
	mobonetURL := strings.TrimSpace(os.Getenv("MOBONET_URL"))
	mobonetKey := strings.TrimSpace(os.Getenv("MOBONET_KEY"))
	seedURL := strings.TrimSpace(os.Getenv("SEED_DB_URL"))
	if mobonetURL == "" && mobonetKey == "" && seedURL == "" {
		log.Fatal("[sync] MOBONET_URL/MOBONET_KEY y SEED_DB_URL vacíos: abortando, no se ejecuta upsert")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("[sync] failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("[sync] failed to ping database: %v", err)
	}

	if err := ensureEvaluationProfiles(ctx, db); err != nil {
		log.Fatalf("[sync] failed to ensure evaluation profiles: %v", err)
	}

	svc := syncsvc.NewService(db)
	svc.SyncPuestos = *syncPuestos
	disabled, err := svc.Run(ctx)
	if err != nil {
		log.Fatalf("[sync] run failed: %v", err)
	}
	log.Printf("[sync] done (disabled=%d sync_puestos=%t)", disabled, *syncPuestos)
}

// ensureEvaluationProfiles inserta idempotentemente los perfiles que
// resolveProfileName/jobTitleToProfileName pueden asignar (gerente/coordinador
// más fallback jefe/colaborador). Reusa los mismos IDs deterministas que
// cmd/import setupProfiles (seed.SeedID("profile-"+name)); seed.Run es no-op,
// por eso no se reutiliza.
func ensureEvaluationProfiles(ctx context.Context, db *sql.DB) error {
	profiles := []struct{ name, desc string }{
		{"colaborador", "Colaborador de la organización"},
		{"jefe", "Jefe de equipo o departamento"},
		{"gerente", "Gerente con alcance a su equipo"},
		{"coordinador", "Coordinador con alcance a su equipo"},
		{"director", "Director de área"},
		{"director-general", "Director general de la organización"},
		{"rh", "Recursos humanos"},
	}
	for _, p := range profiles {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO evaluation_profiles (id, name, description)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (id) DO NOTHING`,
			seed.SeedID("profile-"+p.name), p.name, p.desc,
		); err != nil {
			return err
		}
	}
	return nil
}
