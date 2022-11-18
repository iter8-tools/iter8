package autox

import (
	"fmt"
	"os"

	"github.com/iter8-tools/iter8/base/log"

	"helm.sh/helm/v3/pkg/cli"
)

const (
	// Name of environment variable with file path to spec group configuration yaml file
	configEnv = "CONFIG"
)

var k8sClient *kubeClient

func validateConfig(c config) error {
	var err error

	triggerStrings := map[string]bool{}

	for releaseGroupSpecID, releaseGroupSpec := range c.Specs {
		// validate trigger
		if releaseGroupSpec.Trigger.Name == "" {
			err = fmt.Errorf("trigger in spec group \"%s\" does not have a name", releaseGroupSpecID)
			break
		}

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

		// check for trigger uniqueness
		triggerString := fmt.Sprintf("%s/%s/%s/%s/%s", releaseGroupSpec.Trigger.Name, releaseGroupSpec.Trigger.Namespace, releaseGroupSpec.Trigger.Group, releaseGroupSpec.Trigger.Version, releaseGroupSpec.Trigger.Resource)
		if _, ok := triggerStrings[triggerString]; ok {
			err = fmt.Errorf("multiple release specs with the same trigger: name: \"%s\", namespace: \"%s\", group: \"%s\", version: \"%s\", resource: \"%s\",", releaseGroupSpec.Trigger.Name, releaseGroupSpec.Trigger.Namespace, releaseGroupSpec.Trigger.Group, releaseGroupSpec.Trigger.Version, releaseGroupSpec.Trigger.Resource)
			break
		}
		triggerStrings[triggerString] = true
	}

	return err
}

// Start is entry point to configure services and start them
func Start(stopCh chan struct{}, autoxK *kubeClient) error {
	if autoxK == nil {
		// get a default client
		k8sClient = newKubeClient(cli.New())
	} else {
		// set it here
		k8sClient = autoxK
	}

	// initialize kubernetes driver
	if err := k8sClient.init(); err != nil {
		log.Logger.Fatal("unable to init k8s client")
	}

	// read group config (apps and Helm charts to install)
	configFile, ok := os.LookupEnv(configEnv)
	if !ok {
		log.Logger.Fatal("group configuration file is required")
	}

	// set up resource watching as defined by config
	autoXConfig := readConfig(configFile)

	err := validateConfig(autoXConfig)
	if err != nil {
		return err
	}

	log.Logger.Debug("config:", autoXConfig)

	w := newIter8Watcher(autoXConfig)
	go w.start(stopCh)
	return nil
}
