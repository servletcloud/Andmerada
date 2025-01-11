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

func (id MigrationID) asUint64() uint64 {
	return uint64(id)
}

func newIDFromTime(t time.Time) MigrationID {
	timestamp := t.Format(idFormatTimeYYYYMMDDHHMMSS)

	return newIDFromStringUnsafe(timestamp)
}

func newIDFromString(str string) MigrationID {
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

func newIDFromStringUnsafe(s string) MigrationID {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(err)
	}

	return MigrationID(id)
}
