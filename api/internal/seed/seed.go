package seed

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/sed-evaluacion-desempeno/api/internal"
)

// Deterministic namespace for seed UUIDs.
var ns = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

// SeedID generates a deterministic UUID from a string.
// Used by seed, import, export, and integration tests.
func SeedID(s string) uuid.UUID {
	return uuid.NewSHA1(ns, []byte(s))
}

// Run is a no-op. Data is now imported via cmd/import from an external DB.
func Run(_ context.Context, _ *internal.Client) error {
	log.Println("[seed] data is imported via cmd/import — Run() is a no-op")
	return nil
}
