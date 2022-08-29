package abn

import (
	"errors"
	"strconv"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/iter8-tools/iter8/base/log"
)

// writeMetricInternal is detailed implementation of gRPC method WriteMetric
func writeMetricInternal(application, user, metric, valueStr string) error {
	a, err := abnapp.Applications.Get(application, false)
	if err != nil || a == nil {
		return errors.New("unexpected: cannot find record of application " + application)
	}

	track, err := lookupInternal(application, user)
	if err != nil || track == nil {
		return err
	}

	version, ok := a.Tracks[*track]
	if !ok {
		return errors.New("track not mapped to version")
	}

	v, _ := a.GetVersion(version, false)
	if v == nil {
		return errors.New("unexpected: trying to write metrics for unknown version")
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		log.Logger.Warn("Unable to parse metric value ", valueStr)
		return nil
	}

	m, _ := v.GetMetric(metric, true)
	m.Add(value)

	// persist updated metric
	abnapp.Applications.BatchedWrite(a)

	return err
}
