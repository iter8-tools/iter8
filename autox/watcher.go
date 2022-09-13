package autox

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/iter8-tools/iter8/base/log"
	"helm.sh/helm/v3/pkg/cli"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	// Name of environment variable with file path to configuration yaml file
	RESOURCE_CONFIG_ENV    = "RESOURCE_CONFIG"
	CHART_GROUP_CONFIG_ENV = "CHART_GROUP_CONFIG"
)

// Start is entry point to configure services and start them
func Start() {
	// initialize kubernetes driver
	Client = *newKubeClient(cli.New())
	Client.init()

	// read resource config (resources and namespaces to watch)
	resourceConfigFile, ok := os.LookupEnv(RESOURCE_CONFIG_ENV)
	if !ok {
		log.Logger.Fatal("resource configuation file is required")
	}

	// read group config (apps and helm charts to install)
	chartGroupConfigFile, ok := os.LookupEnv(CHART_GROUP_CONFIG_ENV)
	if !ok {
		log.Logger.Fatal("group configuation file is required")
	}

	stopCh := make(chan struct{})

	// set up resource watching as defined by config
	resourceConfig := readResourceConfig(resourceConfigFile)
	chartGroupConfig := readChartGroupConfig(chartGroupConfigFile)

	log.Logger.Debug("chartGroupConfig:", chartGroupConfig)

	w := newIter8Watcher(resourceConfig.Resources, resourceConfig.Namespaces, chartGroupConfig)
	go w.start(stopCh)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)

	<-sigCh
	close(stopCh)
}
