package cmd

import (
	"time"

	ia "github.com/iter8-tools/iter8/action"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/spf13/cobra"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli/values"
)

const kLaunchDesc = `
This command installs a chart archive.

The install argument must be a chart reference, a path to a packaged chart,
a path to an unpacked chart directory or a URL.

To override values in a chart, use either the '--values' flag and pass in a file
or use the '--set' flag and pass configuration from the command line, to force
a string value use '--set-string'. You can use '--set-file' to set individual
values from a file when the value itself is too long for the command line
or is dynamically generated.

    $ helm install -f myvalues.yaml myredis ./redis

or

    $ helm install --set name=prod myredis ./redis

or

    $ helm install --set-string long_int=1234567890 myredis ./redis

or

    $ helm install --set-file my_script=dothings.sh myredis ./redis

You can specify the '--values'/'-f' flag multiple times. The priority will be given to the
last (right-most) file specified. For example, if both myvalues.yaml and override.yaml
contained a key called 'Test', the value set in override.yaml would take precedence:

    $ helm install -f myvalues.yaml -f override.yaml  myredis ./redis

You can specify the '--set' flag multiple times. The priority will be given to the
last (right-most) set specified. For example, if both 'bar' and 'newbar' values are
set for a key called 'foo', the 'newbar' value would take precedence:

    $ helm install --set foo=bar --set foo=newbar  myredis ./redis


To check the generated manifests of a release without installing the chart,
the '--debug' and '--dry-run' flags can be combined.

If --verify is set, the chart MUST have a provenance file, and the provenance
file MUST pass all verification steps.

There are five different ways you can express the chart you want to install:

1. By chart reference: helm install mymaria example/mariadb
2. By path to a packaged chart: helm install mynginx ./nginx-1.2.3.tgz
3. By path to an unpacked chart directory: helm install mynginx ./nginx
4. By absolute URL: helm install mynginx https://example.com/charts/nginx-1.2.3.tgz
5. By chart reference and repo url: helm install --repo https://example.com/charts/ mynginx nginx

CHART REFERENCES

A chart reference is a convenient way of referencing a chart in a chart repository.

When you use a chart reference with a repo prefix ('example/mariadb'), Helm will look in the local
configuration for a chart repository named 'example', and will then look for a
chart in that repository whose name is 'mariadb'. It will install the latest stable version of that chart
until you specify '--devel' flag to also include development version (alpha, beta, and release candidate releases), or
supply a version number with the '--version' flag.

To see the list of chart repositories, use 'helm repo list'. To search for
charts in a repository, use 'helm search'.
`

func newKLaunchCmd(cfg *action.Configuration) *cobra.Command {
	actor := ia.NewLaunch(cfg)
	valueOpts := &values.Options{}

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "launch an experiment in Kubernetes",
		Long:  kLaunchDesc,
		RunE: func(_ *cobra.Command, _ []string) error {
			err := actor.RunKubernetes(valueOpts)
			if err != nil {
				log.Logger.Error(err)
				return err
			}
			return nil
		},
	}
	addKLaunchFlags(cmd, actor, valueOpts)
	return cmd
}

func addKLaunchFlags(cmd *cobra.Command, actor *ia.Launch, valueOpts *values.Options) {
	cmd.Flags().BoolVar(&actor.CreateNamespace, "create-namespace", false, "create the experiment namespace if not present")
	cmd.Flags().DurationVar(&actor.Timeout, "timeout", 300*time.Second, "time to wait for the experiment to complete")
	cmd.Flags().BoolVar(&actor.WaitForJobs, "wait-for-experiment", false, "if set and --wait enabled, will wait until the experiment has completed before marking the launch as successful. It will wait for as long as --timeout")
	cmd.Flags().BoolVar(&actor.Atomic, "atomic", false, "if set, the launch process deletes all experiment related resources on failure. The --wait flag will be set automatically if --atomic is used")
	addExperimentGroupFlag(cmd, &actor.ExperimentGroup)
	addLaunchFlags(cmd, actor, valueOpts)
}