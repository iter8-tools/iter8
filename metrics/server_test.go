package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/iter8-tools/iter8/abn"
	util "github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/controllers"
	"github.com/iter8-tools/iter8/storage/badgerdb"
	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	file, err := os.CreateTemp(".", "test")
	assert.NoError(t, err)
	defer func() {
		err := os.Remove(file.Name())
		assert.NoError(t, err)
	}()

	err = os.Setenv("METRICS_CONFIG_FILE", file.Name())
	assert.NoError(t, err)

	err = Start(ctx.Done())
	assert.Equal(t, err, http.ErrServerClosed)
}

func TestReadConfigDefaultPort(t *testing.T) {
	file, err := os.CreateTemp(".", "test")
	assert.NoError(t, err)
	defer func() {
		err := os.Remove(file.Name())
		assert.NoError(t, err)
	}()

	err = os.Setenv("METRICS_CONFIG_FILE", file.Name())
	assert.NoError(t, err)
	conf := &metricsConfig{}
	err = util.ReadConfig(configEnv, conf, func() {
		if nil == conf.Port {
			conf.Port = util.IntPointer(defaultPortNumber)
		}
	})
	assert.NoError(t, err)

	assert.Equal(t, defaultPortNumber, *conf.Port)
}

func TestReadConfigSetPort(t *testing.T) {
	expectedPortNumber := 8888

	file, err := os.CreateTemp(".", "test")
	assert.NoError(t, err)
	defer func() {
		err := os.Remove(file.Name())
		assert.NoError(t, err)
	}()

	_, err = file.Write([]byte(fmt.Sprintf("port: %d", expectedPortNumber)))
	assert.NoError(t, err)

	err = os.Setenv("METRICS_CONFIG_FILE", file.Name())
	assert.NoError(t, err)
	conf := &metricsConfig{}
	err = util.ReadConfig(configEnv, conf, func() {
		if nil == conf.Port {
			conf.Port = util.IntPointer(defaultPortNumber)
		}
	})
	assert.NoError(t, err)

	assert.Equal(t, expectedPortNumber, *conf.Port)
}

func TestGetMetricsInvalidMethod(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	getMetrics(w, req)
	res := w.Result()
	defer func() {
		err := res.Body.Close()
		assert.NoError(t, err)
	}()
	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
}

func TestGetMetricsMissingParameter(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	getMetrics(w, req)
	res := w.Result()
	defer func() {
		err := res.Body.Close()
		assert.NoError(t, err)
	}()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestGetMetricsNoRouteMap(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics?application=default%2Ftest", nil)
	getMetrics(w, req)
	res := w.Result()
	defer func() {
		err := res.Body.Close()
		assert.NoError(t, err)
	}()
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

type testRoutemaps struct {
	allroutemaps testroutemaps
}

func (cm *testRoutemaps) GetAllRoutemaps() controllers.RoutemapsInterface {
	return &cm.allroutemaps
}

func TestGetMetrics(t *testing.T) {
	testRM := testRoutemaps{
		allroutemaps: setupRoutemaps(t, *getTestRM("default", "test")),
	}
	allRoutemaps = &testRM

	tempDirPath := t.TempDir()

	client, err := badgerdb.GetClient(badger.DefaultOptions(tempDirPath), badgerdb.AdditionalOptions{})
	assert.NoError(t, err)

	app := "default/test"
	version := 0
	signature := "123456789"
	metric := "my-metric"
	user := "my-user"
	transaction := "my-transaction"
	value := 50.0

	err = client.SetMetric(app, version, signature, metric, user, transaction, value)
	assert.NoError(t, err)

	app = "default/test"
	version = 1
	signature = "987654321"
	metric = "my-metric"
	user = "my-user"
	transaction = "my-transaction-1"
	value = 75.0

	err = client.SetMetric(app, version, signature, metric, user, transaction, value)
	assert.NoError(t, err)

	abn.MetricsClient = client

	w := httptest.NewRecorder()
	rm := allRoutemaps.GetAllRoutemaps().GetRoutemapFromNamespaceName("default", "test")
	assert.NotNil(t, rm)
	req := httptest.NewRequest(http.MethodGet, "/metrics?application=default%2Ftest", nil)
	getMetrics(w, req)
	res := w.Result()
	defer func() {
		err := res.Body.Close()
		assert.NoError(t, err)
	}()

	var v map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&v)
	assert.NoError(t, err)
	//assert.Equal(t, "", string(data))

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestCalculateSummarizedMetric(t *testing.T) {
	summarizedMetric, err := calculateSummarizedMetric([]float64{1, 2, 3, 4, 5})
	assert.NoError(t, err)

	assert.Equal(t, 3.0, summarizedMetric.Mean)
	assert.Equal(t, 1.4142135623730951, summarizedMetric.StdDev)
	assert.Equal(t, 1.0, summarizedMetric.Min)
	assert.Equal(t, 5.0, summarizedMetric.Max)
	assert.Equal(t, uint64(5), summarizedMetric.Count)

	summarizedMetric, err = calculateSummarizedMetric([]float64{-1, -1, -1, -2, 5})
	assert.NoError(t, err)

	assert.Equal(t, 0.0, summarizedMetric.Mean)
	assert.Equal(t, 2.5298221281347035, summarizedMetric.StdDev)
	assert.Equal(t, -2.0, summarizedMetric.Min)
	assert.Equal(t, 5.0, summarizedMetric.Max)
	assert.Equal(t, uint64(5), summarizedMetric.Count)

	summarizedMetric, err = calculateSummarizedMetric([]float64{})
	assert.NoError(t, err)

	assert.Equal(t, 0.0, summarizedMetric.Mean)
	assert.Equal(t, 0.0, summarizedMetric.StdDev)
	assert.Equal(t, 0.0, summarizedMetric.Min)
	assert.Equal(t, 0.0, summarizedMetric.Max)
	assert.Equal(t, uint64(0), summarizedMetric.Count)
}

func TestCalculateHistogram(t *testing.T) {
	tests := []struct {
		data         map[string][]float64
		numBuckets   int
		decimalPlace float64
		result       string
	}{
		{
			data: map[string][]float64{
				"0": {1, 2, 3},
				"1": {3, 4, 5},
				"5": {10, 10, 10, 10, 10, 20, 30},
			},
			numBuckets:   10,
			decimalPlace: 5,
			result:       "[{\"Version\":\"0\",\"Bucket\":\"1 - 3.9\",\"Value\":3},{\"Version\":\"0\",\"Bucket\":\"3.9 - 6.8\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"6.8 - 9.69999\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"9.69999 - 12.6\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"12.6 - 15.5\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"15.5 - 18.39999\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"18.39999 - 21.3\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"21.3 - 24.2\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"24.2 - 27.1\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"27.1 - 30\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"1 - 3.9\",\"Value\":1},{\"Version\":\"1\",\"Bucket\":\"3.9 - 6.8\",\"Value\":2},{\"Version\":\"1\",\"Bucket\":\"6.8 - 9.69999\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"9.69999 - 12.6\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"12.6 - 15.5\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"15.5 - 18.39999\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"18.39999 - 21.3\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"21.3 - 24.2\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"24.2 - 27.1\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"27.1 - 30\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"1 - 3.9\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"3.9 - 6.8\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"6.8 - 9.69999\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"9.69999 - 12.6\",\"Value\":5},{\"Version\":\"5\",\"Bucket\":\"12.6 - 15.5\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"15.5 - 18.39999\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"18.39999 - 21.3\",\"Value\":1},{\"Version\":\"5\",\"Bucket\":\"21.3 - 24.2\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"24.2 - 27.1\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"27.1 - 30\",\"Value\":1}]",
		},
		{
			data: map[string][]float64{
				"0": {1, 2, 3},
				"1": {3, 4, 5},
				"5": {10, 10, 10, 10, 10, 20, 30},
			},
			numBuckets:   30,
			decimalPlace: 5,
			result:       "[{\"Version\":\"0\",\"Bucket\":\"1 - 1.96666\",\"Value\":1},{\"Version\":\"0\",\"Bucket\":\"1.96666 - 2.93333\",\"Value\":1},{\"Version\":\"0\",\"Bucket\":\"2.93333 - 3.9\",\"Value\":1},{\"Version\":\"0\",\"Bucket\":\"3.9 - 4.86666\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"4.86666 - 5.83333\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"5.83333 - 6.8\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"6.8 - 7.76666\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"7.76666 - 8.73333\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"8.73333 - 9.69999\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"9.69999 - 10.66666\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"10.66666 - 11.63333\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"11.63333 - 12.6\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"12.6 - 13.56666\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"13.56666 - 14.53333\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"14.53333 - 15.5\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"15.5 - 16.46666\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"16.46666 - 17.43333\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"17.43333 - 18.39999\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"18.39999 - 19.36666\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"19.36666 - 20.33333\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"20.33333 - 21.3\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"21.3 - 22.26666\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"22.26666 - 23.23333\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"23.23333 - 24.2\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"24.2 - 25.16666\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"25.16666 - 26.13333\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"26.13333 - 27.1\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"27.1 - 28.06666\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"28.06666 - 29.03333\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"29.03333 - 30\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"1 - 1.96666\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"1.96666 - 2.93333\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"2.93333 - 3.9\",\"Value\":1},{\"Version\":\"1\",\"Bucket\":\"3.9 - 4.86666\",\"Value\":1},{\"Version\":\"1\",\"Bucket\":\"4.86666 - 5.83333\",\"Value\":1},{\"Version\":\"1\",\"Bucket\":\"5.83333 - 6.8\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"6.8 - 7.76666\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"7.76666 - 8.73333\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"8.73333 - 9.69999\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"9.69999 - 10.66666\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"10.66666 - 11.63333\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"11.63333 - 12.6\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"12.6 - 13.56666\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"13.56666 - 14.53333\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"14.53333 - 15.5\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"15.5 - 16.46666\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"16.46666 - 17.43333\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"17.43333 - 18.39999\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"18.39999 - 19.36666\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"19.36666 - 20.33333\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"20.33333 - 21.3\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"21.3 - 22.26666\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"22.26666 - 23.23333\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"23.23333 - 24.2\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"24.2 - 25.16666\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"25.16666 - 26.13333\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"26.13333 - 27.1\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"27.1 - 28.06666\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"28.06666 - 29.03333\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"29.03333 - 30\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"1 - 1.96666\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"1.96666 - 2.93333\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"2.93333 - 3.9\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"3.9 - 4.86666\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"4.86666 - 5.83333\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"5.83333 - 6.8\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"6.8 - 7.76666\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"7.76666 - 8.73333\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"8.73333 - 9.69999\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"9.69999 - 10.66666\",\"Value\":5},{\"Version\":\"5\",\"Bucket\":\"10.66666 - 11.63333\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"11.63333 - 12.6\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"12.6 - 13.56666\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"13.56666 - 14.53333\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"14.53333 - 15.5\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"15.5 - 16.46666\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"16.46666 - 17.43333\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"17.43333 - 18.39999\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"18.39999 - 19.36666\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"19.36666 - 20.33333\",\"Value\":1},{\"Version\":\"5\",\"Bucket\":\"20.33333 - 21.3\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"21.3 - 22.26666\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"22.26666 - 23.23333\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"23.23333 - 24.2\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"24.2 - 25.16666\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"25.16666 - 26.13333\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"26.13333 - 27.1\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"27.1 - 28.06666\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"28.06666 - 29.03333\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"29.03333 - 30\",\"Value\":1}]",
		},
		{
			data: map[string][]float64{
				"0": {1, 2, 3},
				"1": {3, 4, 5},
				"5": {10, 10, 10, 10, 10, 20, 30},
			}, numBuckets: 5,
			decimalPlace: 1,
			result:       "[{\"Version\":\"0\",\"Bucket\":\"1 - 6.8\",\"Value\":3},{\"Version\":\"0\",\"Bucket\":\"6.8 - 12.6\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"12.6 - 18.4\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"18.4 - 24.2\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"24.2 - 30\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"1 - 6.8\",\"Value\":3},{\"Version\":\"1\",\"Bucket\":\"6.8 - 12.6\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"12.6 - 18.4\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"18.4 - 24.2\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"24.2 - 30\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"1 - 6.8\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"6.8 - 12.6\",\"Value\":5},{\"Version\":\"5\",\"Bucket\":\"12.6 - 18.4\",\"Value\":0},{\"Version\":\"5\",\"Bucket\":\"18.4 - 24.2\",\"Value\":1},{\"Version\":\"5\",\"Bucket\":\"24.2 - 30\",\"Value\":1}]",
		},
	}

	for _, test := range tests {
		summarizedMetric, err := calculateHistogram(test.data, test.numBuckets, test.decimalPlace)
		assert.NoError(t, err)

		// Sort summarizedMetric
		// Even though the buckets in each version is sorted, the order of the versions may not
		sort.Slice(summarizedMetric, func(i, j int) bool {
			iVersion := summarizedMetric[i].Version
			ifVersion, err := strconv.ParseFloat(iVersion, 64)
			if err != nil {
				assert.Fail(t, "cannot parse string \"%s\" into float64", iVersion)
			}

			jVersion := summarizedMetric[j].Version
			jfVersion, err := strconv.ParseFloat(jVersion, 64)
			if err != nil {
				assert.Fail(t, "cannot parse string \"%s\" into float64", jVersion)
			}

			if ifVersion == jfVersion {
				// Compare the buckets
				iBucket := summarizedMetric[i].Bucket
				jBucket := summarizedMetric[j].Bucket

				re := regexp.MustCompile("[0-9.]+")
				iBucketMin := re.FindAllString(iBucket, 1)
				jBucketMin := re.FindAllString(jBucket, 1)

				if iBucketMin == nil {
					assert.Fail(t, "cannot parse find number in string \"%s\"", iBucket)
				} else if jBucketMin == nil {
					assert.Fail(t, "cannot parse find number in string \"%s\"", jBucket)
				}

				ifBucketMin, err := strconv.ParseFloat(iBucketMin[0], 64)
				if err != nil {
					assert.Fail(t, "cannot parse string \"%s\" into float64", iBucketMin)
				}

				jfBucketMin, err := strconv.ParseFloat(jBucketMin[0], 64)
				if err != nil {
					assert.Fail(t, "cannot parse string \"%s\" into float64", jBucketMin)
				}

				return ifBucketMin < jfBucketMin
			}

			return ifVersion < jfVersion
		})

		jsonSummarizeMetric, err := json.Marshal(summarizedMetric)
		assert.NoError(t, err)
		assert.Equal(t, test.result, string(jsonSummarizeMetric))

	}
}

func setupRoutemaps(t *testing.T, initialroutemaps ...testroutemap) testroutemaps {
	routemaps := testroutemaps{
		nsRoutemap: make(map[string]testroutemapsByName),
	}

	for i := range initialroutemaps {

		if _, ok := routemaps.nsRoutemap[initialroutemaps[i].namespace]; !ok {
			routemaps.nsRoutemap[initialroutemaps[i].namespace] = make(testroutemapsByName)
		}
		(routemaps.nsRoutemap[initialroutemaps[i].namespace])[initialroutemaps[i].name] = &initialroutemaps[i]
	}

	return routemaps
}

func getTestRM(namespace, name string) *testroutemap {
	return &testroutemap{
		namespace: namespace,
		name:      name,
		versions: []testversion{
			{signature: util.StringPointer("123456789")},
			{signature: util.StringPointer("987654321")},
		},
		normalizedWeights: []uint32{1, 1},
	}

}

func TestGetHTTPDashboardHelper(t *testing.T) {
	fortioResult := util.FortioResult{}
	err := json.Unmarshal([]byte(fortioResultJSON), &fortioResult)
	assert.NoError(t, err)

	dashboard := getHTTPDashboardHelper(fortioResult)
	assert.NotNil(t, dashboard)
	dashboardBytes, err := json.Marshal(dashboard)
	assert.NoError(t, err)

	assert.Equal(
		t,
		fortioDashboardJSON,
		string(dashboardBytes),
	)
}

func TestGetGHZDashboardHelper(t *testing.T) {
	ghzResult := util.GHZResult{}
	err := json.Unmarshal([]byte(ghzResultJSON), &ghzResult)
	assert.NoError(t, err)

	dashboard := getGRPCDashboardHelper(ghzResult)

	assert.NotNil(t, dashboard)
	dashboardBytes, err := json.Marshal(dashboard)
	assert.NoError(t, err)
	assert.Equal(
		t,
		ghzDashboardJSON,
		string(dashboardBytes),
	)
}

func TestPutResultInvalidMethod(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, util.PerformanceResultPath, nil)
	putResult(w, req)
	res := w.Result()
	defer func() {
		err := res.Body.Close()
		assert.NoError(t, err)
	}()
	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
}

func TestPutResultMissingParameter(t *testing.T) {
	tests := []struct {
		queryParams        url.Values
		expectedStatusCode int
	}{
		{
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			queryParams: url.Values{
				"namespace": {"default"},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			queryParams: url.Values{
				"experiment": {"default"},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		u, err := url.ParseRequestURI(util.PerformanceResultPath)
		assert.NoError(t, err)
		u.RawQuery = test.queryParams.Encode()
		urlStr := fmt.Sprintf("%v", u)

		req := httptest.NewRequest(http.MethodPut, urlStr, nil)

		putResult(w, req)
		res := w.Result()
		defer func() {
			err := res.Body.Close()
			assert.NoError(t, err)
		}()

		assert.Equal(t, test.expectedStatusCode, res.StatusCode)
	}
}

func TestPutResult(t *testing.T) {
	// instantiate metrics client
	tempDirPath := t.TempDir()
	client, err := badgerdb.GetClient(badger.DefaultOptions(tempDirPath), badgerdb.AdditionalOptions{})
	assert.NoError(t, err)
	abn.MetricsClient = client

	w := httptest.NewRecorder()

	// construct inputs to putResult
	u, err := url.ParseRequestURI(util.PerformanceResultPath)
	assert.NoError(t, err)
	params := url.Values{
		"namespace":  {"default"},
		"experiment": {"default"},
	}
	u.RawQuery = params.Encode()
	urlStr := fmt.Sprintf("%v", u)

	payload := `{"hello":"world"}`
	req := httptest.NewRequest(http.MethodPut, urlStr, bytes.NewBuffer([]byte(payload)))

	// put result into the metrics client
	putResult(w, req)
	res := w.Result()
	defer func() {
		err := res.Body.Close()
		assert.NoError(t, err)
	}()

	// check to see if the result is stored in the metrics client
	result, err := abn.MetricsClient.GetResult("default", "default")
	assert.NoError(t, err)
	assert.Equal(t, payload, string(result))
}

func TestGetHTTPDashboardInvalidMethod(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/httpDashboard", nil)
	getHTTPDashboard(w, req)
	res := w.Result()
	defer func() {
		err := res.Body.Close()
		assert.NoError(t, err)
	}()
	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
}

func TestGetHTTPDashboardMissingParameter(t *testing.T) {
	tests := []struct {
		queryParams        url.Values
		expectedStatusCode int
	}{
		{
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			queryParams: url.Values{
				"namespace": {"default"},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			queryParams: url.Values{
				"experiment": {"default"},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		u, err := url.ParseRequestURI(util.PerformanceResultPath)
		assert.NoError(t, err)
		u.RawQuery = test.queryParams.Encode()
		urlStr := fmt.Sprintf("%v", u)
		req := httptest.NewRequest(http.MethodGet, urlStr, nil)

		getHTTPDashboard(w, req)
		res := w.Result()
		defer func() {
			err := res.Body.Close()
			assert.NoError(t, err)
		}()

		assert.Equal(t, test.expectedStatusCode, res.StatusCode)
	}
}

const fortioResultJSON = `{
	"EndpointResults": {
		"http://httpbin.default/get": {
			"RunType": "HTTP",
			"Labels": "",
			"StartTime": "2023-07-21T14:00:40.134434969Z",
			"RequestedQPS": "8",
			"RequestedDuration": "exactly 100 calls",
			"ActualQPS": 7.975606391552989,
			"ActualDuration": 12538231589,
			"NumThreads": 4,
			"Version": "1.57.3",
			"DurationHistogram": {
				"Count": 100,
				"Min": 0.004223875,
				"Max": 0.040490042,
				"Sum": 1.5977100850000001,
				"Avg": 0.015977100850000002,
				"StdDev": 0.008340658047253256,
				"Data": [
					{
						"Start": 0.004223875,
						"End": 0.005,
						"Percent": 5,
						"Count": 5
					},
					{
						"Start": 0.005,
						"End": 0.006,
						"Percent": 10,
						"Count": 5
					},
					{
						"Start": 0.006,
						"End": 0.007,
						"Percent": 14,
						"Count": 4
					},
					{
						"Start": 0.007,
						"End": 0.008,
						"Percent": 19,
						"Count": 5
					},
					{
						"Start": 0.008,
						"End": 0.009000000000000001,
						"Percent": 24,
						"Count": 5
					},
					{
						"Start": 0.009000000000000001,
						"End": 0.01,
						"Percent": 28,
						"Count": 4
					},
					{
						"Start": 0.01,
						"End": 0.011,
						"Percent": 33,
						"Count": 5
					},
					{
						"Start": 0.011,
						"End": 0.012,
						"Percent": 36,
						"Count": 3
					},
					{
						"Start": 0.012,
						"End": 0.014,
						"Percent": 48,
						"Count": 12
					},
					{
						"Start": 0.014,
						"End": 0.016,
						"Percent": 55,
						"Count": 7
					},
					{
						"Start": 0.016,
						"End": 0.018000000000000002,
						"Percent": 65,
						"Count": 10
					},
					{
						"Start": 0.018000000000000002,
						"End": 0.02,
						"Percent": 74,
						"Count": 9
					},
					{
						"Start": 0.02,
						"End": 0.025,
						"Percent": 85,
						"Count": 11
					},
					{
						"Start": 0.025,
						"End": 0.03,
						"Percent": 93,
						"Count": 8
					},
					{
						"Start": 0.03,
						"End": 0.035,
						"Percent": 98,
						"Count": 5
					},
					{
						"Start": 0.035,
						"End": 0.04,
						"Percent": 99,
						"Count": 1
					},
					{
						"Start": 0.04,
						"End": 0.040490042,
						"Percent": 100,
						"Count": 1
					}
				],
				"Percentiles": [
					{
						"Percentile": 50,
						"Value": 0.014571428571428572
					},
					{
						"Percentile": 75,
						"Value": 0.020454545454545454
					},
					{
						"Percentile": 90,
						"Value": 0.028125
					},
					{
						"Percentile": 95,
						"Value": 0.032
					},
					{
						"Percentile": 99,
						"Value": 0.04
					},
					{
						"Percentile": 99.9,
						"Value": 0.0404410378
					}
				]
			},
			"ErrorsDurationHistogram": {
				"Count": 0,
				"Min": 0,
				"Max": 0,
				"Sum": 0,
				"Avg": 0,
				"StdDev": 0,
				"Data": null
			},
			"Exactly": 100,
			"Jitter": false,
			"Uniform": false,
			"NoCatchUp": false,
			"RunID": 0,
			"AccessLoggerInfo": "",
			"ID": "2023-07-21-140040",
			"RetCodes": {
				"200": 100
			},
			"IPCountMap": {
				"10.96.108.76:80": 4
			},
			"Insecure": false,
			"MTLS": false,
			"CACert": "",
			"Cert": "",
			"Key": "",
			"UnixDomainSocket": "",
			"URL": "http://httpbin.default/get",
			"NumConnections": 1,
			"Compression": false,
			"DisableFastClient": false,
			"HTTP10": false,
			"H2": false,
			"DisableKeepAlive": false,
			"AllowHalfClose": false,
			"FollowRedirects": false,
			"Resolve": "",
			"HTTPReqTimeOut": 3000000000,
			"UserCredentials": "",
			"ContentType": "",
			"Payload": null,
			"MethodOverride": "",
			"LogErrors": false,
			"SequentialWarmup": false,
			"ConnReuseRange": [
				0,
				0
			],
			"NoResolveEachConn": false,
			"Offset": 0,
			"Resolution": 0.001,
			"Sizes": {
				"Count": 100,
				"Min": 413,
				"Max": 413,
				"Sum": 41300,
				"Avg": 413,
				"StdDev": 0,
				"Data": [
					{
						"Start": 413,
						"End": 413,
						"Percent": 100,
						"Count": 100
					}
				]
			},
			"HeaderSizes": {
				"Count": 100,
				"Min": 230,
				"Max": 230,
				"Sum": 23000,
				"Avg": 230,
				"StdDev": 0,
				"Data": [
					{
						"Start": 230,
						"End": 230,
						"Percent": 100,
						"Count": 100
					}
				]
			},
			"Sockets": [
				1,
				1,
				1,
				1
			],
			"SocketCount": 4,
			"ConnectionStats": {
				"Count": 4,
				"Min": 0.001385875,
				"Max": 0.001724375,
				"Sum": 0.006404583,
				"Avg": 0.00160114575,
				"StdDev": 0.00013101857565508474,
				"Data": [
					{
						"Start": 0.001385875,
						"End": 0.001724375,
						"Percent": 100,
						"Count": 4
					}
				],
				"Percentiles": [
					{
						"Percentile": 50,
						"Value": 0.0014987083333333332
					},
					{
						"Percentile": 75,
						"Value": 0.0016115416666666667
					},
					{
						"Percentile": 90,
						"Value": 0.0016792416666666667
					},
					{
						"Percentile": 95,
						"Value": 0.0017018083333333333
					},
					{
						"Percentile": 99,
						"Value": 0.0017198616666666668
					},
					{
						"Percentile": 99.9,
						"Value": 0.0017239236666666668
					}
				]
			},
			"AbortOn": 0
		}
	},
	"Summary": {
		"numVersions": 1,
		"versionNames": null,
		"metricsInfo": {
			"http/latency": {
				"description": "Latency Histogram",
				"units": "msec",
				"type": "Histogram"
			},
			"http://httpbin.default/get/error-count": {
				"description": "number of responses that were errors",
				"type": "Counter"
			},
			"http://httpbin.default/get/error-rate": {
				"description": "fraction of responses that were errors",
				"type": "Gauge"
			},
			"http://httpbin.default/get/latency-max": {
				"description": "maximum of observed latency values",
				"units": "msec",
				"type": "Gauge"
			},
			"http://httpbin.default/get/latency-mean": {
				"description": "mean of observed latency values",
				"units": "msec",
				"type": "Gauge"
			},
			"http://httpbin.default/get/latency-min": {
				"description": "minimum of observed latency values",
				"units": "msec",
				"type": "Gauge"
			},
			"http://httpbin.default/get/latency-p50": {
				"description": "50-th percentile of observed latency values",
				"units": "msec",
				"type": "Gauge"
			},
			"http://httpbin.default/get/latency-p75": {
				"description": "75-th percentile of observed latency values",
				"units": "msec",
				"type": "Gauge"
			},
			"http://httpbin.default/get/latency-p90": {
				"description": "90-th percentile of observed latency values",
				"units": "msec",
				"type": "Gauge"
			},
			"http://httpbin.default/get/latency-p95": {
				"description": "95-th percentile of observed latency values",
				"units": "msec",
				"type": "Gauge"
			},
			"http://httpbin.default/get/latency-p99": {
				"description": "99-th percentile of observed latency values",
				"units": "msec",
				"type": "Gauge"
			},
			"http://httpbin.default/get/latency-p99.9": {
				"description": "99.9-th percentile of observed latency values",
				"units": "msec",
				"type": "Gauge"
			},
			"http://httpbin.default/get/latency-stddev": {
				"description": "standard deviation of observed latency values",
				"units": "msec",
				"type": "Gauge"
			},
			"http://httpbin.default/get/request-count": {
				"description": "number of requests sent",
				"type": "Counter"
			}
		},
		"nonHistMetricValues": [
			{
				"http://httpbin.default/get/error-count": [
					0
				],
				"http://httpbin.default/get/error-rate": [
					0
				],
				"http://httpbin.default/get/latency-max": [
					40.490041999999995
				],
				"http://httpbin.default/get/latency-mean": [
					15.977100850000001
				],
				"http://httpbin.default/get/latency-min": [
					4.2238750000000005
				],
				"http://httpbin.default/get/latency-p50": [
					14.571428571428571
				],
				"http://httpbin.default/get/latency-p75": [
					20.454545454545453
				],
				"http://httpbin.default/get/latency-p90": [
					28.125
				],
				"http://httpbin.default/get/latency-p95": [
					32
				],
				"http://httpbin.default/get/latency-p99": [
					40
				],
				"http://httpbin.default/get/latency-p99.9": [
					40.441037800000004
				],
				"http://httpbin.default/get/latency-stddev": [
					8.340658047253257
				],
				"http://httpbin.default/get/request-count": [
					100
				]
			}
		],
		"histMetricValues": [
			{
				"http/latency": [
					{
						"lower": 4.2238750000000005,
						"upper": 5,
						"count": 5
					},
					{
						"lower": 5,
						"upper": 6,
						"count": 5
					},
					{
						"lower": 6,
						"upper": 7,
						"count": 4
					},
					{
						"lower": 7,
						"upper": 8,
						"count": 5
					},
					{
						"lower": 8,
						"upper": 9.000000000000002,
						"count": 5
					},
					{
						"lower": 9.000000000000002,
						"upper": 10,
						"count": 4
					},
					{
						"lower": 10,
						"upper": 11,
						"count": 5
					},
					{
						"lower": 11,
						"upper": 12,
						"count": 3
					},
					{
						"lower": 12,
						"upper": 14,
						"count": 12
					},
					{
						"lower": 14,
						"upper": 16,
						"count": 7
					},
					{
						"lower": 16,
						"upper": 18.000000000000004,
						"count": 10
					},
					{
						"lower": 18.000000000000004,
						"upper": 20,
						"count": 9
					},
					{
						"lower": 20,
						"upper": 25,
						"count": 11
					},
					{
						"lower": 25,
						"upper": 30,
						"count": 8
					},
					{
						"lower": 30,
						"upper": 35,
						"count": 5
					},
					{
						"lower": 35,
						"upper": 40,
						"count": 1
					},
					{
						"lower": 40,
						"upper": 40.490041999999995,
						"count": 1
					}
				]
			}
		],
		"SummaryMetricValues": [
			{}
		]
	}
}`

const fortioDashboardJSON = `{"Endpoints":{"http://httpbin.default/get":{"Durations":[{"Version":"0","Bucket":"4.2 - 5","Value":5},{"Version":"0","Bucket":"5 - 6","Value":5},{"Version":"0","Bucket":"6 - 7","Value":4},{"Version":"0","Bucket":"7 - 8","Value":5},{"Version":"0","Bucket":"8 - 9","Value":5},{"Version":"0","Bucket":"9 - 10","Value":4},{"Version":"0","Bucket":"10 - 11","Value":5},{"Version":"0","Bucket":"11 - 12","Value":3},{"Version":"0","Bucket":"12 - 14","Value":12},{"Version":"0","Bucket":"14 - 16","Value":7},{"Version":"0","Bucket":"16 - 18","Value":10},{"Version":"0","Bucket":"18 - 20","Value":9},{"Version":"0","Bucket":"20 - 25","Value":11},{"Version":"0","Bucket":"25 - 30","Value":8},{"Version":"0","Bucket":"30 - 35","Value":5},{"Version":"0","Bucket":"35 - 40","Value":1},{"Version":"0","Bucket":"40 - 40.4","Value":1}],"Statistics":{"Count":100,"Mean":15.977100850000001,"StdDev":8.340658047253257,"Min":4.2238750000000005,"Max":40.490041999999995},"Error durations":[],"Error statistics":{"Count":0,"Mean":0,"StdDev":0,"Min":0,"Max":0},"Return codes":{"200":100}}},"Summary":{"numVersions":1,"versionNames":null,"metricsInfo":{"http/latency":{"description":"Latency Histogram","units":"msec","type":"Histogram"},"http://httpbin.default/get/error-count":{"description":"number of responses that were errors","type":"Counter"},"http://httpbin.default/get/error-rate":{"description":"fraction of responses that were errors","type":"Gauge"},"http://httpbin.default/get/latency-max":{"description":"maximum of observed latency values","units":"msec","type":"Gauge"},"http://httpbin.default/get/latency-mean":{"description":"mean of observed latency values","units":"msec","type":"Gauge"},"http://httpbin.default/get/latency-min":{"description":"minimum of observed latency values","units":"msec","type":"Gauge"},"http://httpbin.default/get/latency-p50":{"description":"50-th percentile of observed latency values","units":"msec","type":"Gauge"},"http://httpbin.default/get/latency-p75":{"description":"75-th percentile of observed latency values","units":"msec","type":"Gauge"},"http://httpbin.default/get/latency-p90":{"description":"90-th percentile of observed latency values","units":"msec","type":"Gauge"},"http://httpbin.default/get/latency-p95":{"description":"95-th percentile of observed latency values","units":"msec","type":"Gauge"},"http://httpbin.default/get/latency-p99":{"description":"99-th percentile of observed latency values","units":"msec","type":"Gauge"},"http://httpbin.default/get/latency-p99.9":{"description":"99.9-th percentile of observed latency values","units":"msec","type":"Gauge"},"http://httpbin.default/get/latency-stddev":{"description":"standard deviation of observed latency values","units":"msec","type":"Gauge"},"http://httpbin.default/get/request-count":{"description":"number of requests sent","type":"Counter"}},"nonHistMetricValues":[{"http://httpbin.default/get/error-count":[0],"http://httpbin.default/get/error-rate":[0],"http://httpbin.default/get/latency-max":[40.490041999999995],"http://httpbin.default/get/latency-mean":[15.977100850000001],"http://httpbin.default/get/latency-min":[4.2238750000000005],"http://httpbin.default/get/latency-p50":[14.571428571428571],"http://httpbin.default/get/latency-p75":[20.454545454545453],"http://httpbin.default/get/latency-p90":[28.125],"http://httpbin.default/get/latency-p95":[32],"http://httpbin.default/get/latency-p99":[40],"http://httpbin.default/get/latency-p99.9":[40.441037800000004],"http://httpbin.default/get/latency-stddev":[8.340658047253257],"http://httpbin.default/get/request-count":[100]}],"histMetricValues":[{"http/latency":[{"lower":4.2238750000000005,"upper":5,"count":5},{"lower":5,"upper":6,"count":5},{"lower":6,"upper":7,"count":4},{"lower":7,"upper":8,"count":5},{"lower":8,"upper":9.000000000000002,"count":5},{"lower":9.000000000000002,"upper":10,"count":4},{"lower":10,"upper":11,"count":5},{"lower":11,"upper":12,"count":3},{"lower":12,"upper":14,"count":12},{"lower":14,"upper":16,"count":7},{"lower":16,"upper":18.000000000000004,"count":10},{"lower":18.000000000000004,"upper":20,"count":9},{"lower":20,"upper":25,"count":11},{"lower":25,"upper":30,"count":8},{"lower":30,"upper":35,"count":5},{"lower":35,"upper":40,"count":1},{"lower":40,"upper":40.490041999999995,"count":1}]}],"SummaryMetricValues":[{}]}}`

func TestGetHTTPDashboard(t *testing.T) {
	// instantiate metrics client
	tempDirPath := t.TempDir()
	client, err := badgerdb.GetClient(badger.DefaultOptions(tempDirPath), badgerdb.AdditionalOptions{})
	assert.NoError(t, err)
	abn.MetricsClient = client

	// preload metric client with result
	err = abn.MetricsClient.SetResult("default", "default", []byte(fortioResultJSON))
	assert.NoError(t, err)

	w := httptest.NewRecorder()

	// construct inputs to getHTTPDashboard
	u, err := url.ParseRequestURI(util.PerformanceResultPath)
	assert.NoError(t, err)
	params := url.Values{
		"namespace":  {"default"},
		"experiment": {"default"},
	}
	u.RawQuery = params.Encode()
	urlStr := fmt.Sprintf("%v", u)

	req := httptest.NewRequest(http.MethodGet, urlStr, nil)

	// get HTTP dashboard based on result in metrics client
	getHTTPDashboard(w, req)
	res := w.Result()
	defer func() {
		err := res.Body.Close()
		assert.NoError(t, err)
	}()

	// check the HTTP dashboard
	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(
		t,
		fortioDashboardJSON,
		string(body),
	)
}

func TestGetGHZDashboardInvalidMethod(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, util.PerformanceResultPath, nil)
	putResult(w, req)
	res := w.Result()
	defer func() {
		err := res.Body.Close()
		assert.NoError(t, err)
	}()
	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
}

func TestGetGHZDashboardMissingParameter(t *testing.T) {
	tests := []struct {
		queryParams        url.Values
		expectedStatusCode int
	}{
		{
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			queryParams: url.Values{
				"namespace": {"default"},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			queryParams: url.Values{
				"experiment": {"default"},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		u, err := url.ParseRequestURI(util.PerformanceResultPath)
		assert.NoError(t, err)
		u.RawQuery = test.queryParams.Encode()
		urlStr := fmt.Sprintf("%v", u)

		req := httptest.NewRequest(http.MethodPut, urlStr, nil)

		putResult(w, req)
		res := w.Result()
		defer func() {
			err := res.Body.Close()
			assert.NoError(t, err)
		}()

		assert.Equal(t, test.expectedStatusCode, res.StatusCode)
	}
}

const ghzResultJSON = `{
	"EndpointResults": {
		"routeguide.RouteGuide.GetFeature": {
			"date": "2023-07-17T12:23:56Z",
			"endReason": "normal",
			"options": {
				"call": "routeguide.RouteGuide.GetFeature",
				"host": "routeguide.default:50051",
				"proto": "/tmp/ghz.proto",
				"import-paths": [
					"/tmp",
					"."
				],
				"insecure": true,
				"load-schedule": "const",
				"load-start": 0,
				"load-end": 0,
				"load-step": 0,
				"load-step-duration": 0,
				"load-max-duration": 0,
				"concurrency": 50,
				"concurrency-schedule": "const",
				"concurrency-start": 1,
				"concurrency-end": 0,
				"concurrency-step": 0,
				"concurrency-step-duration": 0,
				"concurrency-max-duration": 0,
				"total": 200,
				"connections": 1,
				"dial-timeout": 10000000000,
				"data": {
					"latitude": 407838351,
					"longitude": -746143763
				},
				"binary": false,
				"CPUs": 5,
				"count-errors": true
			},
			"count": 200,
			"total": 592907667,
			"average": 25208185,
			"fastest": 32375,
			"slowest": 195740917,
			"rps": 337.3206506368217,
			"errorDistribution": {
				"rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 10.96.20.53:50051: connect: connection refused\"": 200
			},
			"statusCodeDistribution": {
				"Unavailable": 200
			},
			"latencyDistribution": [
				{
					"percentage": 10,
					"latency": 35584
				},
				{
					"percentage": 25,
					"latency": 39958
				},
				{
					"percentage": 50,
					"latency": 86208
				},
				{
					"percentage": 75,
					"latency": 12777625
				},
				{
					"percentage": 90,
					"latency": 106714334
				},
				{
					"percentage": 95,
					"latency": 189847000
				},
				{
					"percentage": 99,
					"latency": 195400792
				}
			],
			"histogram": [
				{
					"mark": 0.000032375,
					"count": 1,
					"frequency": 0.005
				},
				{
					"mark": 0.0196032292,
					"count": 167,
					"frequency": 0.835
				},
				{
					"mark": 0.0391740834,
					"count": 0,
					"frequency": 0
				},
				{
					"mark": 0.05874493759999999,
					"count": 0,
					"frequency": 0
				},
				{
					"mark": 0.07831579179999999,
					"count": 0,
					"frequency": 0
				},
				{
					"mark": 0.097886646,
					"count": 3,
					"frequency": 0.015
				},
				{
					"mark": 0.11745750019999998,
					"count": 13,
					"frequency": 0.065
				},
				{
					"mark": 0.1370283544,
					"count": 0,
					"frequency": 0
				},
				{
					"mark": 0.15659920859999998,
					"count": 0,
					"frequency": 0
				},
				{
					"mark": 0.17617006279999997,
					"count": 0,
					"frequency": 0
				},
				{
					"mark": 0.195740917,
					"count": 16,
					"frequency": 0.08
				}
			],
			"details": [
				{
					"timestamp": "2023-07-17T12:23:56.089998719Z",
					"latency": 14490041,
					"error": "rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 10.96.20.53:50051: connect: connection refused\"",
					"status": "Unavailable"
				},
				{
					"timestamp": "2023-07-17T12:23:56.090471886Z",
					"latency": 13759125,
					"error": "rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 10.96.20.53:50051: connect: connection refused\"",
					"status": "Unavailable"
				},
				{
					"timestamp": "2023-07-17T12:23:56.090528678Z",
					"latency": 194468542,
					"error": "rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 10.96.20.53:50051: connect: connection refused\"",
					"status": "Unavailable"
				},
				{
					"timestamp": "2023-07-17T12:23:56.090079886Z",
					"latency": 105031291,
					"error": "rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 10.96.20.53:50051: connect: connection refused\"",
					"status": "Unavailable"
				},
				{
					"timestamp": "2023-07-17T12:23:56.090224928Z",
					"latency": 100337083,
					"error": "rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 10.96.20.53:50051: connect: connection refused\"",
					"status": "Unavailable"
				},
				{
					"timestamp": "2023-07-17T12:23:56.091097053Z",
					"latency": 12463750,
					"error": "rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 10.96.20.53:50051: connect: connection refused\"",
					"status": "Unavailable"
				},
				{
					"timestamp": "2023-07-17T12:23:56.091135844Z",
					"latency": 12603875,
					"error": "rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 10.96.20.53:50051: connect: connection refused\"",
					"status": "Unavailable"
				},
				{
					"timestamp": "2023-07-17T12:23:56.478469636Z",
					"latency": 86208,
					"error": "rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 10.96.20.53:50051: connect: connection refused\"",
					"status": "Unavailable"
				}
			]
		}
	}
}`

const ghzDashboardJSON = `{"Endpoints":{"routeguide.RouteGuide.GetFeature":{"Durations":[{"Version":"0","Bucket":"0.032","Value":1},{"Version":"0","Bucket":"19.603","Value":167},{"Version":"0","Bucket":"39.174","Value":0},{"Version":"0","Bucket":"58.744","Value":0},{"Version":"0","Bucket":"78.315","Value":0},{"Version":"0","Bucket":"97.886","Value":3},{"Version":"0","Bucket":"117.457","Value":13},{"Version":"0","Bucket":"137.028","Value":0},{"Version":"0","Bucket":"156.599","Value":0},{"Version":"0","Bucket":"176.17","Value":0},{"Version":"0","Bucket":"195.74","Value":16}],"Statistics":{"Count":200,"ErrorCount":200},"Status codes":{"Unavailable":200}}},"Summary":{"numVersions":0,"versionNames":null,"SummaryMetricValues":null}}`

func TestGetGHZDashboard(t *testing.T) {
	// instantiate metrics client
	tempDirPath := t.TempDir()
	client, err := badgerdb.GetClient(badger.DefaultOptions(tempDirPath), badgerdb.AdditionalOptions{})
	assert.NoError(t, err)
	abn.MetricsClient = client

	// preload metric client with result
	err = abn.MetricsClient.SetResult("default", "default", []byte(ghzResultJSON))
	assert.NoError(t, err)

	w := httptest.NewRecorder()

	// construct inputs to getGHZDashboard
	u, err := url.ParseRequestURI(util.PerformanceResultPath)
	assert.NoError(t, err)
	params := url.Values{
		"namespace":  {"default"},
		"experiment": {"default"},
	}
	u.RawQuery = params.Encode()
	urlStr := fmt.Sprintf("%v", u)

	req := httptest.NewRequest(http.MethodGet, urlStr, nil)

	// get ghz dashboard based on result in metrics client
	getGRPCDashboard(w, req)
	res := w.Result()
	defer func() {
		err := res.Body.Close()
		assert.NoError(t, err)
	}()

	// check the ghz dashboard
	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(
		t,
		ghzDashboardJSON,
		string(body),
	)
}
