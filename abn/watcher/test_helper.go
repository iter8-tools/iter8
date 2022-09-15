package watcher

import (
	"strings"
	"testing"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/stretchr/testify/assert"
)

type applicationAssertion struct {
	namespace, name  string
	tracks, versions []string
}

func assertApplication(t *testing.T, a *abnapp.Application, assertion applicationAssertion) bool {
	r := true
	r = r && assert.NotNil(t, a)
	r = r && assert.Contains(t, a.String(), assertion.namespace+"/"+assertion.name)

	namespace, name := splitApplicationKey(a.GetName())
	r = r && assert.Equal(t, assertion.name, name)
	r = r && assert.Equal(t, assertion.namespace, namespace)

	r = r && assert.Len(t, a.GetTracks(), len(assertion.tracks))
	for _, track := range assertion.tracks {
		r = r && assert.Contains(t, a.Versions, a.GetTracks()[track])
	}
	r = r && assert.Len(t, a.Versions, len(assertion.versions))

	for _, v := range a.Versions {
		r = r && assert.NotNil(t, v.Metrics)
	}

	return r
}

func splitApplicationKey(applicationName string) (string, string) {
	var name, namespace string
	names := strings.Split(applicationName, "/")
	if len(names) > 1 {
		namespace, name = names[0], names[1]
	} else {
		namespace, name = "default", names[0]
	}

	return namespace, name
}
