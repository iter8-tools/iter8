package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
)

func TestInitKube(t *testing.T) {
	kubeDriver := NewKubeDriver(cli.New())
	err := kubeDriver.initKube()

	assert.NoError(t, err)
}
