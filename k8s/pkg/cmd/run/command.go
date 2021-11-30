package run

import (
	"fmt"

	basecli "github.com/iter8-tools/iter8/cmd"

	"github.com/spf13/cobra"
)

var example = `
	# Run an experiment in a Kubernetes cluster
	iter8 gen -o k8s | kubectl apply -f -
`

func NewCmd() *cobra.Command {
	cmd := basecli.RunCmd
	cmd.Example = fmt.Sprintf("%s%s\n", cmd.Example, example)

	return cmd
}
