package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/iter8-tools/iter8/base"
	"golang.org/x/sys/unix"
)

// GetVolumeUsage gets the available and total capacity of a volume, in that order
func GetVolumeUsage(path string) (uint64, uint64, error) {
	var stat unix.Statfs_t
	err := unix.Statfs(path, &stat)
	if err != nil {
		return 0, 0, err
	}

	// Available blocks * size per block = available space in bytes
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	// Total blocks * size per block = available space in bytes
	totalBytes := stat.Blocks * uint64(stat.Bsize)

	return availableBytes, totalBytes, nil
}

func validateKeyToken(s string) error {
	if strings.Contains(s, ":") {
		return errors.New("key token contains \":\"")
	}

	return nil
}

// GetMetricKeyPrefix returns the prefix of a metric key
func GetMetricKeyPrefix(applicationName string, version int, signature string) string {
	return fmt.Sprintf("kt-metric::%s::%d::%s::", applicationName, version, signature)
}

// GetMetricKey returns a metric key from the inputs
func GetMetricKey(applicationName string, version int, signature, metric, user, transaction string) (string, error) {
	if err := validateKeyToken(applicationName); err != nil {
		return "", errors.New("application name cannot have \":\"")
	}
	if err := validateKeyToken(signature); err != nil {
		return "", errors.New("signature cannot have \":\"")
	}
	if err := validateKeyToken(metric); err != nil {
		return "", errors.New("metric name cannot have \":\"")
	}
	if err := validateKeyToken(user); err != nil {
		return "", errors.New("user name cannot have \":\"")
	}
	if err := validateKeyToken(transaction); err != nil {
		return "", errors.New("transaction ID cannot have \":\"")
	}

	return fmt.Sprintf("%s%s::%s::%s", GetMetricKeyPrefix(applicationName, version, signature), metric, user, transaction), nil
}

// GetUserKeyPrefix returns the prefix of a user key
func GetUserKeyPrefix(applicationName string, version int, signature string) string {
	prefix := fmt.Sprintf("kt-users::%s::%d::%s::", applicationName, version, signature)
	return prefix
}

// GetUserKey returns a user key from the inputs
func GetUserKey(applicationName string, version int, signature, user string) string {
	key := fmt.Sprintf("%s%s", GetUserKeyPrefix(applicationName, version, signature), user)
	return key
}

// GetExperimentResultKey returns a performance experiment key from the inputs
func GetExperimentResultKey(namespace, experiment string) string {
	// getExperimentResultKey() is just getUserPrefix() with the user appended at the end
	return fmt.Sprintf("kt-result::%s::%s", namespace, experiment)
}

// GetExperimentResult returns an experiment result retrieved from a key value store
func GetExperimentResult(fetch func() ([]byte, error)) (*base.ExperimentResult, error) {
	value, err := fetch()
	if err != nil {
		return nil, err
	}

	experimentResult := base.ExperimentResult{}
	err = json.Unmarshal(value, &experimentResult)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal ExperimentResult: \"%s\": %e", string(value), err)
	}

	return &experimentResult, err
}
