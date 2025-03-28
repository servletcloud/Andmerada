package source

import (
	"fmt"
	"os"
)

func scanAll(projectDir string, callback func(id uint64, name string)) error {
	return scan(projectDir, func(id uint64, name string) bool {
		callback(id, name)

		return true
	})
}

func scan(projectDir string, callback func(id uint64, name string) bool) error {
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

		if id == EmptyMigrationID {
			continue
		}

		if !callback(id, name) {
			break
		}
	}

	return nil
}
