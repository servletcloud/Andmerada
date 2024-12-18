package resources

import (
	_ "embed"
)

//go:embed template.andmerada.yml
var templateAndmeradaYml []byte

func TemplateAndmeradaYml() string {
	return string(templateAndmeradaYml)
}
