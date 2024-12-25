package source

import (
	"strconv"
	"time"
	"unicode"
)

type MigrationID uint64

const (
	EmptyMigrationID = MigrationID(0)

	idLength                 = 14
	formatTimeYYYYMMDDHHMMSS = "20060102150405"
)

func NewIDFromTime(t time.Time) MigrationID {
	timestamp := t.Format(formatTimeYYYYMMDDHHMMSS)

	return fromString(timestamp)
}

func NewIDFromString(str string) MigrationID {
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

	return fromString(str[:14])
}

func fromString(s string) MigrationID {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(err)
	}

	return MigrationID(id)
}
