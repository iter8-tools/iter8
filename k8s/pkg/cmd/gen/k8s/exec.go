package k8s

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	basecli "github.com/iter8-tools/iter8/cmd"

	"github.com/Masterminds/sprig"
	"github.com/spf13/cobra"

	// "helm.sh/helm/v3/pkg/chartutil"
	"sigs.k8s.io/yaml"
)

// complete sets all information needed for processing the command
func (o *Options) complete(cmd *cobra.Command, args []string) (err error) {
	return err
}

// validate ensures that all required arguments and flag values are provided
func (o *Options) validate(cmd *cobra.Command, args []string) (err error) {
	return nil
}

// run runs the command
func (o *Options) run(cmd *cobra.Command, args []string) (err error) {
	// v := chartutil.Values{}
	// err = basecli.ParseValues(values, v)
	// if err != nil {
	// 	return err
	// }

	exp, err := basecli.Build(false, &basecli.FileExpIO{})
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment file")
		return err
	}

	// generate formatted output

	b, err := RenderTpl(exp.Tasks, k8sTemplateFilePath)
	if err != nil {
		return err
	}
	fmt.Println(b.String())
	return nil
}

//go:embed k8s.tpl
var tplBytes []byte

// RenderTpl creates output from go.tpl
func RenderTpl(tasks []base.TaskSpec, filePath string) (*bytes.Buffer, error) {
	var tmpl *template.Template
	var err error

	// already read in the template file via go:embed above

	// add toYAML and other sprig template functions
	// they are all allowed to be used within the custom template
	// ensure it is a valid template
	tmpl, err = template.New("tpl").Funcs(template.FuncMap{
		"toYAML": toYAML,
	}).Option("missingkey=error").Funcs(sprig.TxtFuncMap()).Parse(string(tplBytes))
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to parse template file")
		return nil, err
	}

	// execute template
	var b bytes.Buffer
	err = tmpl.Execute(&b, tasks)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to execute template")
		return nil, err
	}

	// print output
	return &b, nil
}

// toYAML takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}
