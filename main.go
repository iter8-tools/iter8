package main

import (
	"github.com/iter8-tools/iter8/cmd"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	cmd.Execute()
}
