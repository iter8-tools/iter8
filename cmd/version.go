package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	_ "embed"
)

//go:embed gitcommit.txt
var gitCommit string

// versionDesc is the description of the version command
var versionDesc = `
Print the version of Iter8 CLI.

	iter8 version

The output may look as follows:

	$ cmd.BuildInfo{Version:"v0.13.0", GitCommit:"f24e86f3d3eceb02eabbba54b40af2c940f55ad5", GoVersion:"go1.19.3"}

In the sample output shown above:

- Version is the semantic version of the Iter8 CLI.
- GitCommit is the SHA hash for the commit that this version was built from.
- GoVersion is the version of Go that was used to compile Iter8 CLI.
`

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
	// versionCmd represents the version command
	cmd := &cobra.Command{
		Use:           "version",
		Short:         "Print Iter8 CLI version",
		Long:          versionDesc,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			v := getBuildInfo()
			fmt.Printf("%#v", v)
			fmt.Println()
			fmt.Println(gitCommit)
			fmt.Println()
			return nil
		},
	}
	return cmd
}

// get returns build info
func getBuildInfo() BuildInfo {
	v := BuildInfo{
		GitCommit: gitCommit,
		GoVersion: runtime.Version(),
	}
	return v
}
