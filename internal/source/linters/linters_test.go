package linters_test

type TestLintReport struct {
	errors   []string
	warnings []string
}

func (report *TestLintReport) AddError(title string, _ ...string) {
	report.errors = append(report.errors, title)
}

func (report *TestLintReport) AddWarning(title string, _ ...string) {
	report.warnings = append(report.warnings, title)
}
