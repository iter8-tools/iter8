package autox

import (
	"os"

	"github.com/iter8-tools/iter8/base/log"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	// Name of environment variable with file path to resource configuration yaml file
	resourceConfigEnv = "RESOURCE_CONFIG"
	// Name of environment variable with file path to chart group configuration yaml file
	chartGroupConfigEnv = "CHART_GROUP_CONFIG"
)

// Start is entry point to configure services and start them
func Start(stopCh chan struct{}) error {
	// initialize kubernetes driver
	if err := k8sClient.init(); err != nil {
		log.Logger.Fatal("unable to init k8s client")
	}

	// read resource config (resources and namespaces to watch)
	resourceConfigFile, ok := os.LookupEnv(resourceConfigEnv)
	if !ok {
		log.Logger.Fatal("resource configuration file is required")
	}

	// read group config (apps and helm charts to install)
	chartGroupConfigFile, ok := os.LookupEnv(chartGroupConfigEnv)
	if !ok {
		log.Logger.Fatal("group configuration file is required")
	}

	// set up resource watching as defined by config
	resourceConfig := readResourceConfig(resourceConfigFile)
	chartGroupConfig := readChartGroupConfig(chartGroupConfigFile)

	log.Logger.Debug("chartGroupConfig:", chartGroupConfig)

	w := newIter8Watcher(resourceConfig.Resources, resourceConfig.Namespaces, chartGroupConfig)
	go w.start(stopCh)
	return nil
}
