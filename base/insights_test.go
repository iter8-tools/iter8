package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrackVersionStr(t *testing.T) {
	scenarios := map[string]struct {
		in          Insights
		expectedStr string
	}{
		"VersionNames is nil":         {in: Insights{}, expectedStr: "version 0"},
		"Version and Track empty":     {in: Insights{VersionNames: []VersionInfo{}}, expectedStr: "version 0"},
		"Track is empty":              {in: Insights{VersionNames: []VersionInfo{{Version: "version"}}}, expectedStr: "version"},
		"Version is empty":            {in: Insights{VersionNames: []VersionInfo{{Track: "track"}}}, expectedStr: "track"},
		"Version and Track not empty": {in: Insights{VersionNames: []VersionInfo{{Track: "track", Version: "version"}}}, expectedStr: "track (version)"},
	}

	for l, s := range scenarios {
		t.Run(l, func(t *testing.T) {
			assert.Equal(t, s.expectedStr, s.in.TrackVersionStr(0))
		})
	}
}
