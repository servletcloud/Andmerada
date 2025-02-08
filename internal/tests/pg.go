package tests

import (
	"context"
	"fmt"
	"net"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ConnectionURL string

func StartEmbeddedPostgres(t *testing.T) ConnectionURL {
	t.Helper()

	t.Log("Downloading and starting an embedded Postgres database...")

	port := GetRandomAvailablePort(t)
	config := embeddedpostgres.DefaultConfig().DataPath(t.TempDir()).Port(port)

	database := embeddedpostgres.NewDatabase(config)
	require.NoError(t, database.Start())

	t.Cleanup(func() {
		assert.NoError(t, database.Stop())
	})

	return ConnectionURL(config.GetConnectionURL())
}

func OpenPgConnection(t *testing.T, url ConnectionURL) *pgx.Conn {
	t.Helper()

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, string(url))
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, conn.Close(ctx))
	})

	require.NoError(t, conn.Ping(ctx))

	return conn
}

func GetRandomAvailablePort(t *testing.T) uint32 {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)

	return uint32(addr.Port) //nolint:gosec
}

func IsPgTableExist(ctx context.Context, t *testing.T, conn *pgx.Conn, name string) bool {
	t.Helper()

	query := fmt.Sprintf("select * from %v limit 1;", name)

	_, err := conn.Exec(ctx, query)

	if err == nil {
		return true
	}

	var pgError *pgconn.PgError

	require.ErrorAs(t, err, &pgError)

	if pgError.Code != pgerrcode.UndefinedTable {
		require.NoError(t, err)
	}

	return false
}
