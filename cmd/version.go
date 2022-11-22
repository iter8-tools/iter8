package cmd

import (
	"fmt"
	"runtime"

	"github.com/iter8-tools/iter8/base"
	"github.com/spf13/cobra"
)

// versionDesc is the description of the version command
var versionDesc = `
Print the version of Iter8 CLI.

	iter8 version

The output may look as follows:

	$ cmd.BuildInfo{Version:"v0.12.0", GitCommit:"e8e53cc18a2c8898b0abb7e418c61ed932cdac8a", GoVersion:"go1.17.13"}

In the sample output shown above:

- Version is the semantic version of the Iter8 CLI.
- GitCommit is the SHA hash for the commit that this version was built from.
- GoVersion is the version of Go that was used to compile Iter8 CLI.
`

var (
	// gitCommit is the git sha1
	gitCommit = ""
)

// BuildInfo describes the compile time information.
type BuildInfo struct {
	// Version is the semantic version
	Version string `json:"version,omitempty"`
	// GitCommit is the git sha1.
	GitCommit string `json:"git_commit,omitempty"`
	// GoVersion is the version of the Go compiler used to compile Iter8.
	GoVersion string `json:"go_version,omitempty"`
}

// newVersionCmd creates the version command
func newVersionCmd() *cobra.Command {
	var short bool
	// versionCmd represents the version command
	cmd := &cobra.Command{
		Use:           "version",
		Short:         "Print Iter8 CLI version",
		Long:          versionDesc,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			v := getBuildInfo()
			if short {
				if len(v.GitCommit) >= 7 {
					fmt.Printf("%s+g%s", base.Version, v.GitCommit[:7])
					fmt.Println()
					return nil
				}
				fmt.Println(base.Version)
				return nil
			}
			fmt.Printf("%#v", v)
			fmt.Println()
			return nil
		},
	}
	addShortFlag(cmd, &short)
	return cmd
}

// get returns build info
func getBuildInfo() BuildInfo {
	v := BuildInfo{
		Version:   base.Version,
		GitCommit: gitCommit,
		GoVersion: runtime.Version(),
	}
	return v
}

// addShortFlag adds the short flag to the version command
func addShortFlag(cmd *cobra.Command, shortPtr *bool) {
	cmd.Flags().BoolVar(shortPtr, "short", false, "print abbreviated version info")
	cmd.Flags().Lookup("short").NoOptDefVal = "true"
}
