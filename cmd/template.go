package cmd

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"

	"github.com/Masterminds/sprig"
	"github.com/iter8-tools/iter8/core"
	"github.com/iter8-tools/iter8/engine"
	task "github.com/iter8-tools/iter8/tasks"
	"github.com/spf13/cobra"
)

var (
	// Path to template file
	// this variable is intended to be modified in tests, and nowhere else
	filePath = "iter8.tpl"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "generate textual output of the experiment using a Go template",
	Run: func(cmd *cobra.Command, args []string) {
		// read in the template file
		tplBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			core.Logger.WithStackTrace(err.Error()).Error("unable to read template file")
			os.Exit(1)
		}

		// ensure it is a valid template
		tmpl, err := template.New("tpl").Funcs(template.FuncMap(engine.FuncMap)).Option("missingkey=error").Funcs(sprig.FuncMap()).Parse(string(tplBytes))
		if err != nil {
			core.Logger.WithStackTrace(err.Error()).Error("unable to parse template file")
			os.Exit(1)
		}

		// build experiment
		exp := &core.Experiment{
			TaskMaker: &task.TaskMaker{},
		}
		core.Logger.Trace("build started")
		err = exp.Build(false)
		core.Logger.Trace("build finished")
		if err != nil {
			core.Logger.Error("experiment build failed")
			os.Exit(1)
		}

		// execute template
		var b bytes.Buffer
		err = tmpl.Execute(&b, exp)
		if err != nil {
			core.Logger.WithStackTrace(err.Error()).Error("unable to execute template")
			os.Exit(1)
		}

		// print output
		fmt.Println(b.String())
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
}
