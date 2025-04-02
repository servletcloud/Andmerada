package migrator

import (
	"errors"
	"strings"
)

type ErrType int

const (
	ErrTypeDBConnect ErrType = iota
	ErrTypeCreateDDL
	ErrTypeListMigrationsOnDisk
	ErrTypeScanAppliedMigrations
	ErrTypeApplyMigration
	ErrTypeRegisterMigration
)

func wrapError(err error, errType ErrType) error {
	var migrateError *MigrateError

	if errors.As(err, &migrateError) {
		return migrateError
	}

	return &MigrateError{Cause: err, ErrType: errType}
}

type MigrateError struct {
	Cause   error
	ErrType ErrType
}

func (e *MigrateError) Error() string {
	return e.Cause.Error()
}

func (e *MigrateError) Unwrap() error {
	return e.Cause
}

type ExecSQLError struct {
	Cause error
	SQL   string
}

func (e *ExecSQLError) Error() string {
	return e.Cause.Error()
}

func (e *ExecSQLError) Unwrap() error {
	return e.Cause
}

type ApplyMigrationError struct {
	Cause error
	Name  string
}

func (e *ApplyMigrationError) Error() string {
	return e.Cause.Error()
}

func (e *ApplyMigrationError) Unwrap() error {
	return e.Cause
}

type TransactionNotCommittedError struct {
	RollBackError error
}

func (e *TransactionNotCommittedError) Error() string {
	var sb strings.Builder

	sb.WriteString("The migration SQL opened a transaction, but it was neither committed nor rolled back.\n")
	sb.WriteString("This may leave the database in an inconsistent state.\n")
	sb.WriteString("Please review your migration script to ensure it explicitly commits or rolls back transactions.\n")

	if e.RollBackError == nil {
		sb.WriteString("Rollback was executed successfully.")
	} else {
		sb.WriteString("Rollback failed with the following error: ")
		sb.WriteString(e.RollBackError.Error())
		sb.WriteString("\nManual intervention may be required.")
	}

	return sb.String()
}
