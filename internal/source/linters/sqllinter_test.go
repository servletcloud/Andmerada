package linters_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/servletcloud/Andmerada/internal/osutil"
	"github.com/servletcloud/Andmerada/internal/source/linters"
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

		report := &TestLintReport{}
		linter := &linters.SQLLinter{Reporter: report, ProjectDir: dir, MaxSQLFileSize: sqlContentSize}

		linter.Lint("up.sql")

		assert.Empty(t, report.errors)
		assert.Empty(t, report.warnings)
	})

	//nolint:exhaustruct
	t.Run("No file exist", func(t *testing.T) {
		t.Parallel()

		report := &TestLintReport{}
		linter := &linters.SQLLinter{Reporter: report, ProjectDir: dir}

		linter.Lint("this-file-does-not-exist.sql")

		assert.Contains(t, report.errors, "File referenced by migration.yml does not exist")
		assert.Empty(t, report.warnings)
	})

	//nolint:exhaustruct
	t.Run("File is a directory", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, os.Mkdir(filepath.Join(dir, "sql-up-sub-directory"), osutil.DirPerm0755))

		report := &TestLintReport{}
		linter := &linters.SQLLinter{Reporter: report, ProjectDir: dir}

		linter.Lint("sql-up-sub-directory")

		assert.Contains(t, report.errors, "Must be a file but is a directory")
		assert.Empty(t, report.warnings)
	})

	//nolint:exhaustruct
	t.Run("File is too big", func(t *testing.T) {
		t.Parallel()

		maxSize := sqlContentSize - 1

		report := &TestLintReport{}
		linter := &linters.SQLLinter{Reporter: report, ProjectDir: dir, MaxSQLFileSize: maxSize}

		linter.Lint("up.sql")

		expectedError := fmt.Sprintf("File is too big: %v B exceeds the limit of %v B", sqlContentSize, maxSize)
		assert.Contains(t, report.errors, expectedError)
		assert.Empty(t, report.warnings)
	})

	//nolint:exhaustruct
	t.Run("File is empty", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(dir, "empty-up.sql")
		require.NoError(t, osutil.WriteFileExcl(path, ""))

		report := &TestLintReport{}
		linter := &linters.SQLLinter{Reporter: report, ProjectDir: dir, ErrEmptyMsg: "File is empty"}

		linter.Lint("empty-up.sql")

		assert.Empty(t, report.errors)
		assert.Contains(t, report.warnings, "File is empty")
	})

	//nolint:exhaustruct
	t.Run("File equals its template", func(t *testing.T) {
		t.Parallel()

		report := &TestLintReport{}
		linter := &linters.SQLLinter{
			Reporter:            report,
			ProjectDir:          dir,
			MaxSQLFileSize:      sqlContentSize,
			CreatedFromTemplate: []byte(sqlContent),
			ErrUntouchedMsg:     "File is untouched",
		}

		linter.Lint("up.sql")

		assert.Empty(t, report.errors)
		assert.Contains(t, report.warnings, "File is untouched")
	})
}
