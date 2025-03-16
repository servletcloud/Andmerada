package tests

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func AssertPgTableExist(ctx context.Context, t *testing.T, conn *pgx.Conn, tableName string) {
	t.Helper()

	assert.True(t, isPgTableExist(ctx, t, conn, tableName), "Table %s does not exist", tableName)
}

func AssertPgTableNotExist(ctx context.Context, t *testing.T, conn *pgx.Conn, tableName string) {
	t.Helper()

	assert.False(t, isPgTableExist(ctx, t, conn, tableName), "Table %s exists", tableName)
}

func AssertFileContains(t *testing.T, path string, expected string) {
	t.Helper()

	content := ReadFileToString(t, path)

	assert.Contains(t, content, expected)
}

func AssertPlaceholdersResolved(t *testing.T, content string) {
	t.Helper()

	assert.NotContains(t, content, "{{")
	assert.NotContains(t, content, "}}")
}
