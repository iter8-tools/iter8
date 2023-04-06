package cmd

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/controllers"
	"github.com/iter8-tools/iter8/controllers/k8sclient/fake"
	"github.com/stretchr/testify/assert"
)

func TestControllers(t *testing.T) {
	// set pod name
	_ = os.Setenv(controllers.PodNameEnvVariable, "pod-0")
	// set pod namespace
	_ = os.Setenv(controllers.PodNamespaceEnvVariable, "default")
	// set config file
	_ = os.Setenv(controllers.ConfigEnv, base.CompletePath("../", "testdata/controllers/config.yaml"))

	kubeClient = fake.New(nil, nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	cmd := newControllersCmd(ctx.Done(), kubeClient)
	err := cmd.RunE(cmd, nil)
	assert.NoError(t, err)

}
