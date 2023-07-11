package storageclient

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetClient(t *testing.T) {
	metricSummary, err := getTestGrafana(grafanaConfig{
		seed: 123,
		metricConfigs: map[string]metricConfig{
			"metric 0": {
				numBuckets:   10,
				decimalPlace: 1,
				versionConfigs: []testMetricVersionSummaryConfig{
					{
						numPoints: 20,
						mean:      5,
						stdDev:    3,
					},
					{
						numPoints: 15,
						mean:      10,
						stdDev:    5,
					},
					{
						numPoints: 10,
						mean:      7.5,
						stdDev:    2,
					},
				},
			},
			"metric 1": {
				numBuckets:   5,
				decimalPlace: 2,
				versionConfigs: []testMetricVersionSummaryConfig{
					{
						numPoints: 30,
						mean:      50,
						stdDev:    20,
					},
					{
						numPoints: 30,
						mean:      50,
						stdDev:    20,
					},
					{
						numPoints: 60,
						mean:      50,
						stdDev:    20,
					},
				},
			},
		},
	})
	assert.NoError(t, err)

	metricSummaryJSON, _ := json.Marshal(metricSummary)

	assert.Equal(t,
		"{\"metric 0\":{\"HistogramsOverTransactions\":[{\"Version\":\"0\",\"Bucket\":\"-7.7 - -5.2\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"-5.2 - -2.6\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"-2.6 - -0.1\",\"Value\":1},{\"Version\":\"0\",\"Bucket\":\"-0.1 - 2.5\",\"Value\":3},{\"Version\":\"0\",\"Bucket\":\"2.5 - 5\",\"Value\":5},{\"Version\":\"0\",\"Bucket\":\"5 - 7.6\",\"Value\":8},{\"Version\":\"0\",\"Bucket\":\"7.6 - 10.1\",\"Value\":3},{\"Version\":\"0\",\"Bucket\":\"10.1 - 12.6\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"12.6 - 15.2\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"15.2 - 17.7\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"-7.7 - -5.2\",\"Value\":1},{\"Version\":\"1\",\"Bucket\":\"-5.2 - -2.6\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"-2.6 - -0.1\",\"Value\":1},{\"Version\":\"1\",\"Bucket\":\"-0.1 - 2.5\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"2.5 - 5\",\"Value\":2},{\"Version\":\"1\",\"Bucket\":\"5 - 7.6\",\"Value\":1},{\"Version\":\"1\",\"Bucket\":\"7.6 - 10.1\",\"Value\":3},{\"Version\":\"1\",\"Bucket\":\"10.1 - 12.6\",\"Value\":1},{\"Version\":\"1\",\"Bucket\":\"12.6 - 15.2\",\"Value\":4},{\"Version\":\"1\",\"Bucket\":\"15.2 - 17.7\",\"Value\":2},{\"Version\":\"2\",\"Bucket\":\"-7.7 - -5.2\",\"Value\":0},{\"Version\":\"2\",\"Bucket\":\"-5.2 - -2.6\",\"Value\":0},{\"Version\":\"2\",\"Bucket\":\"-2.6 - -0.1\",\"Value\":0},{\"Version\":\"2\",\"Bucket\":\"-0.1 - 2.5\",\"Value\":0},{\"Version\":\"2\",\"Bucket\":\"2.5 - 5\",\"Value\":0},{\"Version\":\"2\",\"Bucket\":\"5 - 7.6\",\"Value\":4},{\"Version\":\"2\",\"Bucket\":\"7.6 - 10.1\",\"Value\":4},{\"Version\":\"2\",\"Bucket\":\"10.1 - 12.6\",\"Value\":2},{\"Version\":\"2\",\"Bucket\":\"12.6 - 15.2\",\"Value\":0},{\"Version\":\"2\",\"Bucket\":\"15.2 - 17.7\",\"Value\":0}],\"HistogramsOverUsers\":[{\"Version\":\"0\",\"Bucket\":\"-8 - -3.7\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"-3.7 - 0.5\",\"Value\":1},{\"Version\":\"0\",\"Bucket\":\"0.5 - 4.8\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"4.8 - 9.1\",\"Value\":2},{\"Version\":\"0\",\"Bucket\":\"9.1 - 13.3\",\"Value\":7},{\"Version\":\"0\",\"Bucket\":\"13.3 - 17.6\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"17.6 - 21.9\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"21.9 - 26.1\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"26.1 - 30.4\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"30.4 - 34.7\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"-8 - -3.7\",\"Value\":1},{\"Version\":\"1\",\"Bucket\":\"-3.7 - 0.5\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"0.5 - 4.8\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"4.8 - 9.1\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"9.1 - 13.3\",\"Value\":3},{\"Version\":\"1\",\"Bucket\":\"13.3 - 17.6\",\"Value\":0},{\"Version\":\"1\",\"Bucket\":\"17.6 - 21.9\",\"Value\":1},{\"Version\":\"1\",\"Bucket\":\"21.9 - 26.1\",\"Value\":1},{\"Version\":\"1\",\"Bucket\":\"26.1 - 30.4\",\"Value\":1},{\"Version\":\"1\",\"Bucket\":\"30.4 - 34.7\",\"Value\":1},{\"Version\":\"2\",\"Bucket\":\"-8 - -3.7\",\"Value\":0},{\"Version\":\"2\",\"Bucket\":\"-3.7 - 0.5\",\"Value\":0},{\"Version\":\"2\",\"Bucket\":\"0.5 - 4.8\",\"Value\":0},{\"Version\":\"2\",\"Bucket\":\"4.8 - 9.1\",\"Value\":0},{\"Version\":\"2\",\"Bucket\":\"9.1 - 13.3\",\"Value\":1},{\"Version\":\"2\",\"Bucket\":\"13.3 - 17.6\",\"Value\":3},{\"Version\":\"2\",\"Bucket\":\"17.6 - 21.9\",\"Value\":1},{\"Version\":\"2\",\"Bucket\":\"21.9 - 26.1\",\"Value\":0},{\"Version\":\"2\",\"Bucket\":\"26.1 - 30.4\",\"Value\":0},{\"Version\":\"2\",\"Bucket\":\"30.4 - 34.7\",\"Value\":0}],\"SummaryOverTransactions\":[{\"Version\":0,\"Count\":20,\"Mean\":4.8023885851534525,\"StdDev\":2.4947543459307813,\"Min\":-0.4889230723160627,\"Max\":7.9524206496018515},{\"Version\":1,\"Count\":15,\"Mean\":8.859880972453066,\"StdDev\":6.617864445244916,\"Min\":-7.648554418856204,\"Max\":17.774405961730672},{\"Version\":2,\"Count\":10,\"Mean\":8.061186259617305,\"StdDev\":2.139641430949272,\"Min\":5.434443321976806,\"Max\":11.873716954373307}],\"SummaryOverUsers\":[{\"Version\":0,\"Count\":10,\"Mean\":9.604777170306903,\"StdDev\":3.579794217677512,\"Min\":0.06131593953661474,\"Max\":13.297339846819604},{\"Version\":1,\"Count\":8,\"Mean\":16.6122768233495,\"StdDev\":12.217159381735678,\"Min\":-7.939341183662007,\"Max\":34.71695240678541},{\"Version\":2,\"Count\":5,\"Mean\":16.12237251923461,\"StdDev\":2.842086834214423,\"Min\":13.241121546087511,\"Max\":21.09455029055873}]},\"metric 1\":{\"HistogramsOverTransactions\":[{\"Version\":\"0\",\"Bucket\":\"-6.83 - 13.58\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"13.58 - 33.98\",\"Value\":2},{\"Version\":\"0\",\"Bucket\":\"33.98 - 54.38\",\"Value\":12},{\"Version\":\"0\",\"Bucket\":\"54.38 - 74.78\",\"Value\":11},{\"Version\":\"0\",\"Bucket\":\"74.78 - 95.19\",\"Value\":5},{\"Version\":\"1\",\"Bucket\":\"-6.83 - 13.58\",\"Value\":4},{\"Version\":\"1\",\"Bucket\":\"13.58 - 33.98\",\"Value\":5},{\"Version\":\"1\",\"Bucket\":\"33.98 - 54.38\",\"Value\":6},{\"Version\":\"1\",\"Bucket\":\"54.38 - 74.78\",\"Value\":14},{\"Version\":\"1\",\"Bucket\":\"74.78 - 95.19\",\"Value\":1},{\"Version\":\"2\",\"Bucket\":\"-6.83 - 13.58\",\"Value\":0},{\"Version\":\"2\",\"Bucket\":\"13.58 - 33.98\",\"Value\":9},{\"Version\":\"2\",\"Bucket\":\"33.98 - 54.38\",\"Value\":31},{\"Version\":\"2\",\"Bucket\":\"54.38 - 74.78\",\"Value\":18},{\"Version\":\"2\",\"Bucket\":\"74.78 - 95.19\",\"Value\":2}],\"HistogramsOverUsers\":[{\"Version\":\"0\",\"Bucket\":\"59.92 - 79.74\",\"Value\":0},{\"Version\":\"0\",\"Bucket\":\"79.74 - 99.57\",\"Value\":4},{\"Version\":\"0\",\"Bucket\":\"99.57 - 119.39\",\"Value\":6},{\"Version\":\"0\",\"Bucket\":\"119.39 - 139.22\",\"Value\":3},{\"Version\":\"0\",\"Bucket\":\"139.22 - 159.04\",\"Value\":2},{\"Version\":\"1\",\"Bucket\":\"59.92 - 79.74\",\"Value\":5},{\"Version\":\"1\",\"Bucket\":\"79.74 - 99.57\",\"Value\":3},{\"Version\":\"1\",\"Bucket\":\"99.57 - 119.39\",\"Value\":4},{\"Version\":\"1\",\"Bucket\":\"119.39 - 139.22\",\"Value\":3},{\"Version\":\"1\",\"Bucket\":\"139.22 - 159.04\",\"Value\":0},{\"Version\":\"2\",\"Bucket\":\"59.92 - 79.74\",\"Value\":5},{\"Version\":\"2\",\"Bucket\":\"79.74 - 99.57\",\"Value\":10},{\"Version\":\"2\",\"Bucket\":\"99.57 - 119.39\",\"Value\":9},{\"Version\":\"2\",\"Bucket\":\"119.39 - 139.22\",\"Value\":4},{\"Version\":\"2\",\"Bucket\":\"139.22 - 159.04\",\"Value\":2}],\"SummaryOverTransactions\":[{\"Version\":0,\"Count\":30,\"Mean\":56.61644418342632,\"StdDev\":17.734707008601955,\"Min\":26.200385147928095,\"Max\":95.19106919208639},{\"Version\":1,\"Count\":30,\"Mean\":46.65607577599676,\"StdDev\":24.53958623244361,\"Min\":-6.821139542819239,\"Max\":93.39617801208192},{\"Version\":2,\"Count\":60,\"Mean\":50.26022010530817,\"StdDev\":14.936706739289914,\"Min\":21.38467602511753,\"Max\":90.7177851576927}],\"SummaryOverUsers\":[{\"Version\":0,\"Count\":15,\"Mean\":113.23288836685265,\"StdDev\":20.481124666301522,\"Min\":80.55520335809948,\"Max\":159.04931385850972},{\"Version\":1,\"Count\":15,\"Mean\":93.31215155199354,\"StdDev\":22.830723755797994,\"Min\":62.71842850937553,\"Max\":130.11467527182526},{\"Version\":2,\"Count\":30,\"Mean\":100.52044021061634,\"StdDev\":23.008216401436634,\"Min\":59.92056660370633,\"Max\":149.2391636497419}]}}",
		string(metricSummaryJSON),
	)
}
