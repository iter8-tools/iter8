// Package badgerdb provides a client for BadgerDB
package badgerdb

import (
	"errors"
	"os"

	"github.com/dgraph-io/badger"
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

// Size gets the current size and the maximum size of the BadgerDB
func (cl Client) Size() (int64, int64, error) {
	lsm, vlog := cl.db.Size()

	// TODO: get total storage of DB
	return lsm + vlog, -1, nil
}
