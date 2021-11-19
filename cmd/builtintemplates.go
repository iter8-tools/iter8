package cmd

import (
	"os"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/iter8-tools/iter8/base/log"
)

// built-in template that can be used for reporting experiment results
var builtInTemplates = make(map[string]*template.Template)

// Register template adds a template to builtInTemplates
func RegisterTemplate(name string, tpl *template.Template) {
	builtInTemplates[name] = tpl
}

func init() {

	// create text template
	tmpl, err := template.New("text").Funcs(template.FuncMap{
		"formatText": formatText,
	}).Option("missingkey=error").Funcs(sprig.TxtFuncMap()).Parse("{{ formatText . }}")
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to parse text template")
		os.Exit(1)
	}

	// register text template
	RegisterTemplate("text", tmpl)
}
