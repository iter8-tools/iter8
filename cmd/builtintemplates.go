package cmd

import (
	"text/template"

	"github.com/Masterminds/sprig"
)

// built-in template that can be used for reporting experiment results
var builtInTemplates map[string]*template.Template

func init() {
	builtInTemplates = map[string]*template.Template{
		"text": template.New("text").Funcs(template.FuncMap{
			"formatText": formatText,
		}).Option("missingkey=error").Funcs(sprig.TxtFuncMap()),
	}
}
