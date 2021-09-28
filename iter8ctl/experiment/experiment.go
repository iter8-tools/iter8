// Package experiment enables extraction of useful information from experiment objects and their formatting.
package experiment

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	tasks "github.com/iter8-tools/etc3/taskrunner/core"
	"github.com/sirupsen/logrus"
	"gopkg.in/inf.v0"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var log *logrus.Logger

func init() {
	log = tasks.GetLogger()
}

// Experiment is an enhancement of v2alpha2.Experiment struct, and supports various methods used in describing an experiment.
type Experiment struct {
	v2alpha2.Experiment
}

// ConditionType is a type for conditions that can be asserted
type ConditionType string

const (
	// Completed implies experiment is complete
	Completed ConditionType = "completed"
	// Successful     ConditionType = "successful"
	// Failure        ConditionType = "failure"
	// HandlerFailure ConditionType = "handlerFailure"

	// WinnerFound implies experiment has found a winner
	WinnerFound ConditionType = "winnerFound"
	// CandidateWon   ConditionType = "candidateWon"
	// BaselineWon    ConditionType = "baselineWon"
	// NoWinner       ConditionType = "noWinner"
)

// for mocking in tests
var k8sClient client.Client

// GetConfig variable is useful for test mocks.
var GetConfig = func() (*rest.Config, error) {
	return config.GetConfig()
}

// GetClient constructs and returns a K8s client.
// The returned client has experiment types registered.
var GetClient = func() (rc client.Client, err error) {
	var restConf *rest.Config
	restConf, err = GetConfig()
	if err != nil {
		return nil, err
	}

	var addKnownTypes = func(scheme *runtime.Scheme) error {
		// register iter8.GroupVersion and type
		metav1.AddToGroupVersion(scheme, v2alpha2.GroupVersion)
		scheme.AddKnownTypes(v2alpha2.GroupVersion, &v2alpha2.Experiment{})
		scheme.AddKnownTypes(v2alpha2.GroupVersion, &v2alpha2.ExperimentList{})
		return nil
	}

	var schemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	scheme := runtime.NewScheme()
	err = schemeBuilder.AddToScheme(scheme)

	if err == nil {
		rc, err = client.New(restConf, client.Options{
			Scheme: scheme,
		})
		if err == nil {
			return rc, nil
		}
	}
	return nil, errors.New("cannot get client using rest config")
}

// GetExperiment gets the experiment from cluster
func GetExperiment(latest bool, name string, namespace string) (*Experiment, error) {
	results := v2alpha2.ExperimentList{}
	var exp *v2alpha2.Experiment
	var err error

	ns := namespace
	if ns == "" {
		ns, err = getNamespaceFromCurrentContext()
		if err != nil {
			log.Warn("Unable to get namespace from current context: " + err.Error())
			ns = "default"
		}
	}
	// log.Infof("latest: %t", latest)
	// log.Infof("using namespace: %s", ns)

	// get all experiments
	var rc client.Client
	if rc, err = GetClient(); err == nil {
		err = rc.List(context.Background(), &results, &client.ListOptions{Namespace: ns})
	}

	// get latest experiment
	if latest && err == nil {
		if len(results.Items) > 0 {
			exp = &results.Items[len(results.Items)-1]
		} else {
			err = errors.New("no experiments found in cluster")
		}
	}

	// get named experiment
	if !latest && err == nil {
		for i := range results.Items {
			if results.Items[i].Name == name && results.Items[i].Namespace == ns {
				exp = &results.Items[i]
				break
			}
		}
		if exp == nil {
			err = errors.New("Experiment " + name + " not found in namespace " + ns)
		}
	}

	// return error
	if err != nil {
		return nil, err
	}

	// Return experiment
	return &Experiment{
		*exp,
	}, nil
}

// Started indicates if at least one iteration of the experiment has completed.
func (e *Experiment) Started() bool {
	if e == nil {
		return false
	}
	c := e.Status.CompletedIterations
	return c != nil && *c > 0
}

// Completed indicates if the experiment has completed.
func (e *Experiment) Completed() bool {
	if e == nil {
		return false
	}
	c := e.Status.GetCondition(v2alpha2.ExperimentConditionExperimentCompleted)
	return c != nil && c.IsTrue()
}

// WinnerFound indicates if the experiment has found a winning version (winner).
func (e *Experiment) WinnerFound() bool {
	if e == nil {
		return false
	}
	if a := e.Status.Analysis; a != nil {
		if w := a.WinnerAssessment; w != nil {
			return w.Data.WinnerFound
		}
	}
	return false
}

// GetVersions returns the slice of version name strings. If the VersionInfo section is not present in the experiment's spec, then this slice is empty.
func (e *Experiment) GetVersions() []string {
	if e.Spec.VersionInfo == nil {
		return nil
	}
	versions := []string{e.Spec.VersionInfo.Baseline.Name}
	for _, c := range e.Spec.VersionInfo.Candidates {
		versions = append(versions, c.Name)
	}
	return versions
}

// GetMetricStr returns the metric value as a string for a given metric and a given version.
func (e *Experiment) GetMetricStr(metric string, version string) string {
	am := e.Status.Analysis.AggregatedMetrics
	if am == nil {
		return "unavailable"
	}
	if vals, ok := am.Data[metric]; ok {
		if val, ok := vals.Data[version]; ok {
			if val.Value != nil {
				z := new(inf.Dec).Round(val.Value.AsDec(), 3, inf.RoundCeil)
				return z.String()
			}
		}
	}
	return "unavailable"
}

// GetMetricStrs returns the given metric's value as a slice of strings, whose elements correspond to versions.
func (e *Experiment) GetMetricStrs(metric string) []string {
	versions := e.GetVersions()
	reqs := make([]string, len(versions))
	for i, v := range versions {
		reqs[i] = e.GetMetricStr(metric, v)
	}
	return reqs
}

// GetMetricNameAndUnits extracts the name, and if specified, units for the given metricInfo object and combines them into a string.
func GetMetricNameAndUnits(metricInfo v2alpha2.MetricInfo) string {
	r := metricInfo.Name
	if metricInfo.MetricObj.Spec.Units != nil {
		r += fmt.Sprintf(" (" + *metricInfo.MetricObj.Spec.Units + ")")
	}
	return r
}

// StringifyObjective returns a string representation of the given objective.
func StringifyObjective(objective v2alpha2.Objective) string {
	r := ""
	if objective.LowerLimit != nil {
		z := new(inf.Dec).Round(objective.LowerLimit.AsDec(), 3, inf.RoundCeil)
		r += z.String() + " <= "
	}
	r += objective.Metric
	if objective.UpperLimit != nil {
		z := new(inf.Dec).Round(objective.UpperLimit.AsDec(), 3, inf.RoundCeil)
		r += " <= " + z.String()
	}
	return r
}

// GetSatisfyStr returns a true/false/unavailable valued string denotating if a version satisfies the objective.
func (e *Experiment) GetSatisfyStr(objectiveIndex int, version string) string {
	ana := e.Status.Analysis
	if ana == nil {
		return "unavailable"
	}
	va := ana.VersionAssessments
	if va == nil {
		return "unavailable"
	}
	if vals, ok := va.Data[version]; ok {
		if len(vals) > objectiveIndex {
			return fmt.Sprintf("%v", vals[objectiveIndex])
		}
	}
	return "unavailable"
}

// GetSatisfyStrs returns a slice of true/false/unavailable valued strings for an objective denoting if it is satisfied by versions.
func (e *Experiment) GetSatisfyStrs(objectiveIndex int) []string {
	versions := e.GetVersions()
	sat := make([]string, len(versions))
	for i, v := range versions {
		sat[i] = e.GetSatisfyStr(objectiveIndex, v)
	}
	return sat
}

// StringifyReward returns a string representation of the given reward.
func StringifyReward(reward v2alpha2.Reward) string {
	r := ""
	r += reward.Metric
	if reward.PreferredDirection == v2alpha2.PreferredDirectionHigher {
		r += " (higher better)"
	} else {
		r += " (lower better)"
	}
	return r
}

// GetMetricDec returns the metric value as a string for a given metric and a given version.
func (e *Experiment) GetMetricDec(metric string, version string) *inf.Dec {
	am := e.Status.Analysis.AggregatedMetrics
	if am == nil {
		return nil
	}
	if vals, ok := am.Data[metric]; ok {
		if val, ok := vals.Data[version]; ok {
			if val.Value != nil {
				z := new(inf.Dec).Round(val.Value.AsDec(), 3, inf.RoundCeil)
				return z
			}
		}
	}
	return nil
}

// GetAnnotatedMetricStrs returns a slice of values for a reward
func (e *Experiment) GetAnnotatedMetricStrs(reward v2alpha2.Reward) []string {
	versions := e.GetVersions()
	row := make([]string, len(versions))
	var currentBestIndex *int
	var currentBestValue *inf.Dec
	for i, v := range versions {
		val := e.GetMetricDec(reward.Metric, v)

		if val == nil {
			row[i] = "unavailable"
			continue
		}

		row[i] = val.String()

		// set currentBest if not already set
		if currentBestIndex == nil {
			currentBestIndex, currentBestValue = &i, val
			continue
		}

		// update currentBest

		if reward.PreferredDirection == v2alpha2.PreferredDirectionHigher {
			if -1 == currentBestValue.Cmp(val) {
				currentBestIndex, currentBestValue = &i, val
			}
			continue
		}

		// reward.PreferredDirection == v2alpha2.PreferredDirectionLower
		if currentBestValue.Cmp(val) == 1 {
			currentBestIndex, currentBestValue = &i, val
		}
	}

	// mark current best with '*'
	if currentBestIndex != nil {
		row[*currentBestIndex] = row[*currentBestIndex] + " *"
	}
	return row
}

// Assert verifies a given set of conditions for the experiment.
func (e *Experiment) Assert(conditions []ConditionType) error {
	for _, cond := range conditions {
		switch cond {
		case Completed:
			if !e.Completed() {
				return errors.New("experiment has not completed")
			}
		case WinnerFound:
			if !e.WinnerFound() {
				return errors.New("no winner found in experiment")
			}
		default:
			return errors.New("unsupported condition found in assertion")
		}
	}
	return nil
}

// Methods to get namespace defined in current context
// These are inspired by the methods to get the config in sigs.k8s.io/controller-runtime/pkg/client/config
// Namespace does not seem to be otherwise readily available

// getNamespaceFromCurrentContext gets the namespace specified in the current kubernetes context
func getNamespaceFromCurrentContext() (string, error) {
	// start -- the following should be included if we support a --kubeconfg flag
	// // If a flag is specified with the config location, use that
	// if len(kubeconfig) > 0 {
	// 	return loadNamespaceWithContext("", &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}, context)
	// }
	// end -- support --kubeconfig

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if _, ok := os.LookupEnv("HOME"); !ok {
		u, err := user.Current()
		if err != nil {
			return "default", fmt.Errorf("could not get current user: %v", err)
		}
		loadingRules.Precedence = append(loadingRules.Precedence, path.Join(u.HomeDir, clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName))
	}

	return getNamespaceWithContext("", loadingRules)
}

func getNamespaceWithContext(apiServerURL string, loader clientcmd.ClientConfigLoader) (string, error) {
	ns, _, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loader,
		&clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{
				Server: apiServerURL,
			},
			CurrentContext: "",
		}).Namespace()
	return ns, err
}
