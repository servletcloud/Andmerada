package linter

type Report struct {
	Errors   []LintError
	Warnings []LintError
}

func (r *Report) AddError(title string, files ...string) {
	r.Errors = append(r.Errors, LintError{Title: title, Files: files})
}

func (r *Report) AddWarning(title string, files ...string) {
	r.Warnings = append(r.Warnings, LintError{Title: title, Files: files})
}
