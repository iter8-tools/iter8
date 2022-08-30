package watcher

import (
	"testing"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/stretchr/testify/assert"
)

type applicationAssertion struct {
	namespace, name  string
	tracks, versions []string
}

func assertApplication(t *testing.T, a *abnapp.Application, assertion applicationAssertion) {
	assert.NotNil(t, a)
	assert.Contains(t, a.String(), assertion.namespace+"/"+assertion.name)

	assert.Equal(t, assertion.name, abnapp.GetNameFromKey(a.Name))
	assert.Equal(t, assertion.namespace, abnapp.GetNamespaceFromKey(a.Name))

	assert.Len(t, a.Tracks, len(assertion.tracks))
	for _, track := range assertion.tracks {
		assert.Contains(t, a.Versions, a.Tracks[track])
	}
	assert.Len(t, a.Versions, len(assertion.versions))

	for _, v := range a.Versions {
		assert.NotNil(t, v.Metrics)
	}
}
