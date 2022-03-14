package action

import (
	"testing"

	"github.com/iter8-tools/iter8/driver"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
)

// import (
// 	"context"
// 	"io/ioutil"
// 	"testing"

// 	batchv1 "k8s.io/api/batch/v1"
// 	corev1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// 	"github.com/iter8-tools/iter8/base"
// 	"github.com/iter8-tools/iter8/base/log"
// 	"github.com/jarcoal/httpmock"
// 	"github.com/stretchr/testify/assert"
// 	"helm.sh/helm/v3/pkg/cli"
// 	"k8s.io/client-go/kubernetes/fake"
// )

func TestLocalLaunch(t *testing.T) {
	httpmock.Activate()
	// Exact URL match
	httpmock.RegisterResponder("GET", "https://httpbin.org/get",
		httpmock.NewStringResponder(200, `[{"id": 1, "name": "My Great Thing"}]`))

	// fix lOpts
	lOpts := NewLaunchOpts(driver.NewFakeKubeDriver(cli.New()))
	lOpts.DestDir = t.TempDir()
	lOpts.ChartName = "load-test-http"
	lOpts.Values = []string{"url=https://httpbin.org/get"}
	err := lOpts.LocalRun()
	assert.NoError(t, err)

	httpmock.DeactivateAndReset()
}

func TestKubeLaunch(t *testing.T) {
	var err error

	// fix lOpts
	lOpts := NewLaunchOpts(driver.NewFakeKubeDriver(cli.New()))
	lOpts.ChartName = "load-test-http"
	lOpts.DestDir = t.TempDir()
	lOpts.Values = []string{"url=https://iter8.tools"}

	err = lOpts.KubeRun()
	assert.NoError(t, err)

	rel, err := lOpts.Releases.Last(lOpts.Group)
	assert.NotNil(t, rel)
	assert.NoError(t, err)
}
