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

func TestReadResourceConfig(t *testing.T) {
	for _, tt := range []struct {
		name          string
		file          string
		numNamespaces int
		numResources  int
	}{
		{"empty", "config.empty.yaml", 0, 0},
		{"invalid", "config.invalid.yaml", 0, 0},
		{"garbage", "config.garbage.yaml", 0, 0},
		{"nofile", "config.nofile.yaml", 0, 0},
		{"nonamespaces", "resource_config.nonamespaces.yaml", 0, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := readResourceConfig(completePath("../testdata/autox_inputs", tt.file))
			assert.Equal(t, tt.numNamespaces, len(c.Namespaces))
			assert.Equal(t, tt.numResources, len(c.Resources))
		})
	}
}

func TestReadGroupChartConfig(t *testing.T) {
	for _, tt := range []struct {
		name           string
		file           string
		numChartGroups int
	}{
		{"empty", "config.empty.yaml", 0},
		{"invalid", "config.invalid.yaml", 1},
		{"garbage", "config.garbage.yaml", 0},
		{"nofile", "config.nofile.yaml", 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := readChartGroupConfig(completePath("../testdata/autox_inputs", tt.file))
			assert.Equal(t, tt.numChartGroups, len(c))
		})
	}

	c := readChartGroupConfig(completePath("../testdata/autox_inputs", "group_config.example.yaml"))
	assert.Equal(t, 2, len(c))
	assert.Equal(t, 2, len(c["myApp"].Charts))
	assert.Equal(t, 1, len(c["myApp2"].Charts))
}
