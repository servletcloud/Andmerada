package cmd

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

type pgErrorTranslator struct {
}

func (translator *pgErrorTranslator) prettyPrint(err *pgconn.PgError, sql string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("PostgreSQL Error: %s\n", err.Message))

	translator.writeKv(&sb, "Severity", err.Severity)

	if err.Code != "" {
		helpURL := "https://www.postgresql.org/docs/current/errcodes-appendix.html#ERRCODES-TABLE"
		sb.WriteString(fmt.Sprintf("Code: %v (more details at %v)\n", err.Code, helpURL))
	}

	translator.writeKv(&sb, "Where", err.Where)
	translator.writeKv(&sb, "Hint", err.Hint)
	translator.writeKv(&sb, "Detail", err.Detail)

	if err.Position > 0 && int(err.Position) <= len(sql) && sql != "" {
		lineNumber, colNumber, highlightedSQL := translator.highlightSQLPosition(sql, int(err.Position))
		sb.WriteString(fmt.Sprintf("\nError Location at line %d, column %d:\n", lineNumber, colNumber))
		sb.WriteString(highlightedSQL)
	}

	translator.writeKv(&sb, "Schema", err.SchemaName)
	translator.writeKv(&sb, "Table", err.TableName)
	translator.writeKv(&sb, "Column", err.ColumnName)
	translator.writeKv(&sb, "Constraint", err.ConstraintName)

	if err.InternalQuery != "" {
		sb.WriteString("\n**Internal Query Debugging:**\n")
		sb.WriteString(fmt.Sprintf("Position: %v\n", err.InternalPosition))
		sb.WriteString(err.InternalQuery + "\n")
	}

	return sb.String()
}

func (translator *pgErrorTranslator) writeKv(sb *strings.Builder, key, value string) {
	if value == "" {
		return
	}

	sb.WriteString(fmt.Sprintf("%s: %s\n", key, value))
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
