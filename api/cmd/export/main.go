package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"github.com/sed-evaluacion-desempeno/api/internal"
)

type exportData struct {
	Organizations     []*internal.Organization `json:"organizations"`
	OrgNodes          []*internal.OrgNode      `json:"org_nodes"`
	EvaluationProfiles []*internal.EvaluationProfile `json:"evaluation_profiles"`
	Employees         []*internal.Employee     `json:"employees"`
}

func main() {
	outputPath := flag.String("output", "", "write JSON to file instead of stdout")
	flag.Parse()

	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("[export] DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("[export] failed to open database: %v", err)
	}
	defer db.Close()

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := internal.NewClient(internal.Driver(drv))
	defer client.Close()

	ctx := context.Background()

	data := exportData{}

	data.Organizations, err = client.Organization.Query().All(ctx)
	if err != nil {
		log.Fatalf("[export] failed to query organizations: %v", err)
	}

	data.OrgNodes, err = client.OrgNode.Query().All(ctx)
	if err != nil {
		log.Fatalf("[export] failed to query org_nodes: %v", err)
	}

	data.EvaluationProfiles, err = client.EvaluationProfile.Query().All(ctx)
	if err != nil {
		log.Fatalf("[export] failed to query evaluation_profiles: %v", err)
	}

	data.Employees, err = client.Employee.Query().All(ctx)
	if err != nil {
		log.Fatalf("[export] failed to query employees: %v", err)
	}

	log.Printf("[export] organizations: %d, org_nodes: %d, evaluation_profiles: %d, employees: %d",
		len(data.Organizations), len(data.OrgNodes), len(data.EvaluationProfiles), len(data.Employees))

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("[export] failed to marshal JSON: %v", err)
	}

	if *outputPath != "" {
		if err := os.WriteFile(*outputPath, jsonBytes, 0644); err != nil {
			log.Fatalf("[export] failed to write file: %v", err)
		}
		log.Printf("[export] written to %s", *outputPath)
	} else {
		fmt.Println(string(jsonBytes))
	}
}
