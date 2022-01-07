package basecli

import (
	"testing"
)

func TestVersion(t *testing.T) {
	versionCmd.RunE(nil, nil)
	short = true
	versionCmd.RunE(nil, nil)
}
