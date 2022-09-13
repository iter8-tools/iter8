package autox

import (
	"os"
	"testing"
	"time"
)

func TestAutoXStart(t *testing.T) {
	// Start requires some environment variables to be set
	os.Setenv(RESOURCE_CONFIG_ENV, "../testdata/autox_inputs/resource_config.example.yaml")
	os.Setenv(CHART_GROUP_CONFIG_ENV, "../testdata/autox_inputs/group_config.example.yaml")

	stopCh := make(chan struct{})
	Start(stopCh)

	time.Sleep(1 * time.Second)

	close(stopCh)
}
