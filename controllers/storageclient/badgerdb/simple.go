// Package badgerdb implements the storageclient interface with BadgerDB
package badgerdb

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/imdario/mergo"
	"github.com/iter8-tools/iter8/controllers/storageclient"
)

// Client is a client for the BadgerDB
type Client struct {
	db *badger.DB

	additionalOptions AdditionalOptions
}

// AdditionalOptions are additional options for setting up BadgerDB
type AdditionalOptions struct {
	TTL time.Duration
}

// GetClient gets a client for the BadgerDB
func GetClient(opts badger.Options, additionalOptions AdditionalOptions) (*Client, error) {
	// check if Dir and ValueDir are set and are equal
	dir := opts.Dir           // Dir is the path of the directory where key data will be stored in.
	valueDir := opts.ValueDir // ValueDir is the path of the directory where value data will be stored in.

	if dir == "" {
		return nil, errors.New("dir not set")
	} else if valueDir == "" {
		return nil, errors.New("valueDir not set")

	} else if dir != valueDir {
		return nil, errors.New("dir and valueDir are different values")
	}

	// check if path exists (if volume has been mounted)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, errors.New("path does not exist; volume has not been mounted")
	}

	// set default options
	mergedOpts := badger.DefaultOptions("")
	if err := mergo.Merge(&mergedOpts, opts); err != nil {
		return nil, errors.New("cannot configure default options for BadgerDB")
	}

	// initialize BadgerDB instance
	client := Client{}
	db, err := badger.Open(mergedOpts)
	if err != nil {
		return nil, errors.New("cannot open BadgerDB")
	}
	client.db = db

	// add additionalOptions
	a := getDefaultAdditionalOptions()
	err = mergo.Merge(&a, additionalOptions)
	if err != nil {
		return nil, fmt.Errorf("cannot merge additionalOptions with defaultOptions for BadgerDB: %e", err)
	}
	client.additionalOptions = a

	return &client, nil
}

func getDefaultAdditionalOptions() AdditionalOptions {
	return AdditionalOptions{
		TTL: time.Hour * 24,
	}
}

func validateKeyToken(s string) error {
	if strings.Contains(s, ":") {
		return errors.New("key token contains \":\"")
	}

	return nil
}

func getMetricPrefix(applicationName string, version int, signature string) string {
	return fmt.Sprintf("kt-metric::%s::%d::%s::", applicationName, version, signature)
}

func getMetricKey(applicationName string, version int, signature, metric, user, transaction string) (string, error) {
	if err := validateKeyToken(applicationName); err != nil {
		return "", errors.New("application name cannot have \":\"")
	}
	if err := validateKeyToken(signature); err != nil {
		return "", errors.New("signature cannot have \":\"")
	}
	if err := validateKeyToken(metric); err != nil {
		return "", errors.New("metric name cannot have \":\"")
	}
	if err := validateKeyToken(user); err != nil {
		return "", errors.New("user name cannot have \":\"")
	}
	if err := validateKeyToken(transaction); err != nil {
		return "", errors.New("transaction ID cannot have \":\"")
	}

	return fmt.Sprintf("%s%s::%s::%s", getMetricPrefix(applicationName, version, signature), metric, user, transaction), nil
}

// SetMetric sets a metric based on the app name, version, signature, metric type, user name, transaction ID, and metric value with BadgerDB
// Example key: kt-metric::my-app::0::my-signature::my-metric::my-user::my-transaction-id -> my-metric-value
func (cl Client) SetMetric(applicationName string, version int, signature, metric, user, transaction string, metricValue float64) error {
	key, err := getMetricKey(applicationName, version, signature, metric, user, transaction)
	if err != nil {
		return err
	}

	err = cl.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), []byte(fmt.Sprintf("%f", metricValue))).WithTTL(cl.additionalOptions.TTL)
		err := txn.SetEntry(e)
		return err
	})
	if err != nil {
		return fmt.Errorf("cannot set metric with key \"%s\": %w", key, err)
	}

	// update user
	err = cl.SetUser(applicationName, version, signature, user)

	return err
}

func getUserPrefix(applicationName string, version int, signature string) string {
	return fmt.Sprintf("kt-users::%s::%d::%s::", applicationName, version, signature)
}

func getUserKey(applicationName string, version int, signature, user string) string {
	// getUserKey() is just getUserPrefix() with the user appended at the end
	return fmt.Sprintf("%s%s", getUserPrefix(applicationName, version, signature), user)
}

// SetUser sets a user based on the app name, version, signature, and user name with BadgerDB
// Example key/value: kt-users::my-app::0::my-signature::my-user -> true
func (cl Client) SetUser(applicationName string, version int, signature, user string) error {
	key := getUserKey(applicationName, version, signature, user)

	return cl.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), []byte("true")).WithTTL(cl.additionalOptions.TTL)
		err := txn.SetEntry(e)
		return err
	})
}

// getUserCount gets the number of users
func (cl Client) getUserCount(applicationName string, version int, signature string) (uint64, error) {
	count := uint64(0)

	err := cl.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(getUserPrefix(applicationName, version, signature))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			count++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetSummaryMetrics gets summarized metrics for a particular application, version, and signature
func (cl Client) GetSummaryMetrics(applicationName string, version int, signature string) (*storageclient.VersionMetricSummary, error) {
	count, err := cl.getUserCount(applicationName, version, signature)
	if err != nil {
		return nil, fmt.Errorf("cannot get user count: %e", err)
	}

	vms := storageclient.VersionMetricSummary{
		NumUsers:        count,
		MetricSummaries: map[string]storageclient.MetricSummary{},
	}

	// get metric summaries
	// iter8 over all metrics with the provided application name, version, and signature
	err = cl.db.View(func(txn *badger.Txn) error {
		// used to capture all the metric values for the current metric
		var overUserData []float64        // cumulative metric value for a particular user
		var overTransactionData []float64 // all metric values

		// used to determine what the previous user and metric are in the previous iteration
		var previousUser string
		var previousMetric string

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(getMetricPrefix(applicationName, version, signature))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			// get key
			key := string(item.Key())
			err := item.Value(func(v []byte) error {
				// get value
				val := string(v)
				fval, err := strconv.ParseFloat(val, 64)
				if err != nil {
					return fmt.Errorf("cannot convert metric value \"%s\" into float64", val)
				}

				// break down key into tokens
				tokens := strings.Split(key[len(prefix):], "::")
				if len(tokens) != 3 {
					return fmt.Errorf("incorrect number of tokens in metric key \"%s\"; expected 7 tokens", key)
				}

				currentMetric := tokens[0]
				currentUser := tokens[1]

				// initial case
				if previousMetric == "" {
					previousMetric = currentMetric
					previousUser = currentUser

					overTransactionData = []float64{fval}
					overUserData = []float64{fval}

					return nil
				}

				// calculate overUserData
				if currentUser == previousUser {
					overUserData[len(overUserData)-1] += fval // cumulative metric value for a particular user
				} else {
					previousUser = currentUser
					overUserData = append(overUserData, fval) // add new entry when the user changes
				}

				// calculate overTransactionData or calculate summarizedMetric
				if currentMetric == previousMetric {
					overTransactionData = append(overTransactionData, fval) // always add new entry
				} else {
					summaryOverTransactions, err := storageclient.CalculateSummarizedMetric(overTransactionData)
					if err != nil {
						return fmt.Errorf("cannot calculate summarized metric: %e", err)
					}
					summaryOverUsers, err := storageclient.CalculateSummarizedMetric(overUserData)
					if err != nil {
						return fmt.Errorf("cannot calculate summarized metric: %e", err)
					}
					vms.MetricSummaries[currentMetric] = storageclient.MetricSummary{
						SummaryOverTransactions: summaryOverTransactions,
						SummaryOverUsers:        summaryOverUsers,
					}

					previousUser = currentUser
					previousMetric = currentMetric

					// reset data
					overUserData = []float64{fval}
					overTransactionData = []float64{fval}
				}

				return nil
			})

			if err != nil {
				return err
			}
		}

		summaryOverTransactions, err := storageclient.CalculateSummarizedMetric(overTransactionData)
		if err != nil {
			return fmt.Errorf("cannot calculate summarized metric: %e", err)
		}
		summaryOverUsers, err := storageclient.CalculateSummarizedMetric(overUserData)
		if err != nil {
			return fmt.Errorf("cannot calculate summarized metric: %e", err)
		}

		// flush last sequence of metric data
		vms.MetricSummaries[previousMetric] = storageclient.MetricSummary{
			SummaryOverTransactions: summaryOverTransactions,
			SummaryOverUsers:        summaryOverUsers,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &vms, nil
}

// GetMetrics returns a nested map of the metrics data for a particular application, version, and signature
// Example:
//
//	{
//		"my-metric": {
//			"MetricsOverTransactions": [1, 1, 3, 4, 5]
//			"MetricsOverUsers": [2, 7, 5]
//		}
//	}
func (cl Client) GetMetrics(applicationName string, version int, signature string) (*storageclient.VersionMetrics, error) {
	metrics := storageclient.VersionMetrics{}

	err := cl.db.View(func(txn *badger.Txn) error {
		// used to determine what the previous user and metric are in the previous iteration
		var currentMetric string
		var currentUser string

		var cumulativeUserValue float64

		var metricsOverTransactions []float64
		var metricsOverUsers []float64

		// iterate over all metrics of a particular application name, version, and signature
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(getMetricPrefix(applicationName, version, signature))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := string(item.Key())

			// extract metric and user name from the key
			tokens := strings.Split(key, "::")
			if len(tokens) != 7 {
				return fmt.Errorf("incorrect number of tokens in metrics key: \"%s\": should be 7 (example: kt-metric::my-app::0::my-signature::my-metric::my-user::my-transaction-id)", key)
			}
			metric := tokens[4]
			user := tokens[5]

			err := item.Value(func(v []byte) error {
				floatValue, err := strconv.ParseFloat(string(v), 64)
				if err != nil {
					return err
				}

				if metric != currentMetric && currentMetric != "" {
					metrics[currentMetric] = struct {
						MetricsOverTransactions []float64
						MetricsOverUsers        []float64
					}{
						MetricsOverTransactions: metricsOverTransactions,
						MetricsOverUsers:        metricsOverUsers,
					}

					currentMetric = ""
					currentUser = ""
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

				return nil
			})
			if err != nil {
				return err
			}

		}

		// flush last sequence of metric data
		if currentMetric != "" || currentUser != "" {
			metricsOverUsers = append(metricsOverUsers, cumulativeUserValue)
			metrics[currentMetric] = struct {
				MetricsOverTransactions []float64
				MetricsOverUsers        []float64
			}{
				MetricsOverTransactions: metricsOverTransactions,
				MetricsOverUsers:        metricsOverUsers,
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &metrics, nil
}
