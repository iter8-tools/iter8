package cmd

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/controllers/k8sclient/fake"
	"github.com/stretchr/testify/assert"
)

func TestControllers(t *testing.T) {
	// set pod name
	_ = os.Setenv("POD_NAME", "pod-0")
	// set pod namespace
	_ = os.Setenv("POD_NAMESPACE", "default")
	// set config file
	_ = os.Setenv("CONFIG_FILE", base.CompletePath("../", "testdata/controllers/config.yaml"))

	kubeClient = fake.New(nil, nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	cmd := newControllersCmd(ctx.Done(), kubeClient)
	err := cmd.RunE(cmd, nil)
	assert.NoError(t, err)

}
