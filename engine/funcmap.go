package engine

import (
	"text/template"
)

var FuncMap template.FuncMap = template.FuncMap{
	"toYAML": toYAML,
}
