package hub

import (
	"os"

	basecli "github.com/iter8-tools/iter8/cmd"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/cmd/options"
)

func NewCmd() *cobra.Command {
	cmd := basecli.HubCmd
	cmd.AddCommand(options.NewCmdOptions(os.Stdout))
	return cmd
}
