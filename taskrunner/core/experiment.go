package core

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Experiment is an enhancement of v2alpha2.Experiment struct with useful methods.
type Experiment struct {
	v2alpha2.Experiment
}

// Builder helps in construction of an experiment.
type Builder struct {
	err error
	exp *Experiment
}

// Build returns the built experiment or error.
// Must call FromFile or FromCluster on b prior to invoking Build.
func (b *Builder) Build() (*Experiment, error) {
	log.Trace(b)
	return b.exp, b.err
}

// GetExperimentFromContext gets the experiment object from given context.
func GetExperimentFromContext(ctx context.Context) (*Experiment, error) {
	//	ctx := context.WithValue(context.Background(), base.ContextKey("experiment"), e)
	if v := ctx.Value(ContextKey("experiment")); v != nil {
		log.Debug("found experiment")
		var e *Experiment
		var ok bool
		if e, ok = v.(*Experiment); !ok {
			return nil, errors.New("context has experiment value with wrong type")
		}
		return e, nil
	}
	return nil, errors.New("context has no experiment key")
}

// GetActionStringFromContext gets the action string from given context.
func GetActionStringFromContext(ctx context.Context) (string, error) {
	if v := ctx.Value(ContextKey("action")); v != nil {
		log.Debug("found action")
		var a string
		var ok bool
		if a, ok = v.(string); !ok {
			return "", errors.New("context has action value with wrong type")
		}
		return a, nil
	}
	return "", errors.New("context has no action key")
}

// Interpolate interpolates input arguments based on tags of the version recommended for promotion in the experiment.
// DEPRECATED. Use tags.Interpolate in base package instead
func (exp *Experiment) Interpolate(inputArgs []string) ([]string, error) {
	var recommendedBaseline string
	var args []string
	var err error
	if recommendedBaseline, err = exp.GetVersionRecommendedForPromotion(); err == nil {
		var versionDetail *v2alpha2.VersionDetail
		if versionDetail, err = exp.GetVersionDetail(recommendedBaseline); err == nil {
			// get the tags
			tags := Tags{M: make(map[string]interface{})}
			tags.M["name"] = versionDetail.Name
			for i := 0; i < len(versionDetail.Variables); i++ {
				tags.M[versionDetail.Variables[i].Name] = versionDetail.Variables[i].Value
			}
			log.Trace(tags)
			args = make([]string, len(inputArgs))
			for i := 0; i < len(args); i++ {
				if args[i], err = tags.Interpolate(&inputArgs[i]); err != nil {
					break
				}
				log.Trace("input arg: ", inputArgs[i], " interpolated arg: ", args[i])
			}
		}
	}
	return args, err
}

// ToMap converts exp.Experiment to  a map[string]interface{}
func (exp *Experiment) ToMap() (map[string]interface{}, error) {
	// convert unstructured object to JSON object
	expJSON, err := json.Marshal(exp.Experiment)
	if err != nil {
		log.Error(err, "Unable to convert experiment to JSON")
		return nil, err
	}

	// convert JSON object to Go map
	expObj := make(map[string]interface{})
	err = json.Unmarshal(expJSON, &expObj)
	if err != nil {
		log.Error(err, "Unable to convert JSON to object")
		return nil, err
	}
	return expObj, nil
}

// WinnerFound returns true if Experiment found a winner
func (exp *Experiment) WinnerFound() bool {
	if exp != nil {
		if exp.Status.Analysis != nil {
			if exp.Status.Analysis.WinnerAssessment != nil {
				return exp.Status.Analysis.WinnerAssessment.Data.WinnerFound
			}
		}
	}
	return false
}

// CandidateWon returns true if candidate won in the experiment
func (exp *Experiment) CandidateWon() bool {
	if exp.WinnerFound() {
		if len(exp.Spec.VersionInfo.Candidates) == 1 {
			return exp.Spec.VersionInfo.Candidates[0].Name == *exp.Status.Analysis.WinnerAssessment.Data.Winner
		}
	}
	return false
}

// GetSecret retrieves a secret from the kubernetes cluster
func GetSecret(namespacedname string) (*corev1.Secret, error) {
	// get secret namespace and name
	namespace := viper.GetViper().GetString("experiment_namespace")
	var name string
	nn := strings.Split(namespacedname, "/")
	if len(nn) == 1 {
		name = nn[0]
	} else {
		namespace = nn[0]
		name = nn[1]
	}
	log.Trace("retrieving secret: ", namespace, "/", name)

	secret := corev1.Secret{}
	log.Trace("Getting secret. ", "Namespace: ", namespace, " Name: ", name)
	err := GetTypedObject(&types.NamespacedName{Namespace: namespace, Name: name}, &secret)
	return &secret, err
}
