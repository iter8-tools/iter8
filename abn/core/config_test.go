package core

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	for _, tt := range []struct {
		name          string
		file          string
		numNamespaces int
		numResources  int
	}{
		{"empty", "config.empty.yaml", 0, 0},
		{"nonamespaces", "config.nonamespaces.yaml", 0, 1},
		{"invalid", "config.invalid.yaml", 0, 0},
		{"garbage", "config.garbage.yaml", 0, 0},
		{"nofile", "config.nofile.yaml", 0, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := readConfig(completePath("../../testdata/abninputs", tt.file))
			assert.Equal(t, tt.numNamespaces, len(c.Namespaces))
			assert.Equal(t, tt.numResources, len(c.Resources))
		})
	}
}

// utility method
func completePath(prefix string, suffix string) string {
	_, filename, _, _ := runtime.Caller(1) // one step up the call stack
	return filepath.Join(filepath.Dir(filename), prefix, suffix)
}
