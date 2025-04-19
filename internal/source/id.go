package source

import (
	"fmt"
	"strconv"
	"time"
	"unicode"
)

const (
	IDFormatTimeYYYYMMDDHHMMSS = "20060102150405"
	EmptyMigrationID           = ID(0)
	MaxMigrationID             = ID(99991231235959)
	MinMigrationID             = ID(0)

	idLength = 14
)

type ID uint64

func NewIDFromString(str string) ID {
	if len(str) < idLength {
		return EmptyMigrationID
	}

	for _, ch := range str[:idLength] {
		if !unicode.IsDigit(ch) {
			return EmptyMigrationID
		}
	}

	return newIDFromStringUnsafe(str[:14])
}

func NewIDFromNow() ID {
	return NewIDFromTime(time.Now().UTC())
}

func NewIDFromTime(t time.Time) ID {
	timestamp := t.Format(IDFormatTimeYYYYMMDDHHMMSS)

	return newIDFromStringUnsafe(timestamp)
}

func (id ID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

func (id ID) Time() (time.Time, error) {
	timestamp := id.String()

	result, err := time.Parse(IDFormatTimeYYYYMMDDHHMMSS, timestamp)

	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse ID %v as time: %w", id, err)
	}

	return result, nil
}

func (id ID) Uint64() uint64 {
	return uint64(id)
}

func newIDFromStringUnsafe(s string) ID {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(err)
	}

	return ID(id)
}
