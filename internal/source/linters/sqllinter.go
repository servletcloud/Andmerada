package linters

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/dustin/go-humanize"
)

type SQLLinter struct {
	Reporter
	ProjectDir          string
	MaxSQLFileSize      int64
	CreatedFromTemplate []byte
	ErrEmptyMsg         string
	ErrUntouchedMsg     string
}

func (linter *SQLLinter) Lint(relative string) {
	stat, err := os.Stat(linter.absolutePath(relative))

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			linter.AddError("File referenced by migration.yml does not exist", relative)
		} else {
			message := fmt.Sprint("File referenced by migration.yml cannot be read:\n", err.Error())
			linter.AddError(message, relative)
		}

		return
	}

	if stat.IsDir() {
		linter.AddError("Must be a file but is a directory", relative)

		return
	}

	size := stat.Size()
	linter.lintFileSize(relative, size)
	linter.lintUntouched(relative, size)
}

func (linter *SQLLinter) lintFileSize(relative string, size int64) {
	if size > linter.MaxSQLFileSize {
		title := fmt.Sprintf("File is too big: %v exceeds the limit of %v",
			humanize.Bytes(uint64(size)),                  //nolint:gosec
			humanize.Bytes(uint64(linter.MaxSQLFileSize)), //nolint:gosec
		)
		linter.AddError(title, relative)
	}

	if size == 0 {
		linter.AddWarning(linter.ErrEmptyMsg, relative)
	}
}

func (linter *SQLLinter) lintUntouched(relative string, size int64) {
	templateSize := int64(len(linter.CreatedFromTemplate))

	if size != templateSize {
		return
	}

	content, err := os.ReadFile(linter.absolutePath(relative))
	if err != nil {
		linter.AddError("Unable to read file", relative)
	}

	if slices.Equal(content, linter.CreatedFromTemplate) {
		linter.AddWarning(linter.ErrUntouchedMsg, relative)
	}
}

func (linter *SQLLinter) absolutePath(relative string) string {
	return filepath.Join(linter.ProjectDir, relative)
}
