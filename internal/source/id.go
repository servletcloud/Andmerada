package source

import (
	"strconv"
	"time"
	"unicode"
)

const (
	idLength                   = 14
	idFormatTimeYYYYMMDDHHMMSS = "20060102150405"
)

func newIDFromTime(t time.Time) uint64 {
	timestamp := t.Format(idFormatTimeYYYYMMDDHHMMSS)

	return newIDFromStringUnsafe(timestamp)
}

func newIDFromString(str string) uint64 {
	if len(str) < idLength+1 {
		return EmptyMigrationID
	}

	for _, ch := range str[:idLength] {
		if !unicode.IsDigit(ch) {
			return EmptyMigrationID
		}
	}

	if str[14] != '_' {
		return EmptyMigrationID
	}

	return newIDFromStringUnsafe(str[:14])
}

func newIDFromStringUnsafe(s string) uint64 {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(err)
	}

	return id
}
