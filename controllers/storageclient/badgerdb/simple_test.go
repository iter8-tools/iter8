package badgerdb

import (
	"strconv"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetClient(t *testing.T) {
	tempDirPath := t.TempDir()

	client, err := GetClient(badger.DefaultOptions(tempDirPath), AdditionalOptions{})
	assert.NoError(t, err)

	assert.NotNil(t, client)
	assert.NotNil(t, client.db) // BadgerDB should exist

	err = client.db.Close()
	assert.NoError(t, err)
}

func TestSetMetric(t *testing.T) {
	tempDirPath := t.TempDir()

	client, err := GetClient(badger.DefaultOptions(tempDirPath), AdditionalOptions{})
	assert.NoError(t, err)

	app := "my-application"
	version := 0
	signature := "my-signature"
	metricType := "my-metric-type"
	user := "my-user"
	transaction := "my-transaction"
	value := 50.0

	err = client.SetMetric(app, version, signature, metricType, user, transaction, value)
	assert.NoError(t, err)

	// get metric
	err = client.db.View(func(txn *badger.Txn) error {
		key, err := getMetricKey(app, version, signature, metricType, user, transaction)
		assert.NoError(t, err)

		item, err := txn.Get([]byte(key))
		assert.NoError(t, err)
		assert.NotNil(t, item)

		item.Value(func(val []byte) error {
			// parse val into float64
			fval, err := strconv.ParseFloat(string(val), 64)
			assert.NoError(t, err)

			// assert metric value is the same as the provided one
			assert.Equal(t, value, fval)
			return nil
		})

		return nil
	})
	assert.NoError(t, err)

	// SetMetric() should also add a metric type
	err = client.db.View(func(txn *badger.Txn) error {
		key := getMetricTypeKey(app, metricType)
		item, err := txn.Get([]byte(key))
		assert.NoError(t, err)
		assert.NotNil(t, item)

		item.Value(func(val []byte) error {
			// metric type should be set to "true"
			assert.Equal(t, "true", string(val))
			return nil
		})

		return nil
	})
	assert.NoError(t, err)

	// SetMetric() should also add a user
	err = client.db.View(func(txn *badger.Txn) error {
		key := getUserKey(app, version, signature, user)
		item, err := txn.Get([]byte(key))
		assert.NoError(t, err)
		assert.NotNil(t, item)

		item.Value(func(val []byte) error {
			// user should be set to "true"
			assert.Equal(t, "true", string(val))
			return nil
		})

		return nil
	})
	assert.NoError(t, err)
}

func TestGetMetric(t *testing.T) {
	tempDirPath := t.TempDir()

	client, err := GetClient(badger.DefaultOptions(tempDirPath), AdditionalOptions{})
	assert.NoError(t, err)

	app := "my-application"
	version := 0
	signature := "my-signature"
	metricType := "my-metric-type"
	user := "my-user"
	transaction := "my-transaction"
	value := 50.0

	err = client.SetMetric(app, version, signature, metricType, user, transaction, value)
	assert.NoError(t, err)

	// get metric
	dbval, err := client.GetMetric(app, version, signature, metricType, user, transaction)
	assert.NoError(t, err)
	assert.Equal(t, dbval, value)
}

func TestSetMetricType(t *testing.T) {
	tempDirPath := t.TempDir()

	client, err := GetClient(badger.DefaultOptions(tempDirPath), AdditionalOptions{})
	assert.NoError(t, err)

	app := "my-application"
	metricType := "my-metric-type"

	err = client.SetMetricType(app, metricType)
	assert.NoError(t, err)

	// get metric type
	err = client.db.View(func(txn *badger.Txn) error {
		key := getMetricTypeKey(app, metricType)
		item, err := txn.Get([]byte(key))
		assert.NoError(t, err)
		assert.NotNil(t, item)

		item.Value(func(val []byte) error {
			// metric type should be set to "true"
			assert.Equal(t, "true", string(val))
			return nil
		})

		return nil
	})
	assert.NoError(t, err)
}

func TestSetUser(t *testing.T) {
	tempDirPath := t.TempDir()

	client, err := GetClient(badger.DefaultOptions(tempDirPath), AdditionalOptions{})
	assert.NoError(t, err)

	app := "my-application"
	version := 0
	signature := "my-signature"
	user := "my-user"

	err = client.SetUser(app, version, signature, user)
	assert.NoError(t, err)

	// get user
	err = client.db.View(func(txn *badger.Txn) error {
		key := getUserKey(app, version, signature, user)
		item, err := txn.Get([]byte(key))
		assert.NoError(t, err)
		assert.NotNil(t, item)

		item.Value(func(val []byte) error {
			// metric type should be set to "true"
			assert.Equal(t, "true", string(val))
			return nil
		})

		return nil
	})
	assert.NoError(t, err)
}

func TestValidateKeyToken(t *testing.T) {
	err := validateKeyToken("hello")
	assert.NoError(t, err)

	err = validateKeyToken("::")
	assert.Error(t, err)

	err = validateKeyToken("hello::world")
	assert.Error(t, err)

	err = validateKeyToken("hello :: world")
	assert.Error(t, err)

	err = validateKeyToken("hello:world")
	assert.Error(t, err)

	err = validateKeyToken("hello : world")
	assert.Error(t, err)
}
