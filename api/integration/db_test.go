package integration

import (
	"context"
	"testing"
	"time"
)

// TestDatabaseConnection verifies that we can connect to PostgreSQL.
func TestDatabaseConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.DB.PingContext(ctx); err != nil {
		t.Fatalf("database ping failed: %v", err)
	}
}

// TestDatabaseMigration verifies that all tables were created by auto-migrate.
func TestDatabaseMigration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	// Query information_schema to verify tables exist
	tables := []string{
		"organizations",
		"org_nodes",
		"evaluation_profiles",
		"employees",
		"evaluator_scopes",
		"cycles",
		"phase_definitions",
		"phase_transitions",
		"pillars",
		"competencies",
		"scale_criterions",
		"competency_acceptance_levels",
		"goals",
		"goal_categories",
		"kp_is",
		"goal_kpi_links",
		"goal_assignments",
		"evaluations",
		"evaluation_competencies",
		"evaluation_goals",
		"nine_box_scales",
		"nine_box_quadrants",
		"nine_box_matrixes",
		"nine_box_entries",
	}

	for _, table := range tables {
		t.Run(table, func(t *testing.T) {
			var exists bool
			err := srv.DB.QueryRow(`
				SELECT EXISTS (
					SELECT FROM information_schema.tables
					WHERE table_name = $1
				)
			`, table).Scan(&exists)
			if err != nil {
				t.Fatalf("query failed for table %s: %v", table, err)
			}
			if !exists {
				t.Errorf("table %s does not exist after migration", table)
			}
		})
	}
}

// TestDatabaseSeeding verifies that seed data was inserted.
func TestDatabaseSeeding(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	srv := setupTestServer(t)
	defer srv.Clean()

	ctx := context.Background()

	// Check employees were seeded
	count, err := srv.Client.Employee.Query().Count(ctx)
	if err != nil {
		t.Fatalf("failed to count employees: %v", err)
	}
	if count == 0 {
		t.Error("expected seeded employees, got 0")
	}

	// Check organizations were seeded
	orgCount, err := srv.Client.Organization.Query().Count(ctx)
	if err != nil {
		t.Fatalf("failed to count organizations: %v", err)
	}
	if orgCount == 0 {
		t.Error("expected seeded organizations, got 0")
	}

	// Check pillars were seeded
	pillarCount, err := srv.Client.Pillar.Query().Count(ctx)
	if err != nil {
		t.Fatalf("failed to count pillars: %v", err)
	}
	if pillarCount == 0 {
		t.Error("expected seeded pillars, got 0")
	}

	// Check cycles were seeded
	cycleCount, err := srv.Client.Cycle.Query().Count(ctx)
	if err != nil {
		t.Fatalf("failed to count cycles: %v", err)
	}
	if cycleCount == 0 {
		t.Error("expected seeded cycles, got 0")
	}
}
