package migrator

type PostgresConnectError struct {
	cause error
}

func (e *PostgresConnectError) Error() string {
	return e.cause.Error()
}

func (e *PostgresConnectError) Unwrap() error {
	return e.cause
}

type CreateDDLFailedError struct {
	cause error
	SQL   string
}

func (e *CreateDDLFailedError) Error() string {
	return e.cause.Error()
}

func (e *CreateDDLFailedError) Unwrap() error {
	return e.cause
}
