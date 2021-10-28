package engine

import (
	"text/template"

	"github.com/iter8-tools/iter8/core"
)

var FuncMap template.FuncMap = template.FuncMap{
	"include": func(*core.ExperimentSpec) string { return "not implemented" },
}
