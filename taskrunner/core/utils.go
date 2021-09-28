package core

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	controllers "github.com/iter8-tools/etc3/controllers"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

var log *logrus.Logger

var logLevel logrus.Level = logrus.InfoLevel

// SetLogLevel sets level for logging.
func SetLogLevel(l logrus.Level) {
	logLevel = l
	if log != nil {
		log.SetLevel(logLevel)
	}
}

// GetLogger returns a logger, if needed after creating it.
func GetLogger() *logrus.Logger {
	if log == nil {
		log = logrus.New()
		log.SetLevel(logLevel)
		log.SetFormatter(&logrus.TextFormatter{
			DisableQuote: true,
		})
	}
	return log
}

// Iter8Logger type objects are functions that can be used to print Iter8Logs within tasks
type Iter8Logger func(expName string, expNamespace string, priority controllers.Iter8LogPriority, taskName string, msg string, prec int) string

// // GetIter8Logger returns a logger that prints Iter8logs
// func GetIter8Logger() Iter8Logger {
// 	return func(expName string, expNamespace string, priority controllers.Iter8LogPriority, taskName string, msg string, prec int) string {
// 		// create Iter8log struct
// 		il := controllers.Iter8Log{
// 			IsIter8Log:          true,
// 			ExperimentName:      expName,
// 			ExperimentNamespace: expNamespace,
// 			Source:              "taskrunner",
// 			Priority:            priority,
// 			Message:             fmt.Sprintf("from task %v; %v", taskName, msg),
// 			Precedence:          prec,
// 		}

// 		return il.JSON()
// 	}
// }

// GetIter8LogPrecedence returns the precedence value to be used in an Ite8log
func GetIter8LogPrecedence(exp *Experiment, action string) int {
	loopCount := int32(0)
	ipl := v2alpha2.DefaultIterationsPerLoop
	if exp.Spec.Duration != nil &&
		exp.Spec.Duration.IterationsPerLoop != nil {
		ipl = *exp.Spec.Duration.IterationsPerLoop
	}
	if exp.Status.CompletedIterations != nil {
		loopCount = (*exp.Status.CompletedIterations) / ipl
	}
	if action == "start" {
		return 0
	} else if action == "loop" {
		return int(loopCount + 1)
	} else { // this better be the finish action
		return int(loopCount + 1)
	}
}

// ContextKey is the type of key that will be used to index into context.
type ContextKey string

// CompletePath determines complete path of a file
var CompletePath func(prefix string, suffix string) string = controllers.CompletePath

// UInt32Pointer takes a uint32 as input, creates a new variable with the input value, and returns a pointer to the variable
func UInt32Pointer(u uint32) *uint32 {
	return &u
}

// Int32Pointer takes an int32 as input, creates a new variable with the input value, and returns a pointer to the variable
func Int32Pointer(i int32) *int32 {
	return &i
}

// Float32Pointer takes an float32 as input, creates a new variable with the input value, and returns a pointer to the variable
func Float32Pointer(f float32) *float32 {
	return &f
}

// Float64Pointer takes an float64 as input, creates a new variable with the input value, and returns a pointer to the variable
func Float64Pointer(f float64) *float64 {
	return &f
}

// StringPointer takes a string as input, creates a new variable with the input value, and returns a pointer to the variable
func StringPointer(s string) *string {
	return &s
}

// BoolPointer takes a bool as input, creates a new variable with the input value, and returns a pointer to the variable
func BoolPointer(b bool) *bool {
	return &b
}

// HTTPMethod is either GET or POST
type HTTPMethod string

const (
	// GET method
	GET HTTPMethod = "GET"
	// POST method
	POST = "POST"
)

// HTTPMethodPointer takes an HTTPMethod as input, creates a new variable with the input value, and returns a pointer to the variable
func HTTPMethodPointer(h HTTPMethod) *HTTPMethod {
	return &h
}

// GetPayloadBytes downloads payload from URL and returns a byte slice
func GetPayloadBytes(url string) ([]byte, error) {
	var myClient = &http.Client{Timeout: 10 * time.Second}
	r, err := myClient.Get(url)
	if err != nil || r.StatusCode >= 400 {
		return nil, errors.New("error while fetching payload")
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	return body, err
}

// GetTokenFromSecret gets token from k8s secret object
// can be used in notification, gitops and other tasks that use secret tokens
func GetTokenFromSecret(secret *corev1.Secret) (string, error) {
	token := string(secret.Data["token"])
	if token == "" {
		return "", errors.New("empty token in secret")
	}
	return token, nil
}
