package core

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestService(t *testing.T) {
	// set watcherConfigEnv to test config file
	_, filename, _, _ := runtime.Caller(1) // one step up the call stack
	fname := filepath.Join(filepath.Dir(filename), "../../testdata/abninputs", "config.yaml")
	fn := filepath.Clean(fname)
	_ = os.Setenv(watcherConfigEnv, fn)

	stopCh := make(chan struct{})
	w := initializeServer()
	assert.NotNil(t, w)

	go w.Start(stopCh)
	go launchGRPCServer([]grpc.ServerOption{})

	time.Sleep(3 * time.Second)
	close(stopCh)
}
