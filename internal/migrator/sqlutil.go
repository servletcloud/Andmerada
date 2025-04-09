package migrator

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func execSimple(ctx context.Context, conn *pgconn.PgConn, sql string) error {
	mrr := conn.Exec(ctx, sql)

	for mrr.NextResult() {
		_, _ = mrr.ResultReader().Close()
	}

	return mrr.Close() //nolint:wrapcheck
}

func isConnectionInTransaction(conn *pgconn.PgConn) bool {
	const (
		inTransaction       = 'T'
		inFailedTransaction = 'E'
	)

	status := conn.TxStatus()

	return status == inTransaction || status == inFailedTransaction
}

func isPgErrorOfCode(err error, pgErrorCode string) bool {
	var pgError *pgconn.PgError

	return errors.As(err, &pgError) && pgError.Code == pgErrorCode
}
