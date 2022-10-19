package autox

import (
	"fmt"
	"os"

	"github.com/iter8-tools/iter8/base/log"
)

const (
	// Name of environment variable with file path to spec group configuration yaml file
	configEnv = "CONFIG"
)

var k8sClient *kubeClient
var autoXConfig config

func validateConfig(c config) error {
	var err error

	for releaseGroupSpecID, releaseGroupSpec := range autoXConfig.Specs {
		// validate trigger
		if releaseGroupSpec.Trigger.Namespace == "" {
			err = fmt.Errorf("trigger in spec group \"%s\" does not have a namespace", releaseGroupSpecID)
			break
		}

		if releaseGroupSpec.Trigger.Group == "" {
			err = fmt.Errorf("trigger in spec group \"%s\" does not have a group", releaseGroupSpecID)
			break
		}

		if releaseGroupSpec.Trigger.Version == "" {
			err = fmt.Errorf("trigger in spec group \"%s\" does not have a version", releaseGroupSpecID)
			break
		}

		if releaseGroupSpec.Trigger.Resource == "" {
			err = fmt.Errorf("trigger in spec group \"%s\" does not have a resource", releaseGroupSpecID)
			break
		}
	}

	return err
}

// Start is entry point to configure services and start them
func (opts *Opts) Start(stopCh chan struct{}) error {
	// initialize kubernetes driver
	if err := opts.kubeClient.init(); err != nil {
		log.Logger.Fatal("unable to init k8s client")
	}

	k8sClient = opts.kubeClient

	// read group config (apps and Helm charts to install)
	configFile, ok := os.LookupEnv(configEnv)
	if !ok {
		log.Logger.Fatal("group configuration file is required")
	}

	// set up resource watching as defined by config
	autoXConfig = readConfig(configFile)

	err := validateConfig(autoXConfig)
	if err != nil {
		return err
	}

	log.Logger.Debug("config:", autoXConfig)

	w := newIter8Watcher(opts.kubeClient)
	go w.start(stopCh)
	return nil
}
