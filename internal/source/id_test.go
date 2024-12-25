package source_test

import (
	"testing"
	"time"

	"github.com/servletcloud/Andmerada/internal/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIDFromTime(t *testing.T) {
	t.Parallel()

	timestamp, err := time.Parse("20060102150405", "20241225112129")
	require.NoError(t, err)

	assert.Equal(t, source.MigrationID(20241225112129), source.NewIDFromTime(timestamp))
}

func TestNewIDFromString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, source.MigrationID(20060102150405), source.NewIDFromString("20060102150405_create_users"))
	assert.Equal(t, source.EmptyMigrationID, source.NewIDFromString("2006010215040_create_users"))
	assert.Equal(t, source.EmptyMigrationID, source.NewIDFromString("200601021504056_create_users"))
	assert.Equal(t, source.EmptyMigrationID, source.NewIDFromString("20060102150405create_users"))
}
