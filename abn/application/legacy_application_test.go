// package application contains deprecated legacy data structures used by the A/B/n service
// They will be removed when a replacement for storing metrics in secrets is implemented
package abn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type GetVersionScenario struct {
	version       string
	exists        bool
	allowCreation bool
}

func TestGetVersion(t *testing.T) {
	a := &LegacyApplication{
		Name: "application",
		Versions: map[string]*LegacyVersion{
			"version": {},
		},
	}

	for _, scenario := range map[string]GetVersionScenario{
		"a": {version: "version", exists: true, allowCreation: true},
		"b": {version: "version", exists: true, allowCreation: false},
		"c": {version: "notaversion", exists: false, allowCreation: true},
		"d": {version: "notaveresion", exists: false, allowCreation: false},
	} {
		v, created := a.GetVersion(scenario.version, scenario.allowCreation)
		if scenario.exists {
			assert.NotNil(t, v)
			assert.False(t, created)
		} else {
			if scenario.allowCreation {
				assert.NotNil(t, v)
				assert.True(t, created)
			} else {
				assert.Nil(t, v)
				assert.False(t, created)
			}
		}
	}
}
