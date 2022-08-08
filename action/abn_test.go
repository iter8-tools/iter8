package action

// import (
// 	"os"
// 	"testing"

// 	"helm.sh/helm/v3/pkg/cli"

// 	"github.com/iter8-tools/iter8/base"
// 	"github.com/iter8-tools/iter8/driver"
// 	"github.com/stretchr/testify/assert"
// )

// func TestAbnStart(t *testing.T) {
// 	lOpts := NewAbnOpts(driver.NewFakeKubeDriver(cli.New()))

// 	os.Setenv("WATCHER_CONFIG", base.CompletePath("../", "testdata/abninputs/config.yaml"))

// 	err := lOpts.LocalRun()
// 	assert.NoError(t, err)
// }
