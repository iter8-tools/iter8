package collect

import (
	"testing"

	"github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/stretchr/testify/assert"
)

func TestInitializeCollectDefaults(t *testing.T) {
	ct := CollectTask{
		TaskMeta: core.TaskMeta{
			Task: core.StringPointer(TaskName),
		},
		With: CollectInputs{
			Versions: []Version{{
				Name: "default",
				URL:  "https://httpbin.org",
			}},
		},
	}
	ct.InitializeDefaults()
	assert.Equal(t, uint32(100), *ct.With.NumQueries)
	assert.Equal(t, core.Float32Pointer(8.0), ct.With.QPS)
}

func TestAggregate(t *testing.T) {
	oldResults := map[string]*Result{}
	res := Result{
		DurationHistogram: DurationHist{
			Count: 21,
			Max:   100.0,
			Sum:   600.01,
			Data: []DurationSample{
				{
					Start: 10.0,
					End:   50.0,
					Count: 18,
				},
				{
					Start: 50.0,
					End:   100.0,
					Count: 3,
				},
			},
		},
		RetCodes: map[string]int{
			"200": 20,
			"400": 1,
		},
	}
	o := aggregate(oldResults, "v1", &res)
	assert.NotEmpty(t, o)

	oldResults = nil
	o = aggregate(oldResults, "v1", &res)
	assert.NotEmpty(t, o)

	res2 := res
	res2.DurationHistogram.Count += 2
	res2.DurationHistogram.Max = 200.0
	res2.DurationHistogram.Data = []DurationSample{
		{
			Start: 15.0,
			End:   26.0,
			Count: 5,
		},
	}
	res2.RetCodes = map[string]int{
		"200": 10,
		"400": 9,
		"500": 2,
	}

	u := aggregate(o, "v1", &res2)

	assert.NotEmpty(t, u)
	assert.Equal(t, 44, u["v1"].DurationHistogram.Count)
	assert.Equal(t, 200.0, u["v1"].DurationHistogram.Max)
	assert.Equal(t, 30, u["v1"].RetCodes["200"])
	assert.Equal(t, 10, u["v1"].RetCodes["400"])
	assert.Equal(t, 2, u["v1"].RetCodes["500"])
	assert.Equal(t, 3, len(u["v1"].DurationHistogram.Data))

}

func TestGetResultFromFile(t *testing.T) {
	fileName := core.CompletePath("../../", "testdata/metricscollect/fortiooutput.json")
	res, err := getResultFromFile(fileName)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	fileName = core.CompletePath("../../", "testdata/metricscollect/nooutput.json")
	res, err = getResultFromFile(fileName)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestPayloadFile(t *testing.T) {
	url := "https://httpbin.org/stream/1"
	fileName, err := payloadFile(url)
	assert.NoError(t, err)
	assert.NotEmpty(t, fileName)

	url = "https://httpbin.org/undef"
	fileName, err = payloadFile(url)
	assert.Error(t, err)
	assert.Empty(t, fileName)
}

func TestResultForVersion(t *testing.T) {
	ct := CollectTask{
		TaskMeta: core.TaskMeta{
			Task: core.StringPointer(TaskName),
		},
		With: CollectInputs{
			Versions: []Version{{
				Name: "default",
				URL:  "https://httpbin.org",
			}},
		},
	}
	ct.InitializeDefaults()
	res, err := ct.resultForVersion(0)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}
