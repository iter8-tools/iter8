package core

import (
	"encoding/json"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/iter8-tools/iter8/base/log"
)

// writeMetricInternal is detailed implementation of gRPC method WriteMetric
func getMetricsInternal(application string) (string, error) {
	log.Logger.Trace("getMetricsInternal called")
	defer log.Logger.Trace("getMetricsInternal ended")

	a, err := abnapp.Applications.Get(application)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(a)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
