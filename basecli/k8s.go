package basecli

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"

	"github.com/Masterminds/sprig"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chartutil"
)

const (
	// Path to go template file
	k8sTemplateFilePath = "k8s.tpl"
)

type k8sExperiment struct {
	Tasks  []base.TaskSpec
	Values chartutil.Values
}

var id string
var app string

// run runs the command
func runGetK8sCmd(cmd *cobra.Command, args []string) (err error) {
	result, err := Generate(GenOptions.Values)
	if err != nil {
		return err
	}
	if result != nil {
		fmt.Println(result.String())
	}
	return nil
}

func Generate(values []string) (result *bytes.Buffer, err error) {
	v := chartutil.Values{}

	// add id=id if --id option used
	// note that if both --id=foo and --set id=bar are used,
	// the --id option will take precedence
	if len(id) > 0 {
		values = append(values, "id="+id)
	}

	// add app=app if --app option is used
	// not that if both --app=foo and --set id=bar are used,
	// the --app optiom will take precedence
	if len(app) > 0 {
		values = append(values, "app="+app)
	}

	err = ParseValues(values, v)
	if err != nil {
		return nil, err
	}

	exp, err := Build(false, &FileExpIO{})
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read experiment file")
		return nil, err
	}
	k8sExp := k8sExperiment{
		Tasks:  exp.Tasks,
		Values: v,
	}

	// generate formatted output

	b, err := RenderTpl(k8sExp, k8sTemplateFilePath)
	if err != nil {
		return nil, err
	}
	return b, nil
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
		"toYAML":                 toYAML,
		"defaultImage":           func() string { return "iter8/iter8:" + (RootCmd.Version)[1:] },
		"iter8MajorMinorVersion": func() string { return RootCmd.Version },
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

var k8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "Generate manifest for running experiment in Kubernetes",
	Example: `
# Generate Kubernetes manifest
iter8 gen k8s`,
	// Put any option computation and/or validatiom here
	// PreRunE: func(c *cobra.Command, args []string) error {
	RunE: func(c *cobra.Command, args []string) error {
		return runGetK8sCmd(c, args)
	},
}

func init() {
	// support --id option to set identifier
	k8sCmd.Flags().StringVarP(&id, "id", "i", "", "if not specified, a randomly generated identifier will be used")
	// support --app option to set app
	k8sCmd.Flags().StringVarP(&app, "app", "a", "", "label to be associated with an experiment, default is 'default'")

	// extend gen command with the k8s command
	genCmd.AddCommand(k8sCmd)
}
