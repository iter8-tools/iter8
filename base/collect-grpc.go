package base

import (
	"encoding/json"
	"errors"

	"github.com/bojand/ghz/runner"
	"github.com/go-playground/validator/v10"
	log "github.com/iter8-tools/iter8/base/log"
)

/*
Credit: The inputs for this task, including their tags and go doc comments are derived from https://github.com/bojand/ghz.
*/

// versionGRPC contains header and url information needed to send requests to each version.
type versionGRPC struct {
	// Call A fully-qualified method name in 'package.Service/method' or 'package.Service.Method' format.
	Call string `json:"call" yaml:"call" validate:"required"`
	// Host gRPC host in the hostname[:port] format.
	Host string `json:"host" yaml:"host" validate:"required"`
}

type collectGRPCInputs struct {
	// Most of the following are fields retained verbatim from runner.Config

	// call settings
	// ProtoURL	URL pointing to the proto buf.
	ProtoURL *string `json:"protoURL,omitempty" yaml:"protoURL,omitempty"`
	// EnableCompression	Enable Gzip compression on requests.
	EnableCompression bool `json:"enable-compression,omitempty" toml:"enable-compression,omitempty" yaml:"enable-compression,omitempty"`

	// load settings
	// RPS	Requests per second (RPS) rate limit for constant load schedule. Default is no rate limit.
	RPS uint `json:"rps" toml:"rps" yaml:"rps"`
	// LoadSchedule	Specifies the load schedule. Options are const, step, or line. Default is const.
	LoadSchedule string `json:"load-schedule" toml:"load-schedule" yaml:"load-schedule" default:"const"`
	// LoadStart	Specifies the RPS load start value for step or line schedules.
	LoadStart uint `json:"load-start" toml:"load-start" yaml:"load-start"`
	// LoadEnd	Specifies the load end value for step or line load schedules.
	LoadEnd uint `json:"load-end" toml:"load-end" yaml:"load-end"`
	// LoadStep	Specifies the load step value or slope value.
	LoadStep int `json:"load-step" toml:"load-step" yaml:"load-step"`
	// LoadStepDuration	Specifies the load step duration value for step load schedule.
	LoadStepDuration runner.Duration `json:"load-step-duration" toml:"load-step-duration" yaml:"load-step-duration"`
	// LoadMaxDuration	Specifies the max load duration value for step or line load schedule.
	LoadMaxDuration runner.Duration `json:"load-max-duration" toml:"load-max-duration" yaml:"load-max-duration"`

	// concurrency settings
	// C	Number of request workers to run concurrently for const concurrency schedule. Default is 50.
	C uint `json:"concurrency" toml:"concurrency" yaml:"concurrency" default:"50"`
	// CSchedule	Concurrency change schedule. Options are const, step, or line. Default is const.
	CSchedule string `json:"concurrency-schedule" toml:"concurrency-schedule" yaml:"concurrency-schedule" default:"const"`
	// CStart	Concurrency start value for step and line concurrency schedules.
	CStart uint `json:"concurrency-start" toml:"concurrency-start" yaml:"concurrency-start" default:"1"`
	// CEnd	Concurrency end value for step and line concurrency schedules.
	CEnd uint `json:"concurrency-end" toml:"concurrency-end" yaml:"concurrency-end" default:"0"`
	// CStep	Concurrency step / slope value for step and line concurrency schedules.
	CStep int `json:"concurrency-step" toml:"concurrency-step" yaml:"concurrency-step" default:"0"`
	// CStepDuration Specifies the concurrency step duration value for step concurrency schedule.
	CStepDuration runner.Duration `json:"concurrency-step-duration" toml:"concurrency-step-duration" yaml:"concurrency-step-duration" default:"0"`
	// CMaxDuration	Specifies the max concurrency adjustment duration value for step or line concurrency schedule.
	CMaxDuration runner.Duration `json:"concurrency-max-duration" toml:"concurrency-max-duration" yaml:"concurrency-max-duration" default:"0"`

	// test settings
	// N	Number of requests to run. Default is 200.
	N uint `json:"total" toml:"total" yaml:"total" default:"200"`
	// Async	Make requests asynchronous as soon as possible. Does not wait for request to finish before sending next one.
	Async bool `json:"async,omitempty" toml:"async,omitempty" yaml:"async,omitempty"`

	// connection settings
	// Connections	Number of connections to use. Concurrency is distributed evenly among all the connections. Default is 1.
	Connections uint `json:"connections" toml:"connections" yaml:"connections" default:"1"`

	// timeout settings
	// Z	Duration of application to send requests. When duration is reached, application stops and exits. If duration is specified, n is ignored. Examples: -z 10s -z 3m.
	Z runner.Duration `json:"duration" toml:"duration" yaml:"duration"`
	//	Maximum duration of application to send requests with n setting respected. If duration is reached before n requests are completed, application stops and exits. Examples: -x 10s -x 3m.
	X runner.Duration `json:"max-duration" toml:"max-duration" yaml:"max-duration"`
	//	Timeout	Timeout for each request. Default is 20s, use 0 for infinite.
	Timeout runner.Duration `json:"timeout" toml:"timeout" yaml:"timeout" default:"20s"`
	//	Connection timeout for the initial connection dial. Default is 10s.
	DialTimeout runner.Duration `json:"connect-timeout" toml:"connect-timeout" yaml:"connect-timeout" default:"10s"`
	//	Keepalive time duration. Only used if present and above 0.
	KeepaliveTime runner.Duration `json:"keepalive" toml:"keepalive" yaml:"keepalive"`

	// Stream settings
	//	SI	Interval for stream requests between message sends.
	SI runner.Duration `json:"stream-interval" toml:"stream-interval" yaml:"stream-interval"`
	// StreamCallDuration	Duration after which client will close the stream in each streaming call.
	StreamCallDuration runner.Duration `json:"stream-call-duration" toml:"stream-call-duration" yaml:"stream-call-duration"`
	// StreamCallCount	Count of messages sent, after which client will close the stream in each streaming call.
	StreamCallCount uint `json:"stream-call-count" toml:"stream-call-count" yaml:"stream-call-count"`
	//	In streaming calls, regenerate and apply call template data on every message send.
	StreamDynamicMessages bool `json:"stream-dynamic-messages" toml:"stream-dynamic-messages" yaml:"stream-dynamic-messages"`

	// data and metadata settings
	// JSONDataURL	URL pointing to JSON data to be sent as part of the call.
	JSONDataURL *string `json:"JSONDataURL,omitempty" yaml:"JSONDataURL,omitempty"`
	// JSONDataStr	Stringified JSON call data.
	JSONDataStr *string `json:"JSONDataStr,omitempty" yaml:"JSONDataStr,omitempty"`
	// BinaryDataURL	URL pointing to call data as serialized binary message or multiple count-prefixed messages.
	BinaryDataURL *string `json:"BinaryDataURL,omitempty" yaml:"BinaryDataURL,omitempty"`
	// MetadataURL	URL pointing to call metadata in JSON format.
	MetadataURL *string `json:"MetadataURL,omitempty" yaml:"MetadataURL,omitempty"`
	// Metadata	Call metadata
	Metadata map[string]string `json:"metadata,omitempty" toml:"metadata,omitempty" yaml:"metadata,omitempty"`
	// ReflectMetaDataURL URL pointing to reflect metadata in JSON format.
	ReflectMetadataURL *string `json:"ReflectMetadataURL,omitempty" yaml:"ReflectMetadataURL,omitempty"`
	// ReflectMetadata	Reflect meta data
	ReflectMetadata map[string]string `json:"reflect-metadata,omitempty" toml:"reflect-metadata,omitempty" yaml:"reflect-metadata,omitempty"`

	// misc settings
	//	CountErrors	Count erroneous (non-OK) resoponses in stats calculations.
	CountErrors bool `json:"count-errors" toml:"count-errors" yaml:"count-errors"`
	// SkipFirst	Skip the first X requests when doing the results tally.
	SkipFirst uint `json:"skipFirst" toml:"skipFirst" yaml:"skipFirst"`
	//	CPUs	Number of cpu cores to use.
	CPUs uint `json:"cpus" toml:"cpus" yaml:"cpus"`

	// VersionInfo is a non-empty list of versionGRPC values.
	VersionInfo []*versionGRPC `json:"versionInfo" yaml:"versionInfo" validate:"required,notallnil"`
}

const (
	// CollectGPRCTaskName is the name of this task which performs load generation and metrics collection for gRPC services.
	CollectGPRCTaskName = "gen-load-and-collect-metrics-grpc"
)

// collectGRPCTask enables load testing of gRPC services.
type collectGRPCTask struct {
	taskMeta
	With collectGRPCInputs `json:"with" yaml:"with" validate:"required"`
}

// MakeCollectGRPC constructs a CollectGRPCTask out of a collect grpc task spec
func MakeCollectGRPC(t *TaskSpec) (Task, error) {
	if *t.Task != CollectGPRCTaskName {
		return nil, errors.New("task need to be " + CollectGPRCTaskName)
	}
	var err error
	var jsonBytes []byte
	var bt Task
	// convert t to jsonBytes
	jsonBytes, err = json.Marshal(t)
	// convert jsonString to CollectTask
	if err == nil {
		ct := &collectGRPCTask{}
		err = json.Unmarshal(jsonBytes, &ct)
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("invalid collect grpc task specification")
			return nil, err
		}

		// validate enablese validation of grpc task specifications
		validate := validator.New()
		// returns nil or ValidationErrors ( []FieldError )
		validate.RegisterValidation("notallnil", notAllNil)

		err = validate.Struct(ct)
		if err != nil {
			log.Logger.WithStackTrace(err.Error()).Error("invalid collect grpc task specification")
			return nil, err
		}

		bt = ct
	}
	return bt, err
}

// initializeDefaults sets default values for the collect task
func (t *collectGRPCTask) initializeDefaults() {
}

// GetName returns the name of the collect grpc task
func (t *collectGRPCTask) GetName() string {
	return CollectGPRCTaskName
}

// Run executes this task
func (t *collectGRPCTask) Run(exp *Experiment) error {
	t.initializeDefaults()
	return errors.New("not implemented")
}
