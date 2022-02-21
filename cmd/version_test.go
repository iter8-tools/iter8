package cmd

import (
	"testing"
)

func TestVersion(t *testing.T) {
	versionCmd.Run(nil, nil)
	short = true
	versionCmd.Run(nil, nil)
}
