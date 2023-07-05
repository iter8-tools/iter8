package badgerdb

import (
	"encoding/json"
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
	metric := "my-metric"
	user := "my-user"
	transaction := "my-transaction"
	value := 50.0

	err = client.SetMetric(app, version, signature, metric, user, transaction, value)
	assert.NoError(t, err)

	// get metric
	err = client.db.View(func(txn *badger.Txn) error {
		key, err := getMetricKey(app, version, signature, metric, user, transaction)
		assert.NoError(t, err)

		item, err := txn.Get([]byte(key))
		assert.NoError(t, err)
		assert.NotNil(t, item)

		err = item.Value(func(val []byte) error {
			// parse val into float64
			fval, err := strconv.ParseFloat(string(val), 64)
			assert.NoError(t, err)

			// assert metric value is the same as the provided one
			assert.Equal(t, value, fval)
			return nil
		})
		assert.NoError(t, err)

		return nil
	})
	assert.NoError(t, err)

	// SetMetric() should also add a user
	err = client.db.View(func(txn *badger.Txn) error {
		key := getUserKey(app, version, signature, user)
		item, err := txn.Get([]byte(key))
		assert.NoError(t, err)
		assert.NotNil(t, item)

		err = item.Value(func(val []byte) error {
			// user should be set to "true"
			assert.Equal(t, "true", string(val))
			return nil
		})
		assert.NoError(t, err)

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

		err = item.Value(func(val []byte) error {
			// metric type should be set to "true"
			assert.Equal(t, "true", string(val))
			return nil
		})
		assert.NoError(t, err)

		return nil
	})
	assert.NoError(t, err)
}

// TestGetMetricsWithExtraUsers tests if GetMetrics adds 0 for all users that did not produce metrics
func TestGetMetricsWithExtraUsers(t *testing.T) {
	tempDirPath := t.TempDir()

	client, err := GetClient(badger.DefaultOptions(tempDirPath), AdditionalOptions{})
	assert.NoError(t, err)

	app := "my-application"
	version := 0
	signature := "my-signature"
	extraUser := "my-extra-user"

	err = client.SetUser(app, version, signature, extraUser) // extra user
	assert.NoError(t, err)

	metric := "my-metric"
	user := "my-user"
	transaction := "my-transaction"

	err = client.SetMetric(app, version, signature, metric, user, transaction, 25)
	assert.NoError(t, err)

	metric2 := "my-metric2"

	err = client.SetMetric(app, version, signature, metric2, user, transaction, 50)
	assert.NoError(t, err)

	metrics, err := client.GetMetrics(app, version, signature)
	assert.NoError(t, err)

	jsonMetrics, err := json.Marshal(metrics)
	assert.NoError(t, err)
	// 0s have been added to the MetricsOverUsers due to extraUser, [50,0]
	assert.Equal(t, "{\"my-metric\":{\"MetricsOverTransactions\":[25],\"MetricsOverUsers\":[25,10]},\"my-metric2\":{\"MetricsOverTransactions\":[50],\"MetricsOverUsers\":[50,0]}}", string(jsonMetrics))
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

func TestGetMetrics(t *testing.T) {
	tempDirPath := t.TempDir()

	client, err := GetClient(badger.DefaultOptions(tempDirPath), AdditionalOptions{})
	assert.NoError(t, err)

	err = client.SetMetric("my-application", 0, "my-signature", "my-metric", "my-user", "my-transaction", 50.0)
	assert.NoError(t, err)
	err = client.SetMetric("my-application", 0, "my-signature", "my-metric", "my-user2", "my-transaction2", 10.0)
	assert.NoError(t, err)
	err = client.SetMetric("my-application", 1, "my-signature2", "my-metric2", "my-user", "my-transaction3", 20.0)
	assert.NoError(t, err)
	err = client.SetMetric("my-application", 2, "my-signature3", "my-metric3", "my-user2", "my-transaction4", 30.0)
	assert.NoError(t, err)
	err = client.SetMetric("my-application", 2, "my-signature3", "my-metric3", "my-user2", "my-transaction4", 40.0) // overwrites the previous set
	assert.NoError(t, err)

	metrics, err := client.GetMetrics("my-application", 0, "my-signature")
	assert.NoError(t, err)
	jsonMetrics, err := json.Marshal(metrics)
	assert.NoError(t, err)
	assert.Equal(t, "{\"my-metric\":{\"MetricsOverTransactions\":[10,50],\"MetricsOverUsers\":[10,50]}}", string(jsonMetrics))

	metrics, err = client.GetMetrics("my-application", 1, "my-signature2")
	assert.NoError(t, err)
	jsonMetrics, err = json.Marshal(metrics)
	assert.NoError(t, err)
	assert.Equal(t, "{\"my-metric2\":{\"MetricsOverTransactions\":[20],\"MetricsOverUsers\":[20]}}", string(jsonMetrics))

	metrics, err = client.GetMetrics("my-application", 2, "my-signature3")
	assert.NoError(t, err)
	jsonMetrics, err = json.Marshal(metrics)
	assert.NoError(t, err)
	assert.Equal(t, "{\"my-metric3\":{\"MetricsOverTransactions\":[40],\"MetricsOverUsers\":[40]}}", string(jsonMetrics))

	metrics, err = client.GetMetrics("my-application", 3, "my-signature")
	assert.NoError(t, err)
	jsonMetrics, err = json.Marshal(metrics)
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(jsonMetrics))
}
