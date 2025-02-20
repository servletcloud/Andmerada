package migrator

type ErrType int

const (
	ErrTypeDBConnect ErrType = iota
	ErrTypeCreateDDL
	ErrTypeListMigrationsOnDisk
	ErrTypeScanAppliedMigrations
)

func wrapError(err error, errType ErrType) error {
	return &MigrateError{Cause: err, ErrType: errType, SQL: ""}
}

func wrapErrorWithSQL(err error, errType ErrType, sql string) error {
	return &MigrateError{Cause: err, ErrType: errType, SQL: sql}
}

type MigrateError struct {
	Cause   error
	ErrType ErrType
	SQL     string
}

func (e *MigrateError) Error() string {
	return e.Cause.Error()
}

func (e *MigrateError) Unwrap() error {
	return e.Cause
}
