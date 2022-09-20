package base

import (
	"context"
	"errors"
	"time"

	abnapp "github.com/iter8-tools/iter8/abn/application"
	pb "github.com/iter8-tools/iter8/abn/grpc"
	k8sclient "github.com/iter8-tools/iter8/abn/k8sclient"
	log "github.com/iter8-tools/iter8/base/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// abnClientInterface is interface for calling gRPC services
type abnClientInterface interface {
	callGetApplicationJSON(appName string) (string, error)
}

// defaultABNClient is default implementation of interface that calls the service
type defaultABNClient struct {
	endpoint string
}

func (wc *defaultABNClient) callGetApplicationJSON(appName string) (string, error) {
	// setup client
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(wc.endpoint, opts...)
	if err != nil {
		return "", err
	}
	c := pb.NewABNClient(conn)

	// get application
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := c.GetApplicationData(
		ctx,
		&pb.ApplicationRequest{
			Application: appName,
		},
	)
	if err != nil {
		return "", err
	}
	applicationJSON := s.GetApplicationJson()

	return applicationJSON, nil
}

const (
	// CollectABNMetrics is the name of the task to read A/B(/n) metric values
	CollectABNMetrics = "abnmetrics"
	// abnMetricProvider is prefix for abn metrics
	abnMetricProvider = "abn"
)

// ABNMetricsInputs is the inputs for for the abnmetrics task
type ABNMetricsInputs struct {
	// Application is name of application to evaluate
	Application string `json:"application" yaml:"application"`
}

// collectABNMetricsTask is task defintion for abnmetrics task
type collectABNMetricsTask struct {
	TaskMeta
	With   ABNMetricsInputs `json:"with" yaml:"with"`
	client abnClientInterface
}

// initializeDefaults sets default values for the task
func (t *collectABNMetricsTask) initializeDefaults() {
	k8sclient.Client = *k8sclient.NewKubeClient(cli.New())
	if err := k8sclient.Client.Initialize(); err != nil {
		log.Logger.WithStackTrace("unable to initialize k8s client").Fatal(err)
	}

	if t.client == nil {
		t.client = &defaultABNClient{
			endpoint: "iter8-abn:50051",
		}
	}
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

	// get application json from abn service
	applicationJSON, err := t.client.callGetApplicationJSON(t.With.Application)
	if err != nil {
		return err
	}

	// convert to Application
	a := &abnapp.Application{}
	err = yaml.Unmarshal([]byte(applicationJSON), a)
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
	in.VersionNames = make([]string, in.NumVersions)
	for track, version := range a.Tracks {
		in.VersionNames[versionIndex] = track
		v, _ := a.GetVersion(version, false)
		if v == nil {
			log.Logger.Debugf("expected version %s not found", version)
			return errors.New("expected version not found")
		}
		log.Logger.Tracef("version %s is mapped to track %s; using index %d", version, track, versionIndex)
		for metric, m := range v.Metrics {
			log.Logger.Tracef("   updating metric %s with data %+v", metric, m)
			err := in.updateMetric(
				abnMetricProvider+"/"+metric,
				MetricMeta{
					Description: "summary metric",
					Type:        SummaryMetricType,
				},
				versionIndex,
				m,
			)
			if err != nil {
				return err
			}
		}
		versionIndex++
	}

	return nil
}
