package migrator

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/servletcloud/Andmerada/internal/migrator/sqlres"
	"github.com/servletcloud/Andmerada/internal/project"
	"github.com/servletcloud/Andmerada/internal/source"
)

type Report struct {
	SourcesOnDisk int
}

type Applier struct {
	DatabaseURL string
	Project     project.Project
}

func (applier *Applier) ApplyPending(ctx context.Context, report *Report) error {
	sourceIDToName := make(map[source.MigrationID]string)
	dupeIDToName := make(map[source.MigrationID]string)
	idMin := source.MigrationID(math.MaxUint64)
	idMax := source.MigrationID(0)

	err := source.ScanAll(applier.Project.Dir, func(id source.MigrationID, name string) {
		_, found := sourceIDToName[id]
		if found {
			dupeIDToName[id] = name
		} else {
			sourceIDToName[id] = name
		}

		idMax = max(idMax, id)
		idMin = min(idMin, id)
	})

	if err != nil {
		return fmt.Errorf("failed to scan migration files on disk: %w", err)
	}

	report.SourcesOnDisk = len(sourceIDToName)

	if len(sourceIDToName) == 0 {
		return nil
	}

	connection, err := pgx.Connect(ctx, applier.DatabaseURL)
	if err != nil {
		return &PostgresConnectError{cause: err}
	}

	defer func() { _ = connection.Close(ctx) }()

	ddl := sqlres.DDL(applier.Project.Configuration.TableNames.AppliedMigrations)

	if err := execSimple(ctx, connection.PgConn(), ddl); err != nil {
		return &CreateDDLFailedError{cause: err, SQL: ddl}
	}

	return nil
}

func (applier *Applier) ScanAppliedMigrations(
	ctx context.Context,
	conn *pgx.Conn,
	minID, maxID uint64,
	callback func(uint64) bool,
) error {
	queryTemplate := "COPY (SELECT id FROM %s WHERE id >= %d AND id <= %d) TO STDOUT"
	query := fmt.Sprintf(queryTemplate, applier.migrationsTableName(), minID, maxID)

	var buf bytes.Buffer
	_, err := conn.PgConn().CopyTo(ctx, &buf, query)

	if err != nil {
		return fmt.Errorf("failed to execute COPY query %q: %w", query, err)
	}

	scanner := bufio.NewScanner(strings.NewReader(buf.String()))

	var id uint64

	for scanner.Scan() {
		if _, err := fmt.Sscanf(scanner.Text(), "%d", &id); err != nil {
			return fmt.Errorf("failed to parse migration ID from %q: %w", scanner.Text(), err)
		}

		if !callback(id) {
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading COPY output: %w", err)
	}

	return nil
}

func (applier *Applier) migrationsTableName() string {
	return applier.Project.Configuration.TableNames.AppliedMigrations
}
