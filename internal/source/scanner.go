package source

import (
	"fmt"
	"os"
)

func Scan(projectDir string, callback func(id MigrationID, name string) bool) error {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return fmt.Errorf("cannot read directory %v because: %w", projectDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		id := NewIDFromString(name)

		if id != EmptyMigrationID {
			if !callback(id, name) {
				break
			}
		}
	}

	return nil
}
