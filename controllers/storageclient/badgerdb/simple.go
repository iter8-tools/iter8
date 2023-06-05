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
)

// const (
// 	builtInUserCountID = "user-count"
// )

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

func getValueFromBadgerDB(db *badger.DB, key string) ([]byte, error) {
	var valCopy []byte

	err := db.View(func(txn *badger.Txn) error {
		// query for key/value
		item, err := txn.Get([]byte(key))
		if err != nil {
			return fmt.Errorf("cannot get signature with key \"%s\": %w", key, err)
		}

		// copy value
		err = item.Value(func(val []byte) error {
			// Copying or parsing val is valid.
			valCopy = append([]byte{}, val...)

			return nil
		})

		return err
	})

	if err != nil {
		return []byte{}, err
	}

	return valCopy, nil
}

func validateKeyToken(s string) error {
	if strings.Contains(s, ":") {
		return errors.New("key token contains \":\"")
	}

	return nil
}

func getMetricKey(applicationName string, version int, signature, metricType, user, transaction string) (string, error) {
	if err := validateKeyToken(applicationName); err != nil {
		return "", errors.New("application name cannot have \":\"")
	}
	if err := validateKeyToken(signature); err != nil {
		return "", errors.New("signature cannot have \":\"")
	}
	if err := validateKeyToken(metricType); err != nil {
		return "", errors.New("metric type cannot have \":\"")
	}
	if err := validateKeyToken(user); err != nil {
		return "", errors.New("user name cannot have \":\"")
	}
	if err := validateKeyToken(transaction); err != nil {
		return "", errors.New("transaction ID cannot have \":\"")
	}

	return fmt.Sprintf("kt-metric::%s::%d::%s::%s::%s::%s", applicationName, version, signature, metricType, user, transaction), nil
}

// GetMetric gets a metric based on the app name, version, signature, metric type, user, and transaction ID from BadgerDB
// Example key/value: kt-metric::my-app::0::my-signature::my-metric-type::my-user::my-transaction-id -> my-metric-value
func (cl Client) GetMetric(applicationName string, version int, signature, metricType, user, transaction string) (float64, error) {
	key, err := getMetricKey(applicationName, version, signature, metricType, user, transaction)
	if err != nil {
		return 0, err
	}

	val, err := getValueFromBadgerDB(cl.db, key)
	if err != nil {
		return 0, err
	}

	f, err := strconv.ParseFloat(string(val), 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse metric into float64 with key \"%s\": %w", key, err)
	}

	return f, nil
}

// SetMetric sets a metric based on the app name, version, signature, metric type, user name, transaction ID, and metric value with BadgerDB
func (cl Client) SetMetric(applicationName string, version int, signature, metricType, user, transaction string, metricValue float64) error {
	key, err := getMetricKey(applicationName, version, signature, metricType, user, transaction)
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

	// update metrics
	err = cl.SetMetricType(applicationName, metricType)
	if err != nil {
		return err
	}

	// update user
	err = cl.SetUser(applicationName, version, signature, user)

	return err
}

func getMetricTypeKey(applicationName, metricType string) string {
	return fmt.Sprintf("kt-metric-types::%s::%s", applicationName, metricType)
}

// SetMetricType sets a metric based on the app name and metric type with BadgerDB
// Example key/value: kt-metric-types::my-app::my-metric -> true
func (cl Client) SetMetricType(applicationName, metricType string) error {
	key := getMetricTypeKey(applicationName, metricType)

	return cl.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), []byte("true")).WithTTL(cl.additionalOptions.TTL)
		err := txn.SetEntry(e)
		return err
	})
}

// func (cl Client) GetMetrics(applicationName string) ([]string, error) {
// 	return []string{}, nil
// }

func getUserKey(applicationName string, version int, signature, user string) string {
	return fmt.Sprintf("kt-metric::%s::%d::%s::%s", applicationName, version, signature, user)
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

// func (cl Client) GetUsers(applicationName string, version int, signature string) ([]string, error) {
// 	return []string{}, nil
// }

// // GetSummaryMetrics gets a summary of all the metrics from all versions of an application
// func (cl Client) GetSummaryMetrics(applicationName string) (*map[int]storageclient.VersionMetricSummary, error) {
// 	metrics := map[string]float64{}

// 	// prefix scan of metrics using applicationName
// 	err := cl.db.View(func(txn *badger.Txn) error {
// 		it := txn.NewIterator(badger.DefaultIteratorOptions)
// 		defer it.Close()
// 		prefix := []byte(fmt.Sprintf("kt-metric::%s", applicationName))
// 		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
// 			item := it.Item()
// 			key := item.Key()

// 			// save data
// 			err := item.Value(func(val []byte) error {
// 				fmt.Printf("key=%s, value=%s\n", key, val)

// 				fval, err := strconv.ParseFloat(string(val), 64)
// 				if err != nil {
// 					return fmt.Errorf("cannot parse float from metric \"%s\": \"%s\": %w", key, string(val), err)
// 				}

// 				metrics[string(key)] = fval
// 				return nil
// 			})
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	// prefix scan of users using applicationName
// 	users := []string{}
// 	err = cl.db.View(func(txn *badger.Txn) error {
// 		it := txn.NewIterator(badger.DefaultIteratorOptions)
// 		defer it.Close()
// 		prefix := []byte(fmt.Sprintf("kt-users::%s", applicationName))
// 		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
// 			// save data
// 			users = append(users, string(it.Item().Key()))
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	// loop through metrics and aggregate data for result
// 	result := map[int]storageclient.VersionMetricSummary{}
// 	for key := range metrics {
// 		s := storageclient.VersionMetricSummary{}

// 		// check if the number of tokens is correct (7)
// 		tokens := strings.Split(key, "::")
// 		if len(tokens) != 7 {
// 			return nil, fmt.Errorf("incorrect number of tokens in metric key: \"%s\": %w", key, err)
// 		}
// 		version := tokens[2]

// 		// convert version to integer
// 		iversion, err := strconv.Atoi(version)
// 		if err != nil {
// 			return nil, fmt.Errorf("cannot parse version number from metric key \"%s\" into integer: %w", key, err)
// 		}

// 		// TODO: compute summary

// 		result[iversion] = s
// 	}
// 	if err != nil {
// 		return nil, err
// 	}

// 	// loop through users and add user count for result
// 	for _, user := range users {
// 		// check if the number of tokens is correct (5)
// 		tokens := strings.Split(user, "::")
// 		if len(tokens) != 5 {
// 			return nil, fmt.Errorf("incorrect number of tokens in user key: \"%s\": %w", user, err)
// 		}
// 		version := tokens[2]

// 		// convert version to integer
// 		iversion, err := strconv.Atoi(version)
// 		if err != nil {
// 			return nil, fmt.Errorf("cannot parse version number from user key \"%s\" into integer: %w", version, err)
// 		}

// 		// TODO: increment userCount
// 		x := result[iversion][builtInUserCountID]
// 		x.Add(1)
// 	}
// 	if err != nil {
// 		return nil, err
// 	}

// 	// validate result
// 	// loop through result to ensure that each summary has a user count
// 	for version, versionMetricSummary := range result {
// 		_, ok := versionMetricSummary[builtInUserCountID]

// 		if !ok {
// 			return nil, fmt.Errorf("summary with version number \"%d\" does not contain user count: %w", version, err)
// 		}
// 	}

// 	return &result, nil
// }
