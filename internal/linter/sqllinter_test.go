package linter_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/servletcloud/Andmerada/internal/linter"
	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqlLinter(t *testing.T) { //nolint:funlen
	t.Parallel()

	sqlContent := "DROP TABLE users;"
	sqlContentSize := int64(len([]byte(sqlContent)))
	dir := t.TempDir()
	path := filepath.Join(dir, "up.sql")
	require.NoError(t, osutil.WriteFileExcl(path, sqlContent))

	//nolint:exhaustruct
	t.Run("SQL File is valid", func(t *testing.T) {
		t.Parallel()

		report := linter.Report{}
		linter := &linter.SQLLinter{ProjectDir: dir, MaxSQLFileSize: sqlContentSize}

		linter.Lint(&report, "up.sql")

		assert.Empty(t, report.Errors)
		assert.Empty(t, report.Warnings)
	})

	//nolint:exhaustruct
	t.Run("No file exist", func(t *testing.T) {
		t.Parallel()

		report := linter.Report{}
		linter := &linter.SQLLinter{ProjectDir: dir}

		linter.Lint(&report, "this-file-does-not-exist.sql")

		assertContainsError(t, report.Errors, "File referenced by migration.yml does not exist")
		assert.Empty(t, report.Warnings)
	})

	//nolint:exhaustruct
	t.Run("File is a directory", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, os.Mkdir(filepath.Join(dir, "sql-up-sub-directory"), osutil.DirPerm0755))

		report := linter.Report{}
		linter := &linter.SQLLinter{ProjectDir: dir}

		linter.Lint(&report, "sql-up-sub-directory")

		assertContainsError(t, report.Errors, "Must be a file but is a directory")
		assert.Empty(t, report.Warnings)
	})

	//nolint:exhaustruct
	t.Run("File is too big", func(t *testing.T) {
		t.Parallel()

		maxSize := sqlContentSize - 1

		report := linter.Report{}
		linter := &linter.SQLLinter{ProjectDir: dir, MaxSQLFileSize: maxSize}

		linter.Lint(&report, "up.sql")

		expectedError := fmt.Sprintf("File is too big: %v B exceeds the limit of %v B", sqlContentSize, maxSize)
		assertContainsError(t, report.Errors, expectedError)
		assert.Empty(t, report.Warnings)
	})

	//nolint:exhaustruct
	t.Run("File is empty", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(dir, "empty-up.sql")
		require.NoError(t, osutil.WriteFileExcl(path, ""))

		report := linter.Report{}
		linter := &linter.SQLLinter{ProjectDir: dir, ErrEmptyMsg: "File is empty"}

		linter.Lint(&report, "empty-up.sql")

		assert.Empty(t, report.Errors)
		assertContainsError(t, report.Warnings, "File is empty")
	})

	//nolint:exhaustruct
	t.Run("File equals its template", func(t *testing.T) {
		t.Parallel()

		report := linter.Report{}
		linter := &linter.SQLLinter{
			ProjectDir:          dir,
			MaxSQLFileSize:      sqlContentSize,
			CreatedFromTemplate: []byte(sqlContent),
			ErrUntouchedMsg:     "File is untouched",
		}

		linter.Lint(&report, "up.sql")

		assert.Empty(t, report.Errors)
		assertContainsError(t, report.Warnings, "File is untouched")
	})
}
