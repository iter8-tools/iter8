package hub

import (
	basecli "github.com/iter8-tools/iter8/cmd"

	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	return basecli.HubCmd
}
