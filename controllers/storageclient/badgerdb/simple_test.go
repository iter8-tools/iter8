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

func TestGetSummaryMetrics(t *testing.T) {
	// MetricName: sales
	// user u1; transaction t1: 30
	// user u1; transaction t2: 20
	// user u1; transaction t3: 20

	// user u2; transaction t4: 50

	// user u3: transaction t5: 0.5
	// user u3: transaction t6: 1.5

	// summaryOverTransactions = {
	// 	Count: 6
	// 	Mean: (30 + 20 + 20 + 50 + 0.5 + 1.5)/6
	// 	StdDev: ...
	// 	Min: 0.5
	// 	Max: 50
	// }

	// // metric values are added for a given user
	// summaryOverUsers = {
	// 	Count: 3
	// 	Mean: (30 + 20 + 20 + 50 + 0.5 + 1.5)/3
	// 	StdDev: ...
	// 	Min: 2.0
	// 	Max: 70
	// }

	tempDirPath := t.TempDir()

	client, err := GetClient(badger.DefaultOptions(tempDirPath), AdditionalOptions{})
	assert.NoError(t, err)

	app := "my-application"
	version := 0
	signature := "my-signature"
	metric := "sales"
	user1 := "u1"
	user2 := "u2"
	user3 := "u3"

	err = client.SetMetric(app, version, signature, metric, user1, "t1", 30)
	assert.NoError(t, err)
	err = client.SetMetric(app, version, signature, metric, user1, "t2", 20)
	assert.NoError(t, err)
	err = client.SetMetric(app, version, signature, metric, user1, "t3", 20)
	assert.NoError(t, err)
	err = client.SetMetric(app, version, signature, metric, user2, "t4", 50)
	assert.NoError(t, err)
	err = client.SetMetric(app, version, signature, metric, user3, "t5", 0.5)
	assert.NoError(t, err)
	err = client.SetMetric(app, version, signature, metric, user3, "t6", 1.5)
	assert.NoError(t, err)

	vms, err := client.GetSummaryMetrics(app, version, signature)
	assert.NoError(t, err)

	assert.Equal(t, uint64(3), vms.NumUsers)

	assert.Equal(t, uint64(6), vms.MetricSummaries[metric].SummaryOverTransactions.Count)
	assert.Equal(t, 20.333333333333332, vms.MetricSummaries[metric].SummaryOverTransactions.Mean)
	assert.Equal(t, 16.940254491070146, vms.MetricSummaries[metric].SummaryOverTransactions.StdDev)
	assert.Equal(t, 0.5, vms.MetricSummaries[metric].SummaryOverTransactions.Min)
	assert.Equal(t, 50.0, vms.MetricSummaries[metric].SummaryOverTransactions.Max)

	assert.Equal(t, uint64(3), vms.MetricSummaries[metric].SummaryOverUsers.Count)
	assert.Equal(t, 40.666666666666664, vms.MetricSummaries[metric].SummaryOverUsers.Mean)
	assert.Equal(t, 28.534579412043595, vms.MetricSummaries[metric].SummaryOverUsers.StdDev)
	assert.Equal(t, 2.0, vms.MetricSummaries[metric].SummaryOverUsers.Min)
	assert.Equal(t, 70.0, vms.MetricSummaries[metric].SummaryOverUsers.Max)
}
