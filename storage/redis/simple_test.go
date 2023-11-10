package redis

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/storage"
	"github.com/stretchr/testify/assert"
)

func TestSetMetric(t *testing.T) {
	server, _ := miniredis.Run()
	assert.NotNil(t, server)

	client, err := GetClient(ClientConfig{Address: base.StringPointer(server.Addr())})
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

	key, err := storage.GetMetricKey(app, version, signature, metric, user, transaction)
	assert.NoError(t, err)
	val, err := client.rdb.Get(context.Background(), key).Result()
	assert.NoError(t, err)
	fval, err := strconv.ParseFloat(string(val), 64)
	assert.NoError(t, err)

	assert.Equal(t, value, fval)

	// SetMetric() should also add a user
	userKey := storage.GetUserKey(app, version, signature, user)
	u, err := client.rdb.Get(context.Background(), userKey).Result()
	assert.NoError(t, err)
	assert.Equal(t, "true", u)
}

func TestSetMetricInvalid(t *testing.T) {
	server, _ := miniredis.Run()
	assert.NotNil(t, server)

	client, err := GetClient(ClientConfig{Address: base.StringPointer(server.Addr())})
	assert.NoError(t, err)

	err = client.SetMetric("invalid:application", 0, "signature", "metric", "user", "transaction", float64(0))
	assert.Error(t, err)
}

func TestSetUser(t *testing.T) {
	server, _ := miniredis.Run()
	assert.NotNil(t, server)

	client, err := GetClient(ClientConfig{Address: base.StringPointer(server.Addr())})
	assert.NoError(t, err)

	app := "my-application"
	version := 0
	signature := "my-signature"
	user := "my-user"

	err = client.SetUser(app, version, signature, user)
	assert.NoError(t, err)

	userKey := storage.GetUserKey(app, version, signature, user)
	u, err := client.rdb.Get(context.Background(), userKey).Result()
	assert.NoError(t, err)
	assert.Equal(t, "true", u)
}

// TestGetMetricsWithExtraUsers tests if GetMetrics adds 0 for all users that did not produce metrics
func TestGetMetricsWithExtraUsers(t *testing.T) {
	server, _ := miniredis.Run()
	assert.NotNil(t, server)

	client, err := GetClient(ClientConfig{Address: base.StringPointer(server.Addr())})
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

func TestGetMetrics(t *testing.T) {
	server, _ := miniredis.Run()
	assert.NotNil(t, server)

	client, err := GetClient(ClientConfig{Address: base.StringPointer(server.Addr())})
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

func TestGetExperimentResult(t *testing.T) {
	server, _ := miniredis.Run()
	assert.NotNil(t, server)

	client, err := GetClient(ClientConfig{Address: base.StringPointer(server.Addr())})
	assert.NoError(t, err)

	namespace := "my-namespace"
	experiment := "my-experiment"

	experimentResult := base.ExperimentResult{
		Name:      experiment,
		Namespace: namespace,
	}

	err = client.SetExperimentResult(namespace, experiment, &experimentResult)
	assert.NoError(t, err)

	result, err := client.GetExperimentResult(namespace, experiment)
	assert.NoError(t, err)
	assert.Equal(t, &experimentResult, result)
}
