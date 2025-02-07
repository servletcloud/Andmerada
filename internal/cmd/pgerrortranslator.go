package cmd

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

type pgErrorTranslator struct {
	writeString func(message string)
}

func (translator *pgErrorTranslator) prettyPrint(err *pgconn.PgError, sql string) {
	translator.writeString(fmt.Sprintf("PostgreSQL Error: %s\n", err.Message))

	translator.writeKv("Severity", err.Severity)

	if err.Code != "" {
		helpURL := "https://www.postgresql.org/docs/current/errcodes-appendix.html#ERRCODES-TABLE"
		translator.writeString(fmt.Sprintf("Code: %v (more details at %v)\n", err.Code, helpURL))
	}

	translator.writeKv("Where", err.Where)
	translator.writeKv("Hint", err.Hint)
	translator.writeKv("Detail", err.Detail)

	if err.Position > 0 && int(err.Position) <= len(sql) && sql != "" {
		lineNumber, colNumber, highlightedSQL := translator.highlightSQLPosition(sql, int(err.Position))
		translator.writeString(fmt.Sprintf("\nError Location at line %d, column %d:\n", lineNumber, colNumber))
		translator.writeString(highlightedSQL)
	}

	translator.writeKv("Schema", err.SchemaName)
	translator.writeKv("Table", err.TableName)
	translator.writeKv("Column", err.ColumnName)
	translator.writeKv("Constraint", err.ConstraintName)

	if err.InternalQuery != "" {
		translator.writeString("\n**Internal Query Debugging:**\n")
		translator.writeString(fmt.Sprintf("Position: %v\n", err.InternalPosition))
		translator.writeString(err.InternalQuery + "\n")
	}
}

func (translator *pgErrorTranslator) writeKv(key, value string) {
	if value == "" {
		return
	}

	translator.writeString(fmt.Sprintf("%s: %s\n", key, value))
}

func (translator *pgErrorTranslator) highlightSQLPosition(sql string, pos int) (int, int, string) {
	lines := strings.Split(sql, "\n")
	cursor := 0

	for i, line := range lines {
		if cursor+len(line)+1 >= pos {
			lineNumber := i + 1
			columnNumber := pos - cursor
			highlighted := line + "\n" + strings.Repeat(" ", columnNumber-1) + "^" + "\n"

			return lineNumber, columnNumber, highlighted
		}

		cursor += len(line) + 1 // +1 for newline
	}

	return 0, 0, sql
}
