// Package redis implements the storage interface with Redis
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/storage"
	"github.com/redis/go-redis/v9"
)

// SetMetric records a metric value; see storage.Interface
func (cl Client) SetMetric(applicationName string, version int, signature, metric, user, transaction string, metricValue float64) error {
	key, err := storage.GetMetricKey(applicationName, version, signature, metric, user, transaction)
	if err != nil {
		return err
	}

	err = cl.rdb.Set(context.Background(), key, metricValue, 0).Err()
	if err != nil {
		return fmt.Errorf("cannot set metric with key \"%s\": %w", key, err)
	}

	err = cl.SetUser(applicationName, version, signature, user)
	return err
}

// SetUser records the name of a user. See storage.Inferface
func (cl Client) SetUser(applicationName string, version int, signature, user string) error {
	key := storage.GetUserKey(applicationName, version, signature, user)

	err := cl.rdb.Set(context.Background(), key, []byte("true"), 0).Err()
	if err != nil {
		return fmt.Errorf("cannot set metric with key \"%s\": %w", key, err)
	}
	return err
}

// GetMetrics returns all metrics for an app/version. See storage.Inferface
func (cl Client) GetMetrics(applicationName string, version int, signature string) (*storage.VersionMetrics, error) {
	metrics := storage.VersionMetrics{}
	userCount, err := cl.getUserCount(applicationName, version, signature)
	if err != nil {
		return nil, err
	}

	var currentMetric string
	var currentUser string

	var cumulativeUserValue float64

	var metricsOverTransactions []float64
	var metricsOverUsers []float64

	prefix := storage.GetMetricKeyPrefix(applicationName, version, signature)
	ctx := context.Background()
	cursor := uint64(0)
	it := cl.rdb.Scan(ctx, cursor, prefix+"*", int64(0)).Iterator()
	for it.Next(ctx) {
		key := it.Val()
		tokens := strings.Split(key, "::")
		if len(tokens) != 7 {
			return nil, fmt.Errorf("incorrect number of toekns in metrics key")
		}
		metric := tokens[4]
		user := tokens[5]

		value, err := cl.rdb.Get(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}

		if metric != currentMetric && currentMetric != "" {
			metricsOverUsers = append(metricsOverUsers, cumulativeUserValue)

			// add 0s for all the users that did not produce metrics; for example, via Lookup()
			diff := userCount - uint64(len(metricsOverUsers))
			for j := uint64(0); j < diff; j++ {
				metricsOverUsers = append(metricsOverUsers, 0)
			}

			metrics[currentMetric] = struct {
				MetricsOverTransactions []float64
				MetricsOverUsers        []float64
			}{
				MetricsOverTransactions: metricsOverTransactions,
				MetricsOverUsers:        metricsOverUsers,
			}

			// currentMetric = ""
			// currentUser = ""
			cumulativeUserValue = 0
			metricsOverTransactions = []float64{}
			metricsOverUsers = []float64{}
		}

		metricsOverTransactions = append(metricsOverTransactions, floatValue)

		if user != currentUser && currentUser != "" {
			metricsOverUsers = append(metricsOverUsers, cumulativeUserValue)

			cumulativeUserValue = 0
		}
		cumulativeUserValue += floatValue

		currentMetric = metric
		currentUser = user
	}

	// flush last sequence of metric data
	if currentMetric != "" || currentUser != "" {
		metricsOverUsers = append(metricsOverUsers, cumulativeUserValue)

		// add 0s for all the users that did not produce metrics
		// for example, via lookup()
		if uint64(len(metricsOverUsers)) < userCount {
			diff := userCount - uint64(len(metricsOverUsers))
			for j := uint64(0); j < diff; j++ {
				metricsOverUsers = append(metricsOverUsers, 10)
			}
		}

		metrics[currentMetric] = struct {
			MetricsOverTransactions []float64
			MetricsOverUsers        []float64
		}{
			MetricsOverTransactions: metricsOverTransactions,
			MetricsOverUsers:        metricsOverUsers,
		}
	}

	return &metrics, nil
}

// SetExperimentResult records an experiment result. See storage.Inferface
func (cl Client) SetExperimentResult(namespace, experiment string, data *base.ExperimentResult) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("cannot JSON marshal ExperimentResult: %e", err)
	}

	key := storage.GetExperimentResultKey(namespace, experiment)
	err = cl.rdb.Set(context.Background(), key, dataBytes, 0).Err()
	return err
}

// GetExperimentResult returns an experiment result. See storage.Interface
func (cl Client) GetExperimentResult(namespace, experiment string) (*base.ExperimentResult, error) {
	return storage.GetExperimentResult(func() ([]byte, error) {
		return cl.rdb.Get(context.Background(), storage.GetExperimentResultKey(namespace, experiment)).Bytes()
	})
}

// Client is a client for Redis
type Client struct {
	rdb *redis.Client
}

// GetClient returns a Redis client
func GetClient(addr string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // default DB
	})

	return &Client{
		rdb: rdb,
	}, nil
}

// getUserCount gets the number of users
func (cl Client) getUserCount(applicationName string, version int, signature string) (uint64, error) {
	ctx := context.Background()

	count := uint64(0)

	prefix := storage.GetUserKeyPrefix(applicationName, version, signature)
	cursor := uint64(0)
	it := cl.rdb.Scan(ctx, cursor, prefix+"*", int64(0)).Iterator()
	for it.Next(ctx) {
		count++
	}

	return count, nil
}
