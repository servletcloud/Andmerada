package osutil

import (
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"
)

const (
	//nolint:stylecheck,revive
	O_CREATE_EXCL_WRONLY = os.O_CREATE | os.O_EXCL | os.O_WRONLY

	FilePerm0644 = 0644 // Owner: read/write, Group/Others: read
	DirPerm0755  = 0755 // Owner: read/write/execute, Group/Others: read/execute

	filenameAllowedChars = "abcdefghijklmnopqrstuvwxyz0123456789._-"
)

func WriteFileExcl(path string, content string) error {
	file, err := os.OpenFile(path, O_CREATE_EXCL_WRONLY, FilePerm0644)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("cannot close file %s %v", path, err)
		}
	}()

	if _, err = file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write string to file %s: %w", path, err)
	}

	return nil
}

func GetwdOrPanic() string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Error getting current directory: %v\n", err))
	}

	return currentDir
}

func NormalizePath(name string) string {
	var result strings.Builder

	var lastChar rune

	for _, char := range name {
		char = unicode.ToLower(char)
		if strings.ContainsRune(filenameAllowedChars, char) {
			result.WriteRune(char)
			lastChar = char
		} else if lastChar != '_' {
			result.WriteRune('_')

			lastChar = '_'
		}
	}

	return strings.Trim(result.String(), "_-.")
}
