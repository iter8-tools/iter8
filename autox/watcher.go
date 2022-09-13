package autox

import (
	"os"

	"github.com/iter8-tools/iter8/base/log"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	// Name of environment variable with file path to configuration yaml file
	RESOURCE_CONFIG_ENV    = "RESOURCE_CONFIG"
	CHART_GROUP_CONFIG_ENV = "CHART_GROUP_CONFIG"
)

// Start is entry point to configure services and start them
func Start(stopCh chan struct{}) {
	// initialize kubernetes driver
	k8sClient.init()

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

	// set up resource watching as defined by config
	resourceConfig := readResourceConfig(resourceConfigFile)
	chartGroupConfig := readChartGroupConfig(chartGroupConfigFile)

	log.Logger.Debug("chartGroupConfig:", chartGroupConfig)

	w := newIter8Watcher(resourceConfig.Resources, resourceConfig.Namespaces, chartGroupConfig)
	go w.start(stopCh)
}
