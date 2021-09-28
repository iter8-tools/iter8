package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/iter8-tools/etc3/iter8ctl/utils"
	"github.com/stretchr/testify/assert"
)

type test struct {
	name           string   // name of this test
	flags          []string // flags supplied to .iter8ctl command
	outputFilename string   // relative to testdata
	// errorFilename  string   // relative to testdata
}

var tests = []test{
	// no flags to iter8ctl
	{name: "no-flags", flags: []string{}, outputFilename: "no-flags.txt"},

	// // proper inputs to iter8ctl -- all of the following need to be converted into suite tests
	// {name: "experiment1", flags: []string{"describe", "-f", utils.CompletePath("testdata", "experiment1.yaml")}, outputFilename: "experiment1.out"},
	// {name: "experiment2", flags: []string{"describe", "-f", utils.CompletePath("testdata", "experiment2.yaml")}, outputFilename: "experiment2.out"},
	// {name: "experiment3", flags: []string{"describe", "-f", utils.CompletePath("testdata", "experiment3.yaml")}, outputFilename: "experiment3.out"},
	// {name: "experiment4", flags: []string{"describe", "-f", utils.CompletePath("testdata", "experiment4.yaml")}, outputFilename: "experiment4.out"},
	// {name: "experiment5", flags: []string{"describe", "-f", utils.CompletePath("testdata", "experiment5.yaml")}, outputFilename: "experiment5.out"},
	// {name: "experiment6", flags: []string{"describe", "-f", utils.CompletePath("testdata", "experiment6.yaml")}, outputFilename: "experiment6.out"},
	// {name: "experiment7", flags: []string{"describe", "-f", utils.CompletePath("testdata", "experiment7.yaml")}, outputFilename: "experiment7.out"},
	// {name: "experiment8", flags: []string{"describe", "-f", utils.CompletePath("testdata", "experiment8.yaml")}, outputFilename: "experiment8.out"},
	// {name: "experiment9", flags: []string{"describe", "-f", utils.CompletePath("testdata", "experiment9.yaml")}, outputFilename: "experiment9.out"},
	// {name: "experiment11", flags: []string{"describe", "-f", utils.CompletePath("testdata", "experiment11.yaml")}, outputFilename: "experiment11.out"},
}

func TestMain(t *testing.T) {
	// store stdout
	rescueStdout := os.Stdout

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// when stuff is written to stdout, it can be read from rout
			rout, wout, _ := os.Pipe()
			os.Stdout = wout

			os.Args = append([]string{"./iter8ctl"}, tc.flags...)

			main()

			// read stdout and convert it into outStr
			wout.Close()
			out, _ := ioutil.ReadAll(rout)
			outStr := string(out)

			// if there is an output file specified, then compare it with outStr
			if tc.outputFilename != "" {
				of := utils.CompletePath("testdata", tc.outputFilename)
				b4, err := ioutil.ReadFile(of)
				if err != nil {
					t.Fatal("Unable to read contents of output file: ", of)
				}
				assert.Equal(t, string(b4), outStr)
			}
		})
	}

	// restore stdout
	os.Stdout = rescueStdout
}
