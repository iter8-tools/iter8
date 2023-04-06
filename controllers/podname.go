package controllers

import (
	"errors"
	"os"
	"strings"

	"github.com/iter8-tools/iter8/base/log"
)

const (
	PodNameEnvVariable      = "POD_NAME"
	PodNamespaceEnvVariable = "POD_NAMESPACE"
	statefulSetSep          = "-"
	leaderSuffix            = "-0"
)

// getPodName returns the name of this pod
func getPodName() (string, bool) {
	podName, ok := os.LookupEnv(PodNameEnvVariable)
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

// getStatefulSetName returns the name of this statefulset
func getStatefulSetName() (string, bool) {
	podName, ok := getPodName()
	if !ok {
		return "", false
	}
	// if statefulset name is x, then
	// podName is x-i, where i is the pod index (integer)
	// hence, statefulset name is simple the prefix of the pod name
	slice := strings.Split(podName, statefulSetSep)
	if len(slice) < 2 {
		return "", false
	}
	slice = slice[:len(slice)-1]
	return strings.Join(slice, statefulSetSep), true
}

// getLeaderName returns the name of the leader pod in this statefulset
// leader pod is the pod with index 0
func getLeaderName() (string, bool) {
	statefulSetName, ok := getStatefulSetName()
	if !ok {
		return "", false
	}
	return statefulSetName + statefulSetSep + "0", true
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
