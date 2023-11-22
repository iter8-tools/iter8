// Package client implements an implementation independent storage client
package client

import (
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
	util "github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/storage"
	"github.com/iter8-tools/iter8/storage/badgerdb"
	"github.com/iter8-tools/iter8/storage/redis"
)

const (
	metricsConfigFileEnv  = "METRICS_CONFIG_FILE"
	defaultImplementation = "badgerdb"
)

var (
	// MetricsClient is storage client
	MetricsClient storage.Interface
)

// metricsStorageConfig is configuration of metrics service
type metricsStorageConfig struct {
	// Implementation method for metrics service
	Implementation *string `json:"implementation,omitempty"`
}

// GetClient creates a metric service client based on configuration
func GetClient() (storage.Interface, error) {
	conf := &metricsStorageConfig{}
	err := util.ReadConfig(metricsConfigFileEnv, conf, func() {
		if conf.Implementation == nil {
			conf.Implementation = util.StringPointer(defaultImplementation)
		}
	})
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(*conf.Implementation) {
	case "badgerdb":
		// badgerConfig defines the configuration of a badgerDB based metrics service
		type mConfig struct {
			badgerdb.ClientConfig `json:"badgerdb,omitempty"`
		}

		conf := &mConfig{}
		err := util.ReadConfig(metricsConfigFileEnv, conf, func() {
			if conf.ClientConfig.Storage == nil {
				conf.ClientConfig.Storage = util.StringPointer("50Mi")
			}
			if conf.ClientConfig.StorageClassName == nil {
				conf.ClientConfig.StorageClassName = util.StringPointer("standard")
			}
			if conf.ClientConfig.Dir == nil {
				conf.ClientConfig.Dir = util.StringPointer("/metrics")
			}
		})
		if err != nil {
			return nil, err
		}

		cl, err := badgerdb.GetClient(badger.DefaultOptions(*conf.ClientConfig.Dir), badgerdb.AdditionalOptions{})
		if err != nil {
			return nil, err
		}
		return cl, nil

	case "redis":
		// redisConfig defines the configuration of a redis based metrics service
		type mConfig struct {
			redis.ClientConfig `json:"redis,omitempty"`
		}

		conf := &mConfig{}
		err := util.ReadConfig(metricsConfigFileEnv, conf, func() {
			if conf.ClientConfig.Address == nil {
				conf.ClientConfig.Address = util.StringPointer("redis:6379")
			}
		})
		if err != nil {
			return nil, err
		}

		cl, err := redis.GetClient(conf.ClientConfig)
		if err != nil {
			return nil, err
		}
		return cl, nil

	default:
		return nil, fmt.Errorf("no metrics store implementation for %s", *conf.Implementation)
	}
}
