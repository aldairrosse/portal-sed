package seed

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/sed-evaluacion-desempeno/api/internal"
)

// Run executes all domain seeders in FK-safe order.
// Each domain runs in its own transaction; errors are logged, and execution continues.
func Run(ctx context.Context, client *internal.Client) error {
	// Check if DB already has data.
	count, err := client.Employee.Query().Count(ctx)
	if err == nil && count > 0 {
		log.Println("[seed] database already has data, skipping")
		return nil
	}

	shouldSeed := flagSeeded() || os.Getenv("SEED_ON_START") == "true"
	if !shouldSeed && count > 0 {
		log.Println("[seed] skipping (data exists and no seed flag/env set)")
		return nil
	}

	seeders := []struct {
		name string
		fn   func(context.Context, *internal.Client) error
	}{
		{"org", SeedOrg},
		{"employees", SeedEmployees},
		{"cycle", SeedCycle},
		{"competency", SeedCompetency},
		{"goals", SeedGoals},
		{"evaluation", SeedEvaluation},
		{"ninebox", SeedNineBox},
	}

	for _, s := range seeders {
		if err := s.fn(ctx, client); err != nil {
			log.Printf("[seed] %s: %v (continuing)", s.name, err)
		}
	}

	return nil
}

// flagSeeded returns true if the --seed flag was provided.
func flagSeeded() bool {
	seeded := false
	if !flag.Parsed() {
		flag.Parse()
	}
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "seed" {
			seeded = true
		}
	})
	return seeded
}

func init() {
	flag.Bool("seed", false, "seed the database with fixture data")
}
