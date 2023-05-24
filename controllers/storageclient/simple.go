package storageclient

import (
	"errors"
	"fmt"
	"os"

	"github.com/dgraph-io/badger"
	"github.com/imdario/mergo"
)

type Client struct {
	db interface{}
}

// Start initializes BadgerDB instance
func (cl *Client) Start(path string, opts interface{}) error {
	// Check if path exists (if volume has been mounted)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("does not exist")
		return errors.New("path does not exist; volume has not been mounted")
	}

	// cast interface{} into BadgerDB options
	dbOpts, ok := opts.(badger.Options)
	if !ok {
		return errors.New("cannot cast opts into BadgerDB options")
	}

	// set default options
	mergedOpts := badger.DefaultOptions(path)
	if err := mergo.Merge(&mergedOpts, dbOpts); err != nil {
		return errors.New("cannot configure default options for BadgerDB")
	}

	// set directory and value directory to path
	mergedOpts.Dir = path
	mergedOpts.ValueDir = path

	// intialize BadgerDB instance
	db, err := badger.Open(mergedOpts)
	if err != nil {
		return errors.New("cannot open BadgerDB")
	}

	cl.db = db
	return nil
}

func (cl *Client) CreateMetric(name string, value interface{}) error {
	db, ok := cl.db.(*badger.DB)
	if !ok {
		return errors.New("cannot use BadgerDB")
	}

	txn := db.NewTransaction(true)
	defer txn.Discard()

	// TODO: How to convert value to byte array?
	err := txn.Set([]byte(name), []byte(fmt.Sprintf("%v", value)))
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (cl *Client) ReadMetric(name string) (interface{}, error) {
	db, ok := cl.db.(*badger.DB)
	if !ok {
		return nil, errors.New("cannot use BadgerDB")
	}

	txn := db.NewTransaction(true)
	defer txn.Discard()

	item, err := txn.Get([]byte(name))
	if err != nil {
		return nil, err
	}

	return item.ValueCopy(nil)
}

func (cl *Client) DeleteMetric(name string, value interface{}) error {
	db, ok := cl.db.(*badger.DB)
	if !ok {
		return errors.New("cannot use BadgerDB")
	}

	txn := db.NewTransaction(true)
	defer txn.Discard()

	err := txn.Delete([]byte(name))
	if err != nil {
		return err
	}

	// TODO: is Commit() necessary here?
	return txn.Commit()
}
