package core

import (
	"errors"
	"fmt"
	"strconv"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	"github.com/iter8-tools/iter8/base/log"
)

// writeMetricInternal is detailed implementation of gRPC method WriteMetric
func writeMetricInternal(application, user, metric, valueStr string) error {
	a, track, err := lookupInternal(application, user)
	if err != nil || track == nil {
		return err
	}
	log.Logger.Debug(fmt.Sprintf("lookup(%s,%s) -> %s", application, user, *track))

	// lock for write; we will modify the metric
	abnapp.Applications.Lock(application)
	defer abnapp.Applications.Unlock(application)

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
		return err
	}

	m, _ := v.GetMetric(metric, true)
	m.Add(value)

	// persist updated metric
	return abnapp.Applications.BatchedWrite(a)
}
