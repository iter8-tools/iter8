package base

import (
	"helm.sh/helm/v3/pkg/cli"
)

var (
	kd = NewKubeDriver(cli.New())
)
