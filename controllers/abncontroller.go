package controllers

import "github.com/iter8-tools/iter8/controllers/k8sclient"

func startABnController(stopCh chan struct{}, config *Config, client k8sclient.Interface) {
	// For each subject, maintain
	// 	variants
	//  variants that exist
	//  variants that are ready
	//  variant weights
	//  normalize weights between 0-1000 (upper limit is const config)
	// implement GetVariant(given a number between 0 and 1000 corresponding to user)
	// use a mutexed data structure for the above with write lock
}
