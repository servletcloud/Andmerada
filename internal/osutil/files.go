package osutil

import (
	"fmt"
	"log"
	"os"
)

const (
	//nolint:stylecheck,revive
	O_CREATE_EXCL_WRONLY = os.O_CREATE | os.O_EXCL | os.O_WRONLY

	FilePerm0644 = 0644 // Owner: read/write, Group/Others: read
	DirPerm0755  = 0755 // Owner: read/write/execute, Group/Others: read/execute
)

func WriteFile(path string, content string, flag int, perm os.FileMode) error {
	file, err := os.OpenFile(path, flag, perm)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Can not close file %s %v", path, err)
		}
	}()

	if _, err = file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write string to file %s: %w", path, err)
	}

	return nil
}
