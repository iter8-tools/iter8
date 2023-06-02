// Package badgerdb implements the storageclient interface with BadgerDB
package badgerdb

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/dgraph-io/badger/v4"
	"github.com/imdario/mergo"
)

// Client is a client for the BadgerDB
type Client struct {
	db *badger.DB
}

// GetClient gets a client for the BadgerDB
func GetClient(opts badger.Options) (*Client, error) {
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

	return &client, nil
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
		item.Value(func(val []byte) error {
			// Copying or parsing val is valid.
			valCopy = append([]byte{}, val...)

			return nil
		})

		return nil
	})

	if err != nil {
		return []byte{}, err
	}

	return valCopy, nil
}

func getMetricKey(applicationName string, version int, signature, metric, user, transaction string) string {
	return fmt.Sprintf("kt-metric::%s::%d::%s::%s::%s::%s", applicationName, version, signature, metric, user, transaction)
}

// Key 2: kt-metric::my-app::0::my-signature::my-metric::my-user::my-transaction-id -> metric-value
func (cl Client) GetMetric(applicationName string, version int, signature, metric, user, transaction string) (float64, error) {
	key := getMetricKey(applicationName, version, signature, metric, user, transaction)
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

func (cl Client) SetMetric(applicationName string, version int, signature, metric, user, transaction string, metricValue float64) error {
	key := getMetricKey(applicationName, version, signature, metric, user, transaction)

	err := cl.db.Update(func(txn *badger.Txn) error {
		// set TTL
		e := badger.NewEntry([]byte(key), []byte(fmt.Sprintf("%f", metricValue)))
		err := txn.SetEntry(e)
		return err
	})

	// update metrics?
	// update user?

	return err
}

func getMetricsKey(applicationName, metric string) string {
	return fmt.Sprintf("kt-app-metrics::%s::%s", applicationName, metric)
}

// Key 3: kt-app-metrics::my-app::my-metric -> true
func (cl Client) SetMetrics(applicationName, metric string) error {
	key := getMetricsKey(applicationName, metric)

	return cl.db.Update(func(txn *badger.Txn) error {
		// set TTL
		e := badger.NewEntry([]byte(key), []byte("true"))
		err := txn.SetEntry(e)
		return err
	})
}

// func (cl Client) GetMetrics(applicationName string) ([]string, error) {
// 	return []string{}, nil
// }

func getUsersKey(applicationName string, version int, signature, user string) string {
	return fmt.Sprintf("kt-metric::%s::%d::%s::%s", applicationName, version, signature, user)
}

// Key 5: kt-users::my-app::0::my-signature::my-user -> true
func (cl Client) SetUsers(applicationName string, version int, signature, user string) error {
	key := getUsersKey(applicationName, version, signature, user)

	return cl.db.Update(func(txn *badger.Txn) error {
		// set TTL
		e := badger.NewEntry([]byte(key), []byte("true"))
		err := txn.SetEntry(e)
		return err
	})
}

// func (cl Client) GetUsers(applicationName string, version int, signature, user string) ([]string, error) {
// 	return []string{}, nil
// }
