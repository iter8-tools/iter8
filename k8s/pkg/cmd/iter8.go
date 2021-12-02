package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	basecli "github.com/iter8-tools/iter8/cmd"
	assert "github.com/iter8-tools/iter8/k8s/pkg/cmd/assert"
	gen "github.com/iter8-tools/iter8/k8s/pkg/cmd/gen"
	get "github.com/iter8-tools/iter8/k8s/pkg/cmd/get"
	hub "github.com/iter8-tools/iter8/k8s/pkg/cmd/hub"
	report "github.com/iter8-tools/iter8/k8s/pkg/cmd/report"
	run "github.com/iter8-tools/iter8/k8s/pkg/cmd/run"
)

func NewCmdIter8Command() *cobra.Command {
	root := basecli.RootCmd
	groups := templates.CommandGroups{
		{
			Message: "Available Commands:",
			Commands: []*cobra.Command{
				hub.NewCmd(),
				gen.NewCmd(),
				run.NewCmd(),
				get.NewCmd(),
				assert.NewCmd(),
				report.NewCmd(),
			},
		},
	}
	groups.Add(root)

	// // Commands to prepare
	// root.AddCommand(hub.NewCmd())
	// root.AddCommand(gen.NewCmd())
	// // Commands to run
	// root.AddCommand(run.NewCmd())
	// // Commands to inspect/analyze
	// root.AddCommand(get.NewCmd())
	// root.AddCommand(assert.NewCmd())
	// root.AddCommand(report.NewCmd())

	filters := []string{}

	templates.ActsAsRootCommand(root, filters, groups...)

	return root
}
