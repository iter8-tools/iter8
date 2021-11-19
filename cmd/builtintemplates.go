package cmd

import (
	"text/template"
)

// built-in template that can be used for reporting experiment results
var builtInTemplates = make(map[string]*template.Template)

// Register template adds a template to builtInTemplates
func RegisterTemplate(name string, tpl *template.Template) {
	builtInTemplates[name] = tpl
}
