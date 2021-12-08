/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"io/ioutil"

	"github.com/go-playground/validator/v10"
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chartutil"
)

const (
	// Path to experiment template file
	expTemplatePath = "experiment.tpl"
)

// expCmd represents the exp command
var expCmd = &cobra.Command{
	Use:   "exp",
	Short: "render experiment template in the file experiment.tpl with values",
	Long: `
	Render experiment template in the file experiment.tpl with values`,
	Example: `
	# render experiment template in the file experiment.tpl with values
	iter8 gen exp --set key=val
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		v := chartutil.Values{}
		err := ParseValues(GenOptions.Values, v)
		if err != nil {
			return err
		}

		// generate formatted output
		b, err := RenderGoTpl(v, expTemplatePath)
		if err != nil {
			return err
		}
		specBytes := b.Bytes()

		// build and validate experiment here...
		s, err := SpecFromBytes(specBytes)
		e := &Experiment{
			Experiment: &base.Experiment{
				Tasks: s,
			}}
		if err != nil {
			return err
		}
		err = e.buildTasks()
		if err != nil {
			return err
		}
		validate := validator.New()
		// returns nil or ValidationErrors ( []FieldError )
		err = validate.Struct(e.Experiment)
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("invalid experiment specification")
			return err
		}

		// write experiment spec file
		err = ioutil.WriteFile(experimentSpecPath, specBytes, 0664)
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("unable to write experiment spec")
			return err
		}
		return err
	},
}

func init() {
	GenCmd.AddCommand(expCmd)
}
