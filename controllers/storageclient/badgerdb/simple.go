// Package badgerdb implements the storageclient interface with BadgerDB
package badgerdb

import (
	"errors"
	"fmt"
	"os"
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

	return fmt.Sprintf("kt-metric::%s::%d::%s::%s::%s::%s", applicationName, version, signature, metric, user, transaction), nil
}

// SetMetric sets a metric based on the app name, version, signature, metric type, user name, transaction ID, and metric value with BadgerDB
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

// GetSummaryMetrics gets a summary of all the metrics from all versions of an application
func (cl Client) GetSummaryMetrics(applicationName string) (*map[int]storageclient.VersionMetricSummary, error) {
	return nil, nil
}
