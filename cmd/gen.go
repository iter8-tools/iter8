package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"helm.sh/helm/pkg/strvals"
	"helm.sh/helm/v3/pkg/chartutil"
)

const (
	// Path to template file
	templateFilePath = "iter8.tpl"
)

var (
	// values are user specified values used during gen
	values []string
)

type ExperimentWithValues struct {
	*Experiment
	Values chartutil.Values
}

func (e *ExperimentWithValues) parseValues(values []string) error {
	// User specified a value via --set
	for _, value := range values {
		if err := strvals.ParseInto(value, e.Values); err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("failed parsing --set data")
			return errors.Wrap(err, "failed parsing --set data")
		}
	}
	return nil
}

// GenCmd represents the gen command
var GenCmd = &cobra.Command{
	Use:   "gen",
	Short: "generate formatted output from experiment spec and result",
	Long:  "Generate formatted output from experiment spec and result",
	Example: `
	# download the load-test experiment
	iter8 hub -e load-test

	cd load-test

	# run it
	iter8 run

	# generate formatted text output
	iter8 gen
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Logger.Trace("build started")
		// build experiment
		// replace FileExpIO with ClusterExpIO to build from cluster
		fio := &FileExpIO{}
		exp, err := Build(true, fio)
		log.Logger.Trace("build finished")
		if err != nil {
			return err
		}

		ev := &ExperimentWithValues{
			Experiment: exp,
		}
		err = ev.parseValues(values)
		if err != nil {
			return err
		}
		// generate formatted output
		err = ev.Gen()
		if err != nil {
			return err
		}
		return nil
	},
}

// Gen creates formatted output from experiment spec and result
func (exp *ExperimentWithValues) Gen() error {
	var tmpl *template.Template
	var err error

	// read in the template file
	tplBytes, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to read template file")
		return err
	}

	// add toYAML and other sprig template functions
	// they are all allowed to be used within the custom template
	// ensure it is a valid template
	tmpl, err = template.New("tpl").Funcs(template.FuncMap{
		"toYAML": toYAML,
	}).Option("missingkey=error").Funcs(sprig.TxtFuncMap()).Parse(string(tplBytes))
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to parse template file")
		return err
	}

	// execute template
	var b bytes.Buffer
	err = tmpl.Execute(&b, exp)
	if err != nil {
		log.Logger.WithStackTrace(err.Error()).Error("unable to execute template")
		return err
	}

	// print output
	fmt.Println(b.String())
	return nil
}

func init() {
	RootCmd.AddCommand(GenCmd)
	GenCmd.Flags().StringSliceVarP(&values, "set", "s", []string{}, "key=value; value can be accessed in templates used by gen {{ Values.key }}")
}
