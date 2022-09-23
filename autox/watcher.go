package autox

import (
	"os"

	"github.com/iter8-tools/iter8/base/log"

	// auth enables automatic authentication to various hosted clouds
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	// Name of environment variable with file path to resource configuration yaml file
	resourceConfigEnv = "RESOURCE_CONFIG"
	// Name of environment variable with file path to chart group configuration yaml file
	chartGroupConfigEnv = "CHART_GROUP_CONFIG"
)

var iter8ResourceConfig resourceConfig
var iter8ChartGroupConfig chartGroupConfig

// Start is entry point to configure services and start them
func (opts *Opts) Start(stopCh chan struct{}) error {
	// initialize kubernetes driver
	if err := opts.KubeClient.init(); err != nil {
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
	iter8ResourceConfig = readResourceConfig(resourceConfigFile)
	iter8ChartGroupConfig = readChartGroupConfig(chartGroupConfigFile)

	log.Logger.Debug("chartGroupConfig:", iter8ChartGroupConfig)

	w := newIter8Watcher(opts.KubeClient)
	go w.start(stopCh)
	return nil
}
