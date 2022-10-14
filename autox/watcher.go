package autox

import (
	"fmt"
	"os"

	"github.com/iter8-tools/iter8/base/log"
)

const (
	// Name of environment variable with file path to chart group configuration yaml file
	chartGroupConfigEnv = "CHART_GROUP_CONFIG"
)

var k8sClient *KubeClient
var iter8ChartGroupConfig chartGroupConfig

func validateChartGroupConfig(cgc chartGroupConfig) error {
	var err error

	for chartGroupID, chartGroup := range iter8ChartGroupConfig.Specs {
		// validate trigger
		if chartGroup.Trigger.Namespace == "" {
			err = fmt.Errorf("trigger in chart group \"%s\" does not have a namespace", chartGroupID)
			break
		}

		if chartGroup.Trigger.Group == "" {
			err = fmt.Errorf("trigger in chart group \"%s\" does not have a group", chartGroupID)
			break
		}

		if chartGroup.Trigger.Version == "" {
			err = fmt.Errorf("trigger in chart group \"%s\" does not have a version", chartGroupID)
			break
		}

		if chartGroup.Trigger.Resource == "" {
			err = fmt.Errorf("trigger in chart group \"%s\" does not have a resource", chartGroupID)
			break
		}
	}

	return err
}

// Start is entry point to configure services and start them
func (opts *Opts) Start(stopCh chan struct{}) error {
	// initialize kubernetes driver
	if err := opts.KubeClient.init(); err != nil {
		log.Logger.Fatal("unable to init k8s client")
	}

	k8sClient = opts.KubeClient

	// read group config (apps and helm charts to install)
	chartGroupConfigFile, ok := os.LookupEnv(chartGroupConfigEnv)
	if !ok {
		log.Logger.Fatal("group configuration file is required")
	}

	// set up resource watching as defined by config
	iter8ChartGroupConfig = readChartGroupConfig(chartGroupConfigFile)

	err := validateChartGroupConfig(iter8ChartGroupConfig)
	if err != nil {
		return err
	}

	log.Logger.Debug("chartGroupConfig:", iter8ChartGroupConfig)

	w := newIter8Watcher(opts.KubeClient)
	go w.start(stopCh)
	return nil
}
