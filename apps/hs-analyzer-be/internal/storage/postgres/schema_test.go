package postgres

import (
	"strings"
	"testing"
)

func TestPostgresMigrationsContainCanonicalMatchTables(t *testing.T) {
	t.Parallel()

	migrations := strings.Join(Migrations(), "\n")
	for _, want := range []string{"CREATE TABLE IF NOT EXISTS matches", "CREATE TABLE IF NOT EXISTS match_events", "idempotency_key", "payload JSONB"} {
		if !strings.Contains(migrations, want) {
			t.Fatalf("migration does not contain %q:\n%s", want, migrations)
		}
	}
}
