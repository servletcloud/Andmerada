package source

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

type FileTooBigError struct {
	Name  string
	Size  int64
	Limit int64
}

func (e *FileTooBigError) Error() string {
	return fmt.Sprintf("File %q is too big: %v exceeds the limit of %v",
		e.Name,
		humanize.Bytes(uint64(e.Size)),  //nolint:gosec
		humanize.Bytes(uint64(e.Limit)), //nolint:gosec
	)
}

type CannotBeDirError struct {
	Name string
}

func (e *CannotBeDirError) Error() string {
	return fmt.Sprintf("File %q must be a file, but is a directory", e.Name)
}

type DuplicateSourceError struct {
	Dir   string
	Paths []string
}

func (e *DuplicateSourceError) Error() string {
	return fmt.Sprintf("duplicate migrations: %v", e.Paths)
}

type CompileFilterError struct {
	Expression string
	Err        error
}

func (e *CompileFilterError) Error() string {
	return fmt.Sprintf("cannot compile filter expression %q: %v", e.Expression, e.Err)
}

func (e *CompileFilterError) Unwrap() error {
	return e.Err
}

type RunFilterError struct {
	Expression string
	ID         ID
	Err        error
}

func (e *RunFilterError) Error() string {
	return fmt.Sprintf("cannot run filter expression %q for ID %v: %v", e.Expression, e.ID, e.Err)
}

func (e *RunFilterError) Unwrap() error {
	return e.Err
}
