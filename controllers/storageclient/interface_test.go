package storageclient

import (
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateSummarizedMetric(t *testing.T) {
	summarizedMetric, err := CalculateSummarizedMetric([]float64{1, 2, 3, 4, 5})
	assert.NoError(t, err)

	assert.Equal(t, 3.0, summarizedMetric.Mean)
	assert.Equal(t, 1.4142135623730951, summarizedMetric.StdDev)
	assert.Equal(t, 1.0, summarizedMetric.Min)
	assert.Equal(t, 5.0, summarizedMetric.Max)
	assert.Equal(t, uint64(5), summarizedMetric.Count)

	summarizedMetric, err = CalculateSummarizedMetric([]float64{-1, -1, -1, -2, 5})
	assert.NoError(t, err)

	assert.Equal(t, 0.0, summarizedMetric.Mean)
	assert.Equal(t, 2.5298221281347035, summarizedMetric.StdDev)
	assert.Equal(t, -2.0, summarizedMetric.Min)
	assert.Equal(t, 5.0, summarizedMetric.Max)
	assert.Equal(t, uint64(5), summarizedMetric.Count)

	summarizedMetric, err = CalculateSummarizedMetric([]float64{})
	assert.NoError(t, err)

	assert.Equal(t, 0.0, summarizedMetric.Mean)
	assert.Equal(t, 0.0, summarizedMetric.StdDev)
	assert.Equal(t, 0.0, summarizedMetric.Min)
	assert.Equal(t, 0.0, summarizedMetric.Max)
	assert.Equal(t, uint64(0), summarizedMetric.Count)
}

func TestGetGrafanaHistogram(t *testing.T) {
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
			result:       "[{\"Version\":\"0\",\"Bucket\":\"1 - 3.9\",\"Count\":3},{\"Version\":\"0\",\"Bucket\":\"3.9 - 6.8\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"6.8 - 9.69999\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"9.69999 - 12.6\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"12.6 - 15.5\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"15.5 - 18.39999\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"18.39999 - 21.3\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"21.3 - 24.2\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"24.2 - 27.1\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"27.1 - 30\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"1 - 3.9\",\"Count\":1},{\"Version\":\"1\",\"Bucket\":\"3.9 - 6.8\",\"Count\":2},{\"Version\":\"1\",\"Bucket\":\"6.8 - 9.69999\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"9.69999 - 12.6\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"12.6 - 15.5\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"15.5 - 18.39999\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"18.39999 - 21.3\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"21.3 - 24.2\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"24.2 - 27.1\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"27.1 - 30\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"1 - 3.9\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"3.9 - 6.8\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"6.8 - 9.69999\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"9.69999 - 12.6\",\"Count\":5},{\"Version\":\"5\",\"Bucket\":\"12.6 - 15.5\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"15.5 - 18.39999\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"18.39999 - 21.3\",\"Count\":1},{\"Version\":\"5\",\"Bucket\":\"21.3 - 24.2\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"24.2 - 27.1\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"27.1 - 30\",\"Count\":1}]",
		},
		{
			data: map[string][]float64{
				"0": {1, 2, 3},
				"1": {3, 4, 5},
				"5": {10, 10, 10, 10, 10, 20, 30},
			},
			numBuckets:   30,
			decimalPlace: 5,
			result:       "[{\"Version\":\"0\",\"Bucket\":\"1 - 1.96666\",\"Count\":1},{\"Version\":\"0\",\"Bucket\":\"1.96666 - 2.93333\",\"Count\":1},{\"Version\":\"0\",\"Bucket\":\"2.93333 - 3.9\",\"Count\":1},{\"Version\":\"0\",\"Bucket\":\"3.9 - 4.86666\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"4.86666 - 5.83333\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"5.83333 - 6.8\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"6.8 - 7.76666\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"7.76666 - 8.73333\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"8.73333 - 9.69999\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"9.69999 - 10.66666\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"10.66666 - 11.63333\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"11.63333 - 12.6\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"12.6 - 13.56666\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"13.56666 - 14.53333\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"14.53333 - 15.5\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"15.5 - 16.46666\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"16.46666 - 17.43333\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"17.43333 - 18.39999\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"18.39999 - 19.36666\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"19.36666 - 20.33333\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"20.33333 - 21.3\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"21.3 - 22.26666\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"22.26666 - 23.23333\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"23.23333 - 24.2\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"24.2 - 25.16666\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"25.16666 - 26.13333\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"26.13333 - 27.1\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"27.1 - 28.06666\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"28.06666 - 29.03333\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"29.03333 - 30\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"1 - 1.96666\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"1.96666 - 2.93333\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"2.93333 - 3.9\",\"Count\":1},{\"Version\":\"1\",\"Bucket\":\"3.9 - 4.86666\",\"Count\":1},{\"Version\":\"1\",\"Bucket\":\"4.86666 - 5.83333\",\"Count\":1},{\"Version\":\"1\",\"Bucket\":\"5.83333 - 6.8\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"6.8 - 7.76666\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"7.76666 - 8.73333\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"8.73333 - 9.69999\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"9.69999 - 10.66666\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"10.66666 - 11.63333\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"11.63333 - 12.6\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"12.6 - 13.56666\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"13.56666 - 14.53333\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"14.53333 - 15.5\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"15.5 - 16.46666\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"16.46666 - 17.43333\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"17.43333 - 18.39999\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"18.39999 - 19.36666\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"19.36666 - 20.33333\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"20.33333 - 21.3\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"21.3 - 22.26666\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"22.26666 - 23.23333\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"23.23333 - 24.2\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"24.2 - 25.16666\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"25.16666 - 26.13333\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"26.13333 - 27.1\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"27.1 - 28.06666\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"28.06666 - 29.03333\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"29.03333 - 30\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"1 - 1.96666\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"1.96666 - 2.93333\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"2.93333 - 3.9\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"3.9 - 4.86666\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"4.86666 - 5.83333\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"5.83333 - 6.8\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"6.8 - 7.76666\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"7.76666 - 8.73333\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"8.73333 - 9.69999\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"9.69999 - 10.66666\",\"Count\":5},{\"Version\":\"5\",\"Bucket\":\"10.66666 - 11.63333\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"11.63333 - 12.6\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"12.6 - 13.56666\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"13.56666 - 14.53333\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"14.53333 - 15.5\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"15.5 - 16.46666\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"16.46666 - 17.43333\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"17.43333 - 18.39999\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"18.39999 - 19.36666\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"19.36666 - 20.33333\",\"Count\":1},{\"Version\":\"5\",\"Bucket\":\"20.33333 - 21.3\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"21.3 - 22.26666\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"22.26666 - 23.23333\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"23.23333 - 24.2\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"24.2 - 25.16666\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"25.16666 - 26.13333\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"26.13333 - 27.1\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"27.1 - 28.06666\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"28.06666 - 29.03333\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"29.03333 - 30\",\"Count\":1}]",
		},
		{
			data: map[string][]float64{
				"0": {1, 2, 3},
				"1": {3, 4, 5},
				"5": {10, 10, 10, 10, 10, 20, 30},
			}, numBuckets: 5,
			decimalPlace: 1,
			result:       "[{\"Version\":\"0\",\"Bucket\":\"1 - 6.8\",\"Count\":3},{\"Version\":\"0\",\"Bucket\":\"6.8 - 12.6\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"12.6 - 18.4\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"18.4 - 24.2\",\"Count\":0},{\"Version\":\"0\",\"Bucket\":\"24.2 - 30\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"1 - 6.8\",\"Count\":3},{\"Version\":\"1\",\"Bucket\":\"6.8 - 12.6\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"12.6 - 18.4\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"18.4 - 24.2\",\"Count\":0},{\"Version\":\"1\",\"Bucket\":\"24.2 - 30\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"1 - 6.8\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"6.8 - 12.6\",\"Count\":5},{\"Version\":\"5\",\"Bucket\":\"12.6 - 18.4\",\"Count\":0},{\"Version\":\"5\",\"Bucket\":\"18.4 - 24.2\",\"Count\":1},{\"Version\":\"5\",\"Bucket\":\"24.2 - 30\",\"Count\":1}]",
		},
	}

	for _, test := range tests {
		summarizedMetric, err := CalculateHistogram(test.data, test.numBuckets, test.decimalPlace)
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
