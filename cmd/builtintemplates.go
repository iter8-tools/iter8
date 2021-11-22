package cmd

import (
	ht "html/template"
	"text/template"
)

// built-in text templates that can be used for reporting experiment results
var builtInTextTemplates = make(map[string]*template.Template)

// built-in HTML templates that can be used for reporting experiment results
var builtInHTMLTemplates = make(map[string]*ht.Template)

// RegisterTextTemplate registers a text template
func RegisterTextTemplate(name string, tpl *template.Template) {
	builtInTextTemplates[name] = tpl
}

// RegisterHTMLTemplate registers an HTML template
func RegisterHTMLTemplate(name string, tpl *ht.Template) {
	builtInHTMLTemplates[name] = tpl
}
