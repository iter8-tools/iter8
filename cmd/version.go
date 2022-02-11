package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var short bool

var (
	// version is intended to be set using LDFLAGS at build time
	// In the absence of complete semantic versioning info, the best we can do is major minor
	version = "v0.9"
	// metadata is extra build time data
	metadata = ""
	// gitCommit is the git sha1
	gitCommit = ""
	// gitTreeState is the state of the git tree
	gitTreeState = ""
)

// BuildInfo describes the compile time information.
type BuildInfo struct {
	// Version is the current semver.
	Version string `json:"version,omitempty"`
	// GitCommit is the git sha1.
	GitCommit string `json:"git_commit,omitempty"`
	// GitTreeState is the state of the git tree.
	GitTreeState string `json:"git_tree_state,omitempty"`
	// GoVersion is the version of the Go compiler used to compile Iter8.
	GoVersion string `json:"go_version,omitempty"`
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Iter8 version information",
	Long: `
Show the version for Iter8.
`,
	Example: `
	iter8 version

	# The output will look like this:

	# version.BuildInfo{Version:"v0.8.32", GitCommit:"fe51cd1e31e6a202cba7aliv9552a6d418ded79a", GitTreeState:"clean", GoVersion:"go1.17.6"}

	# - Version is the semantic version of the release.
	# - GitCommit is the SHA for the commit that this version was built from.
	# - GitTreeState is "clean" if there are no local code changes when this binary was
	# 	built, and "dirty" if the binary was built from locally modified code.
	# - GoVersion is the version of Go that was used to compile Iter8.
`,
	Run: func(cmd *cobra.Command, args []string) {
		v := get()
		if short {
			if len(v.GitCommit) >= 7 {
				fmt.Printf("%s+g%s", v.Version, v.GitCommit[:7])
				fmt.Println()
				return
			}
			fmt.Println(getVersion())
			return
		}
		fmt.Printf("%#v", v)
		fmt.Println()
	},
}

// getVersion returns the semver string of the version
func getVersion() string {
	if metadata == "" {
		return version
	}
	return version + "+" + metadata
}

// get returns build info
func get() BuildInfo {
	v := BuildInfo{
		Version:      getVersion(),
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		GoVersion:    runtime.Version(),
	}
	return v
}

func init() {
	rootCmd.AddCommand(versionCmd)
	f := versionCmd.Flags()
	f.BoolVar(&short, "short", false, "print the version number")
}
