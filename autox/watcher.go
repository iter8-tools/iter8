package autox

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/iter8-tools/iter8/autox/k8sdriver"
	"github.com/iter8-tools/iter8/base/log"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	// Name of environment variable with file path to configuration yaml file
	RESOURCE_CONFIG_ENV    = "RESOURCE_CONFIG"
	CHART_GROUP_CONFIG_ENV = "CHART_GROUP_CONFIG"
)

// Start is entry point to configure services and start them
func Start(kd *k8sdriver.KubeDriver) {
	// initialize kubernetes driver
	if err := kd.Init(); err != nil {
		log.Logger.Fatal("unable to initialize kubedriver")
	}

	// read resource config (resources and namespaces to watch)
	resourceConfigFile, ok := os.LookupEnv(RESOURCE_CONFIG_ENV)
	if !ok {
		log.Logger.Fatal("resource configuation file is required")
	}

	// read group config (apps and helm charts to install)
	groupConfigFile, ok := os.LookupEnv(CHART_GROUP_CONFIG_ENV)
	if !ok {
		log.Logger.Fatal("group configuation file is required")
	}

	stopCh := make(chan struct{})

	// set up resource watching as defined by config
	resourceConfig := readConfig(resourceConfigFile)
	groupConfig := readChartGroupConfig(groupConfigFile)

	fmt.Println("groupConfig:", groupConfig)

	w := newIter8Watcher(kd, resourceConfig.Resources, resourceConfig.Namespaces, groupConfig)
	go w.start(stopCh)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)

	<-sigCh
	close(stopCh)
}
