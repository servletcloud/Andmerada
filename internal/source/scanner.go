package source

import (
	"fmt"
	"os"
)

func ScanAll(projectDir string) (map[ID]string, error) {
	idToName := make(map[ID]string)

	var duplicates []string

	err := Traverse(projectDir, func(id ID, name string) bool {
		if _, found := idToName[id]; found {
			duplicates = []string{idToName[id], name}

			return false
		}

		idToName[id] = name

		return true
	})

	if len(duplicates) > 0 {
		return nil, &DuplicateSourceError{Dir: projectDir, Paths: duplicates}
	}

	return idToName, err
}

func TraverseAll(projectDir string, callback func(id ID, name string)) error {
	return Traverse(projectDir, func(id ID, name string) bool {
		callback(id, name)

		return true
	})
}

func Traverse(projectDir string, callback func(id ID, name string) bool) error {
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
