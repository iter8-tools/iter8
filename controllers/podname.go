package controllers

import (
	"os"
	"strings"
)

const podNameEnvVariable = "PODNAME"
const statefulSetSep = "-"
const leaderSuffix = "-0"

func getPodName() (string, bool) {
	podName, ok := os.LookupEnv(podNameEnvVariable)
	if !ok {
		return "", false
	}
	if len(podName) == 0 {
		return "", false
	}
	return podName, true
}

func getStatefulSetName() (string, bool) {
	podName, ok := getPodName()
	if !ok {
		return "", false
	}
	slice := strings.Split(podName, statefulSetSep)
	if len(slice) < 2 {
		return "", false
	}
	slice = slice[:len(slice)-1]
	return strings.Join(slice, statefulSetSep), true
}

func getLeaderName() (string, bool) {
	statefulSetName, ok := getStatefulSetName()
	if !ok {
		return "", false
	}
	return statefulSetName + statefulSetSep + "0", true
}

func leaderIsMe() bool {
	podName, ok := getPodName()
	if !ok {
		return false
	}
	return strings.HasSuffix(podName, leaderSuffix)
}
