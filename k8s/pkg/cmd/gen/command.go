package k8s

import (
	basecli "github.com/iter8-tools/iter8/cmd"
	k8s "github.com/iter8-tools/iter8/k8s/pkg/cmd/gen/k8s"

	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := basecli.GenCmd

	cmd.AddCommand(k8s.NewCmd())

	return cmd
}
