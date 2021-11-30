package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	Iter8Path   string = `/usr/local/bin/iter8`
	Iter8Run    string = "run"
	Iter8Gen    string = "gen"
	Iter8Assert string = "assert"
	Iter8Hub    string = "hub"
)

func Execute(command string, args ...string) (*string, error) {
	return execute(command, true, args...)
}

func ExecuteSilent(command string, args ...string) (*string, error) {
	return execute(command, false, args...)
}

func execute(command string, output bool, args ...string) (*string, error) {
	arguments := append([]string{command}, args...)
	iter8Cmd := exec.Command(Iter8Path, arguments...)
	// fmt.Printf("Executing command: %s\n", iter8Cmd)
	out, err := iter8Cmd.CombinedOutput()
	outStr := string(out)
	if output {
		fmt.Println(outStr)
	}
	return &outStr, err
}

func FetchResultsAsFile(client *kubernetes.Clientset, ns string, nm string) (err error) {
	result, err := FetchResults(client, ns, nm)
	if err != nil || result == nil {
		return errors.New("unable to process experiment result")
	}

	err = os.WriteFile("result.yaml", []byte(*result), 0644)
	if err != nil {
		return errors.New("unable to process experiment result")
	}

	return nil
}

func FetchResults(client *kubernetes.Clientset, ns string, nm string) (result *string, err error) {
	experiment, err := GetExperiment(client, ns, nm)
	if err != nil {
		return result, err
	}

	s, err := client.CoreV1().Secrets(ns).Get(context.Background(), experiment.GetName()+"-result", metav1.GetOptions{})
	if err != nil {
		return result, err
	}

	res, ok := s.Data["result"]
	if !ok {
		return result, errors.New("results not available")
	}
	resS := string(res)

	// err = os.WriteFile("result.yaml", []byte(result), 0644)
	// if err != nil {
	// 	return errors.New("unable to process experiment result")
	// }

	return &resS, nil
}
