package watcher

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestReadServiceConfig tests invalid/problem cases.
// The good case is actually tested in TestGetApplicationConfig below.
func TestReadServiceConfig(t *testing.T) {
	for _, tt := range []struct {
		name          string
		file          string
		numNamespaces int
		numResources  int
	}{
		{"empty", "config.empty.yaml", 0, 0},
		{"garbage", "config.garbage.yaml", 0, 0},
		{"nofile", "config.nofile.yaml", 0, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := readServiceConfig(completePath("../../testdata/abninputs", tt.file))
			assert.Equal(t, tt.numNamespaces, len(c))
		})
	}
}

// TestGetApplicationConfig reads a good service configurationa
// each application config is then extracted
func TestGetApplicationConfig(t *testing.T) {
	for _, tt := range []struct {
		file             string
		namespace        string
		name             string
		maxNumCandidates int
		numResources     int
	}{
		{"config.yaml", "default", "backend", 2, 2},
		{"config.yaml", "default", "frontend", 1, 1},
		{"config.yaml", "test", "iter8", 1, 1},
		{"config.yaml", "empty", "frontend", 0, 0},
		{"config.yaml", "default", "notpresent", 0, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			svcConfig := readServiceConfig(completePath("../../testdata/abninputs", tt.file))
			assert.Contains(t, svcConfig, tt.namespace)
			appConfig := getApplicationConfig(tt.namespace, tt.name, svcConfig)
			if tt.maxNumCandidates != 0 {
				assert.NotNil(t, appConfig)
				assert.Equal(t, tt.maxNumCandidates, appConfig.MaxNumCandidates)
				assert.Equal(t, tt.numResources, len(appConfig.Resources))
			}
		})
	}
}

// utility method
func completePath(prefix string, suffix string) string {
	_, filename, _, _ := runtime.Caller(1) // one step up the call stack
	return filepath.Join(filepath.Dir(filename), prefix, suffix)
}
