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
