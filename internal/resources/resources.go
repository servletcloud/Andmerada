package resources

import (
	_ "embed"
	"strings"
)

type CommandDescription struct {
	Use   string
	Short string
	Long  string
}

const (
	unixNewLine = "\n"
)

//go:embed template.andmerada.yml
var templateAndmeradaYml []byte

//go:embed command_init_description.txt
var commandInitDiscription []byte

//go:embed msg_init_completed.txt
var msgInitCompleted []byte

//go:embed msg_err_project_exists.txt
var msgErrProjectExists []byte

func TemplateAndmeradaYml() string {
	return string(templateAndmeradaYml)
}

func LoadInitCommandDescription() CommandDescription {
	return loadCommandDescription(commandInitDiscription)
}

func MsgInitCompleted() string {
	return string(msgInitCompleted)
}

func MsgErrProjectExists() string {
	return string(msgErrProjectExists)
}

func loadCommandDescription(s []byte) CommandDescription {
	lines := strings.Split(string(s), unixNewLine)

	return CommandDescription{
		Use:   lines[0],
		Short: lines[1],
		Long:  strings.Join(lines[2:], unixNewLine),
	}
}
