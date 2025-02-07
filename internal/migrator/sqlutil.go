package migrator

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
)

func execSimple(ctx context.Context, conn *pgconn.PgConn, sql string) error {
	mrr := conn.Exec(ctx, sql)

	for mrr.NextResult() {
		_, _ = mrr.ResultReader().Close()
	}

	return mrr.Close() //nolint:wrapcheck
}
