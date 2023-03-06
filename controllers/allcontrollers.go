package controllers

import (
	"github.com/iter8-tools/iter8/controllers/k8sclient"
)

// Start starts all Iter8 controllers if this pod is the leader
func Start(stopCh chan struct{}, client k8sclient.Interface) error {
	config, err := readConfig()
	if err != nil {
		return err
	}

	// validate config
	if err := config.validate(); err != nil {
		return err
	}

	// everyone starts abn controller
	go startABnController(stopCh, config, client)

	// Only the leader pod starts SSA controller
	if leaderIsMe() {

		// start server-side apply controller
		go startSSAController(stopCh, config, client)

		return nil
	}

	return nil
}
