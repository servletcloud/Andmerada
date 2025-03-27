package migrator

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/servletcloud/Andmerada/internal/migrator/sqlres"
)

type Migration struct {
	ID              uint64
	Name            string
	AppliedAt       time.Time
	SQLUp           string
	SQLDown         string
	SQLUpSHA256     string
	SQLDownSHA256   string
	DurationMs      int64
	RollbackBlocked bool
	Meta            map[string]any
}

type Migrations struct {
	TableName string
}

func (m *Migrations) ScanApplied(
	ctx context.Context,
	conn *pgx.Conn,
	minID, maxID uint64,
) ([]uint64, error) {
	queryTemplate := "SELECT id FROM %s WHERE id >= $1 AND id <= $2"
	query := fmt.Sprintf(queryTemplate, m.TableName)

	rows, err := conn.Query(ctx, query, minID, maxID)

	if err != nil {
		return nil, &ExecSQLError{Cause: err, SQL: query}
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[uint64])

	if err != nil {
		return nil, &ExecSQLError{Cause: err, SQL: query}
	}

	return ids, nil
}

func (m *Migrations) Insert(ctx context.Context, conn *pgx.Conn, migration *Migration) error {
	query := sqlres.RegisterMigrationQuery(m.TableName)

	args := pgx.NamedArgs{
		"id":               migration.ID,
		"name":             migration.Name,
		"applied_at":       migration.AppliedAt,
		"sql_up":           migration.SQLUp,
		"sql_down":         migration.SQLDown,
		"sql_up_sha256":    migration.SQLUpSHA256,
		"sql_down_sha256":  migration.SQLDownSHA256,
		"duration_ms":      migration.DurationMs,
		"rollback_blocked": migration.RollbackBlocked,
		"meta":             migration.Meta,
	}

	_, err := conn.Exec(ctx, query, args)

	if err != nil {
		return &ExecSQLError{Cause: err, SQL: query}
	}

	return nil
}
