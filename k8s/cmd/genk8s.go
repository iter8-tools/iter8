package cmd

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
	"helm.sh/helm/v3/pkg/chartutil"
	"sigs.k8s.io/yaml"
)

const (
	// Path to go template file
	k8sTemplateFilePath = "k8s.tpl"
	experimentFilePath  = "experiment.yaml"
)

type GetK8sOptions struct {
}

type k8sExperiment struct {
	Tasks  []base.TaskSpec
	Values chartutil.Values
}

// run runs the command
func (o *GetK8sOptions) run(cmd *cobra.Command, args []string) (err error) {
	v := chartutil.Values{}
	err = basecli.ParseValues(basecli.GenOptions.Values, v)
	if err != nil {
		return err
	}

	exp, err := basecli.Build(false, &basecli.FileExpIO{})
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment file")
		return err
	}
	k8sExp := k8sExperiment{
		Tasks:  exp.Tasks,
		Values: v,
	}

	// generate formatted output

	b, err := RenderTpl(k8sExp, k8sTemplateFilePath)
	if err != nil {
		return err
	}
	fmt.Println(b.String())
	return nil
}

//go:embed k8s.tpl
var tplBytes []byte

// RenderTpl creates output from go.tpl
func RenderTpl(k8sExp k8sExperiment, filePath string) (*bytes.Buffer, error) {
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
	err = tmpl.Execute(&b, k8sExp)
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

func NewGetK8sCmd() *cobra.Command {
	o := &GetK8sOptions{}

	cmd := &cobra.Command{
		Use:   "k8s",
		Short: "Generate manifest for running experiment in Kubernetes",
		Example: `
# Generate Kubernetes manifest
iter8 gen k8s`,
		SilenceUsage: true,
		// Put any optionm computation and/or validatiom here
		// PreRunE: func(c *cobra.Command, args []string) error {
		RunE: func(c *cobra.Command, args []string) error {
			return o.run(c, args)
		},
	}

	return cmd
}
