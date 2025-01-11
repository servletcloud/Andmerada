package linters

type Reporter interface {
	AddError(title string, files ...string)
	AddWarning(title string, files ...string)
}

func NewDupeLinter(reporter Reporter) DupeLinter {
	return DupeLinter{
		Reporter:  reporter,
		idToNames: make(map[uint64][]string),
	}
}
