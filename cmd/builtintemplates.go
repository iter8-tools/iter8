package cmd

import (
	ht "html/template"
	"text/template"
)

// builtInTemplates that can be used for reporting experiment results
var builtInTemplates = make(map[string]executable)

// RegisterTextTemplate registers a text template
func RegisterTextTemplate(name string, tpl *template.Template) {
	builtInTemplates[name] = tpl
}

// RegisterHTMLTemplate registers an HTML template
func RegisterHTMLTemplate(name string, tpl *ht.Template) {
	builtInTemplates[name] = tpl
}
