package autox

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// utility method
func completePath(prefix string, suffix string) string {
	_, filename, _, _ := runtime.Caller(1) // one step up the call stack
	return filepath.Join(filepath.Dir(filename), prefix, suffix)
}

func TestReadGroupChartConfig(t *testing.T) {
	for _, tt := range []struct {
		name           string
		file           string
		numChartGroups int
	}{
		{"empty", "config.empty.yaml", 0},
		{"invalid", "config.invalid.yaml", 0},
		{"garbage", "config.garbage.yaml", 0},
		{"nofile", "config.nofile.yaml", 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := readChartGroupConfig(completePath("../testdata/autox_inputs", tt.file))
			assert.Equal(t, tt.numChartGroups, len(c.Specs))
		})
	}

	c := readChartGroupConfig(completePath("../testdata/autox_inputs", "group_config.example.yaml"))
	assert.Equal(t, 2, len(c.Specs))
	assert.Equal(t, 2, len(c.Specs["myApp"].ReleaseSpecs))
	assert.Equal(t, 1, len(c.Specs["myApp2"].ReleaseSpecs))
}
