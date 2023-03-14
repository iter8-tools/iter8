package controllers

import (
	"errors"
	"os"
	"strings"

	"github.com/iter8-tools/iter8/base/log"
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

func leaderIsMe() (bool, error) {
	log.Logger.Trace("invoking get pod name ...")
	podName, ok := getPodName()
	log.Logger.Trace("invoked get pod name ...")
	if !ok {
		e := errors.New("unable to retrieve pod name")
		log.Logger.Error(e)
		return false, e
	}
	log.Logger.Trace("found podName: ", podName)
	return strings.HasSuffix(podName, leaderSuffix), nil
}
