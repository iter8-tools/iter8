/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Credit: this file is composed by modifying the following two files from Helm in minor ways
// https://raw.githubusercontent.com/helm/helm/974a6030c8514591ab0b0f0c898d37f816f698f6/cmd/helm/helm_test.go
// https://github.com/helm/helm/blob/974a6030c8514591ab0b0f0c898d37f816f698f6/internal/test/test.go#L53

package cmd

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	id "github.com/iter8-tools/iter8/driver"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
)

func testTimestamper() time.Time { return time.Unix(242085845, 0).UTC() }

type testFormatter struct {
	logrus.Formatter
}

func (u testFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = testTimestamper()
	return u.Formatter.Format(e)
}

func runTestActionCmd(t *testing.T, tests []cmdTestCase) {
	// fixed time
	log.Logger.SetFormatter(testFormatter{log.Logger.Formatter})
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer resetEnv()()

			store := storageFixture()
			_, out, err := executeActionCommandC(store, tt.cmd)
			if (err != nil) != tt.wantError {
				t.Errorf("expected error, got '%v'", err)
			}
			if tt.golden != "" {
				AssertGoldenString(t, out, tt.golden)
			}
		})
	}
}

func storageFixture() *storage.Storage {
	return storage.Init(driver.NewMemory())
}

func executeActionCommandC(store *storage.Storage, cmd string) (*cobra.Command, string, error) {
	return executeActionCommandStdinC(store, nil, cmd)
}

func executeActionCommandStdinC(store *storage.Storage, in *os.File, cmd string) (*cobra.Command, string, error) {
	args, err := shellwords.Parse(cmd)
	if err != nil {
		return nil, "", err
	}

	buf := new(bytes.Buffer)

	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	log.Logger.Out = buf
	outStream = buf

	oldStdin := os.Stdin
	if in != nil {
		rootCmd.SetIn(in)
		os.Stdin = in
	}

	if mem, ok := store.Driver.(*driver.Memory); ok {
		mem.SetNamespace(settings.Namespace())
	}

	c, err := rootCmd.ExecuteC()

	result := buf.String()

	os.Stdin = oldStdin

	return c, result, err
}

// cmdTestCase describes a test case that works with releases.
type cmdTestCase struct {
	name      string
	cmd       string
	golden    string
	wantError bool
}

// func executeActionCommand(cmd string) (*cobra.Command, string, error) {
// 	return executeActionCommandC(storageFixture(), cmd)
// }

func resetEnv() func() {
	origEnv := os.Environ()
	return func() {
		os.Clearenv()
		for _, pair := range origEnv {
			kv := strings.SplitN(pair, "=", 2)
			os.Setenv(kv[0], kv[1])
		}
		logLevel = "info"
		*settings = *cli.New()
		*kd = *id.NewFakeKubeDriver(settings)
		log.Logger.Out = os.Stderr
	}
}

// the following test utils are from
// https://github.com/helm/helm/blob/974a6030c8514591ab0b0f0c898d37f816f698f6/internal/test/test.go#L53

// UpdateGolden writes out the golden files with the latest values, rather than failing the test.
var updateGolden = flag.Bool("update", false, "update golden files")

// TestingT describes a testing object compatible with the critical functions from the testing.T type
type TestingT interface {
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	HelperT
}

// HelperT describes a test with a helper function
type HelperT interface {
	Helper()
}

// AssertGoldenBytes asserts that the give actual content matches the contents of the given filename
func AssertGoldenBytes(t TestingT, actual []byte, filename string) {
	t.Helper()

	if err := compare(actual, aPath(filename)); err != nil {
		t.Fatalf("%v", err)
	}
}

// AssertGoldenString asserts that the given string matches the contents of the given file.
func AssertGoldenString(t TestingT, actual, filename string) {
	t.Helper()

	if err := compare([]byte(actual), aPath(filename)); err != nil {
		t.Fatalf("%v", err)
	}
}

// AssertGoldenFile asserts that the content of the actual file matches the contents of the expected file
func AssertGoldenFile(t TestingT, actualFileName string, expectedFilename string) {
	t.Helper()

	actual, err := ioutil.ReadFile(actualFileName)
	if err != nil {
		t.Fatalf("%v", err)
	}
	AssertGoldenBytes(t, actual, expectedFilename)
}

func aPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return base.CompletePath("../", "testdata/"+filename)
}

func compare(actual []byte, filename string) error {
	actual = normalize(actual)
	if err := update(filename, actual); err != nil {
		return err
	}

	expected, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "unable to read testdata %s", filename)
	}
	expected = normalize(expected)
	if !bytes.Equal(expected, actual) {
		return errors.Errorf("does not match golden file %s\n\nWANT:\n'%s'\n\nGOT:\n'%s'\n", filename, expected, actual)
	}
	return nil
}

func update(filename string, in []byte) error {
	if !*updateGolden {
		return nil
	}
	return ioutil.WriteFile(filename, normalize(in), 0666)
}

func normalize(in []byte) []byte {
	return bytes.Replace(in, []byte("\r\n"), []byte("\n"), -1)
}
