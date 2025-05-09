package migrator_test

import (
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/servletcloud/Andmerada/internal/migrator"
	"github.com/servletcloud/Andmerada/internal/migrator/sqlres"
	"github.com/servletcloud/Andmerada/internal/source"
	"github.com/servletcloud/Andmerada/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest
func TestMigrations_ScanApplied(t *testing.T) {
	connectionURL := tests.StartEmbeddedPostgres(t)
	conn := tests.OpenPgConnection(t, connectionURL)

	_, err := conn.Exec(t.Context(), sqlres.DDL("migrations"))
	require.NoError(t, err)

	migrations := &migrator.Migrations{TableName: "migrations"}

	scanAppliedMigrations := func(t *testing.T, minID, maxID source.ID) []source.ID {
		t.Helper()

		result, err := migrations.ScanApplied(t.Context(), conn, minID, maxID)

		require.NoError(t, err)

		return result
	}

	t.Run("empty table", func(t *testing.T) {
		actual := scanAppliedMigrations(t, 0, 99991225112129)
		require.Empty(t, actual)
	})

	t.Run("when there are applied migrations", func(t *testing.T) {
		insertDummyMigration(t, conn, 20241225112129)
		insertDummyMigration(t, conn, 20241225112130)
		insertDummyMigration(t, conn, 20241225112131)

		t.Run("filter covers the boundaries", func(t *testing.T) {
			actual := scanAppliedMigrations(t, 20241225112129, 20241225112131)

			assert.Len(t, actual, 3)
			assert.Contains(t, actual, source.ID(20241225112129))
			assert.Contains(t, actual, source.ID(20241225112130))
			assert.Contains(t, actual, source.ID(20241225112131))
		})

		t.Run("filter includes the boundary values", func(t *testing.T) {
			actual := scanAppliedMigrations(t, 20241225112130, 20241225112130)

			assert.Len(t, actual, 1)
			assert.Contains(t, actual, source.ID(20241225112130))
		})
	})
}

func insertDummyMigration(t *testing.T, conn *pgx.Conn, id uint64) {
	t.Helper()

	name := "create users table"
	sqlUp := "create table users (id bigint primary key);"
	sqlUpSHA256 := "9473f4cfe827e5c29acffc4c80b8194aa3df919577fbf2f6b11df3d0f14cd907"
	durationMS := 10
	meta := make(map[string]struct{})

	query := `
		INSERT INTO migrations (id, name, sql_up, sql_up_sha256, duration_ms, meta)
		VALUES ($1, $2, $3, $4, $5, $6);
	`

	_, err := conn.Exec(t.Context(), query, id, name, sqlUp, sqlUpSHA256, durationMS, meta)
	require.NoError(t, err)
}
