package core

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

// TestingPatternType identifies the type of experiment type
type TestingPatternType string

const (
	// TestingPatternSLOValidation indicates an experiment tests for SLO validation
	TestingPatternSLOValidation TestingPatternType = "SLOValidation"

	// TestingPatternAB indicates an experiment is a A/B experiment
	TestingPatternAB TestingPatternType = "A/B"

	// TestingPatternABN indicates an experiment is a A/B/n experiment
	TestingPatternABN TestingPatternType = "A/B/N"

	// TestingPatternHybridAB indicates an experiment is a Hybrid-A/B experiment
	TestingPatternHybridAB TestingPatternType = "Hybrid-A/B"

	// TestingPatternHybridABN indicates an experiment is a Hybrid-A/B/n experiment
	TestingPatternHybridABN TestingPatternType = "Hybrid-A/B/N"
)

// PreferredDirectionType defines the valid values for reward.PreferredDirection
type PreferredDirectionType string

const (
	// PreferredDirectionHigher indicates that a higher value is "better"
	PreferredDirectionHigher PreferredDirectionType = "High"

	// PreferredDirectionLower indicates that a lower value is "better"
	PreferredDirectionLower PreferredDirectionType = "Low"
)

// ExperimentStageType identifies valid stages of an experiment
type ExperimentStageType string

const (
	// ExperimentStageStarting indicates that start tasks (if any) are now executing
	ExperimentStageStarting ExperimentStageType = "Starting"

	// ExperimentStageLooping indicates that loop tasks (if any) are now executing
	ExperimentStageLooping ExperimentStageType = "Looping"

	// ExperimentStageFinishing indicates that finish tasks (if any) are now executing
	ExperimentStageFinishing ExperimentStageType = "Finishing"

	// ExperimentStageCompletedSuccess indicates an experiment has completed successfully
	ExperimentStageCompletedSuccess ExperimentStageType = "CompletedSuccess"

	// ExperimentStageCompletedFailure indicates an experiment has completed successfully
	ExperimentStageCompletedFailure ExperimentStageType = "CompletedFailure"
)

// MetricType identifies the type of the metric.
type MetricType string

const (
	// CounterMetricType corresponds to Prometheus Counter metric type
	CounterMetricType MetricType = "Counter"

	// GaugeMetricType corresponds to Prometheus Gauge metric type
	GaugeMetricType MetricType = "Gauge"
)

// AuthType identifies the type of authentication used in the HTTP request
type AuthType string

const (
	// BasicAuthType corresponds to authentication with basic auth
	BasicAuthType AuthType = "Basic"

	// BearerAuthType corresponds to authentication with bearer token
	BearerAuthType AuthType = "Bearer"

	// APIKeyAuthType corresponds to authentication with API keys
	APIKeyAuthType AuthType = "APIKey"
)

// MethodType identifies the HTTP request method (aka verb) used in the HTTP request
type MethodType string

const (
	// GETMethodType corresponds to HTTP GET method
	GETMethodType MethodType = "GET"

	// POSTMethodType corresponds to HTTP POST method
	POSTMethodType MethodType = "POST"
)
