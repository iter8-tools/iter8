package badgerdb

import (
	"errors"
	"os"

	"github.com/dgraph-io/badger"
	"github.com/imdario/mergo"
)

type Client struct {
	db *badger.DB
}

func GetClient(opts badger.Options) (*Client, error) {
	dir := opts.Dir
	valueDir := opts.ValueDir

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

	client := Client{}

	// initialize BadgerDB instance
	db, err := badger.Open(mergedOpts)
	if err != nil {
		return nil, errors.New("cannot open BadgerDB")
	}

	client.db = db

	return &client, nil
}
