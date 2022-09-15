package base

import (
	"context"
	"encoding/json"
	"time"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	pb "github.com/iter8-tools/iter8/abn/grpc"
	k8sclient "github.com/iter8-tools/iter8/abn/k8sclient"
	log "github.com/iter8-tools/iter8/base/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"helm.sh/helm/v3/pkg/cli"
)

const (
	CollectABNMetrics = "abnmetrics"
	abnMetricProvider = "abn"
)

type ABNMetricsInputs struct {
	// Application is name of application to evaluate
	Application string `json:"application" yaml:"application"`
}

type collectABNMetricsTask struct {
	TaskMeta
	With ABNMetricsInputs `json:"with" yaml:"with"`
}

// initializeDefaults sets default values for the task
func (t *collectABNMetricsTask) initializeDefaults() {
	k8sclient.Client = *k8sclient.NewKubeClient(cli.New())
	k8sclient.Client.Initialize()
}

// validate task inputs
func (t *collectABNMetricsTask) validateInputs() error {
	return nil
}

// run exeuctes this task
func (t *collectABNMetricsTask) run(exp *Experiment) error {
	var err error

	// validate inputs
	err = t.validateInputs()
	if err != nil {
		return err
	}

	// initialize defaults
	t.initializeDefaults()

	// a, _ := abnapp.Applications.Get(t.With.Application, false)

	// setup client
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	// conn, err := grpc.Dial(lis.Addr().String(), opts...)
	conn, err := grpc.Dial("abn:50051", opts...)
	if err != nil {
		return err
	}
	c := pb.NewABNClient(conn)

	// get application
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := c.GetMetrics(
		ctx,
		&pb.MetricRequest{
			Application: t.With.Application,
		},
	)
	if err != nil {
		return err
	}
	applicationJson := s.GetApplicationJson()
	a := &abnapp.Application{}
	err = json.Unmarshal([]byte(applicationJson), a)
	if err != nil {
		return err
	}

	// count number of tracks
	numTracks := len(a.Tracks)
	if numTracks == 0 {
		log.Logger.Warn("no tracks detected in application")
		return nil
	}

	// initialize insights in Result with number of tracks
	err = exp.Result.initInsightsWithNumVersions(numTracks)
	if err != nil {
		return err
	}
	log.Logger.Tracef("intialized insights with %d versions", numTracks)

	in := exp.Result.Insights

	// add metrics for tracks
	versionIndex := 0
	for version, v := range a.Versions {
		t := a.GetTrack(version)
		if t != nil {
			log.Logger.Tracef("version %s is mapped to track %s; using index %d", version, *t, versionIndex)
			for metric, m := range v.Metrics {
				log.Logger.Tracef("   updating metric %s with data %+v", metric, m)
				in.updateMetric(
					abnMetricProvider+"/"+metric,
					MetricMeta{
						Description: "summary metric",
						Type:        SummaryMetricType,
					},
					versionIndex,
					m,
				)
			}
			versionIndex++
		}
	}

	return nil
}