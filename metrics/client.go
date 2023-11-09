package metrics

import (
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
	util "github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/storage"
	"github.com/iter8-tools/iter8/storage/badgerdb"
	"github.com/iter8-tools/iter8/storage/redis"
)

// metricsConfig is configuration of metrics service
type metricsConfig struct {
	// Port is port number on which the metrics service should listen
	Port *int `json:"port,omitempty"`

	// Implementation method for metrics service
	Implementation *string `json:"implementation,omitempty"`
}

// GetClient creates a metric service client based on configuration
func GetClient() (storage.Interface, error) {
	conf := &metricsConfig{}
	err := util.ReadConfig(MetricsConfigFileEnv, conf, func() {
		if conf.Implementation == nil {
			conf.Implementation = util.StringPointer("badgerdb")
		}
	})
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(*conf.Implementation) {
	case "badgerdb":
		// badgerConfig defines the configuration of a badgerDB based metrics service
		type mConfig struct {
			BadgerConfig struct {
				Storage          *string `json:"storage,omitempty"`
				StorageClassName *string `json:"storageClassName,omitempty"`
				Dir              *string `json:"dir,omitempty"`
			} `json:"badgerdb,omitempty"`
		}

		conf := &mConfig{}
		err := util.ReadConfig(MetricsConfigFileEnv, conf, func() {
			if conf.BadgerConfig.Storage == nil {
				conf.BadgerConfig.Storage = util.StringPointer("50Mi")
			}
			if conf.BadgerConfig.StorageClassName == nil {
				conf.BadgerConfig.StorageClassName = util.StringPointer("standard")
			}
			if conf.BadgerConfig.Dir == nil {
				conf.BadgerConfig.Dir = util.StringPointer("/metrics")
			}
		})
		if err != nil {
			return nil, err
		}

		cl, err := badgerdb.GetClient(badger.DefaultOptions(*conf.BadgerConfig.Dir), badgerdb.AdditionalOptions{})
		if err != nil {
			return nil, err
		}
		return cl, nil

	case "redis":
		// redisConfig defines the configuration of a redis based metrics service
		type mConfig struct {
			RedisConfig struct {
				Address *string `json:"address,omitempty"`
			} `json:"redis,omitempty"`
		}

		conf := &mConfig{}
		err := util.ReadConfig(MetricsConfigFileEnv, conf, func() {
			if conf.RedisConfig.Address == nil {
				conf.RedisConfig.Address = util.StringPointer("redis:6379")
			}
		})
		if err != nil {
			return nil, err
		}

		cl, err := redis.GetClient(*conf.RedisConfig.Address)
		if err != nil {
			return nil, err
		}
		return cl, nil

	default:
		return nil, fmt.Errorf("no metrics store implementation for %s", *conf.Implementation)
	}
}
