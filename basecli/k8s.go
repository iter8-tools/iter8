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
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
)

const (
	// Path to go template file
	k8sTemplateFilePath = "k8s.tpl"
)

type k8sExperiment struct {
	// Tasks are the set of tasks specifying the experiment
	Tasks []base.TaskSpec
	// Values used to generate the experiment
	Values chartutil.Values
}

// id of the experiment
var id string

// app is the name of the app involved in the experiment
var app string

// run runs the command
func runGetK8sCmd(cmd *cobra.Command, args []string) (err error) {
	result, err := Generate()
	if err != nil {
		return err
	}
	if result != nil {
		fmt.Println(result.String())
	}
	return nil
}

// Generate the k8s experiment manifest
func Generate() (result *bytes.Buffer, err error) {
	p := getter.All(cli.New())
	v, err := GenOptions.MergeValues(p)
	if err != nil {
		return nil, err
	}

	// set id if --id option used
	if len(id) > 0 {
		v["id"] = id
	}

	// set app if --app option is used
	// if both id and app are set, then
	// the --id option will take precedence
	if len(app) > 0 {
		v["app"] = app
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

// RenderTpl creates the Kubernetes experiment manifest from k8s.tpl
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
	k8sCmd.Flags().StringVarP(&app, "app", "a", "default", "app to be associated with an experiment, default is 'default'")
	// extend gen command with the k8s command
	genCmd.AddCommand(k8sCmd)
}
