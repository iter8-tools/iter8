package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/iter8-tools/iter8/abn"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/controllers"
	"github.com/iter8-tools/iter8/storage"
)

// Start starts the HTTP server
func Start() error {
	http.HandleFunc("/metrics", getMetrics)
	// http.HandleFunc("/summarymetrics", getSummaryMetrics)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Logger.Errorf("unable to start metrics service: %s", err.Error())
		return err
	}
	return nil
}

type VersionMetrics struct {
	*storage.VersionMetricSummary
	// *storage.VersionMetrics
}

// getMetrics handles POST /metrics
func getMetrics(w http.ResponseWriter, r *http.Request) {
	// verify method
	if r.Method != http.MethodGet {
		http.Error(w, "expected GET", http.StatusMethodNotAllowed)
		return
	}

	// verify request (query parameter)
	application := r.URL.Query().Get("application")
	if application == "" {
		http.Error(w, "no application specified", http.StatusBadRequest)
	}

	// identify the routemap for the application
	namespace, name := splitApplicationKey(application)
	rm := controllers.AllRoutemaps.GetRoutemapFromNamespaceName(namespace, name)
	if rm == nil {
		http.Error(w, fmt.Sprintf("unknown application %s", application), http.StatusBadRequest)
		return
	}

	// get summary metrics for each version
	result := make(map[string]*VersionMetrics, len(rm.GetVersions()))
	for v, version := range rm.GetVersions() {
		if version.GetSignature() == nil {
			result[fmt.Sprintf("%d", v)].VersionMetricSummary = nil

			continue
		}
		r, err := abn.MetricsClient.GetSummaryMetrics(application, v, *version.GetSignature())
		if err != nil {
			log.Logger.Debugf("no summary metrics for %s, %d, %s", application, v, *version.GetSignature())
		}
		result[fmt.Sprintf("%d", v)].VersionMetricSummary = r
	}

	// convert to JSON
	b, err := json.MarshalIndent(result, "", "   ")
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to create JSON response %s", string(b)), http.StatusInternalServerError)
		return
	}

	// finally, send response
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(b)
}

// splitApplicationKey is a utility function that returns the name and namespace from a key of the form "namespace/name"
func splitApplicationKey(applicationKey string) (string, string) {
	var name, namespace string
	names := strings.Split(applicationKey, "/")
	if len(names) > 1 {
		namespace, name = names[0], names[1]
	} else {
		namespace, name = "default", names[0]
	}

	return namespace, name
}
