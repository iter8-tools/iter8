package controllers

import (
	"errors"
	"os"
	"strings"

	"github.com/iter8-tools/iter8/base/log"
)

const (
	// podNameEnvVariable is the name of the environment variable with pod name
	podNameEnvVariable = "POD_NAME"
	// podNamespaceEnvVariable is the name of the environment variable with pod namespace
	podNamespaceEnvVariable = "POD_NAMESPACE"
	// leaderSuffix is used to determine the leader pod
	leaderSuffix = "-0"
)

// getPodName returns the name of this pod
func getPodName() (string, bool) {
	podName, ok := os.LookupEnv(podNameEnvVariable)
	// missing env variable is unacceptable
	if !ok {
		return "", false
	}
	// empty podName is unacceptable
	if len(podName) == 0 {
		return "", false
	}
	return podName, true
}

// leaderIsMe is true if this pod has the leaderSuffix ("-0")
func leaderIsMe() (bool, error) {
	podName, ok := getPodName()
	if !ok {
		e := errors.New("unable to retrieve pod name")
		log.Logger.Error(e)
		return false, e
	}
	return strings.HasSuffix(podName, leaderSuffix), nil
}
