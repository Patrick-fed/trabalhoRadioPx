package db_test

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/radiopx/backend/internal/db"
)

// TestAudioNeverPersisted (RNF01 / TC-RNF01-01): the schema must contain only
// the four expected tables (users, channels, user_channels, schema_migrations)
// and no column able to hold audio - bytea/oid types or audio-ish names. Any
// unexpected table is a privacy audit failure: audio must live only in memory
// (RF06 buffer) and never on disk. Opt-in via RADIOPX_TEST_DB=1.
func TestAudioNeverPersisted(t *testing.T) {
	if os.Getenv("RADIOPX_TEST_DB") == "" {
		t.Skip("set RADIOPX_TEST_DB=1 to run PostgreSQL integration tests")
	}
	loadDotEnv(t)

	database, err := db.Open(db.ConfigFromEnv())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer database.Close()

	if err := db.ApplyMigrations(database); err != nil {
		t.Fatalf("migrations: %v", err)
	}

	// --- tabelas: apenas as 4 esperadas ---
	rows, err := database.Query(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name`)
	if err != nil {
		t.Fatalf("list tables: %v", err)
	}
	defer rows.Close()

	want := []string{"channels", "schema_migrations", "user_channels", "users"}
	got := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan table: %v", err)
		}
		got = append(got, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate tables: %v", err)
	}

	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("unexpected table set:\n got: %v\nwant: %v", got, want)
	}

	// --- colunas: nenhum tipo/b nome capaz de guardar audio ---
	forbiddenName := regexp.MustCompile(`audio|sound|voice|blob|record|media|payload`)
	forbiddenType := map[string]bool{"bytea": true, "oid": true}

	cols, err := database.Query(`
		SELECT table_name, column_name, data_type
		FROM information_schema.columns
		WHERE table_schema = 'public'
		ORDER BY table_name, column_name`)
	if err != nil {
		t.Fatalf("list columns: %v", err)
	}
	defer cols.Close()

	checked := 0
	for cols.Next() {
		var tableName, columnName, dataType string
		if err := cols.Scan(&tableName, &columnName, &dataType); err != nil {
			t.Fatalf("scan column: %v", err)
		}
		checked++
		lower := strings.ToLower(columnName)
		if forbiddenName.MatchString(lower) {
			t.Errorf("audio-like column %s.%s (%s) must not exist", tableName, columnName, dataType)
		}
		if forbiddenType[strings.ToLower(dataType)] {
			t.Errorf("binary column %s.%s (%s) must not exist", tableName, columnName, dataType)
		}
	}
	if err := cols.Err(); err != nil {
		t.Fatalf("iterate columns: %v", err)
	}
	if checked == 0 {
		t.Fatal("no columns found - schema missing?")
	}
}
