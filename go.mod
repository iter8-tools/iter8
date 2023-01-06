module github.com/iter8-tools/iter8

go 1.16

retract (
	// Published v1 too early
	[v1.0.0, v1.0.2]
	// Named iter8-istio controller as iter8 too early
	v1.0.0-rc3
	// Named iter8-istio controller as iter8 too early
	v1.0.0-rc2
	// Named iter8-istio controller as iter8 too early
	v1.0.0-rc1
	// Named iter8-istio controller as iter8 too early
	v1.0.0-preview
	// Named iter8-istio controller as iter8 too early
	[v0.0.1, v0.7.30]
)

require (
	fortio.org/fortio v1.27.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/antonmedv/expr v1.9.0
	github.com/bojand/ghz v0.111.0
	github.com/imdario/mergo v0.3.12
	github.com/itchyny/gojq v0.12.7
	github.com/jarcoal/httpmock v1.1.0
	github.com/mattn/go-shellwords v1.0.12
	github.com/mcuadros/go-defaults v1.2.0
	github.com/montanaflynn/stats v0.6.6
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.5.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.0
	golang.org/x/net v0.0.0-20220722155237-a158d28d115b
	golang.org/x/text v0.3.7
	google.golang.org/grpc v1.47.0
	google.golang.org/protobuf v1.28.0
	helm.sh/helm/v3 v3.10.3
	k8s.io/api v0.25.2
	k8s.io/apimachinery v0.25.2
	k8s.io/client-go v0.25.2
	sigs.k8s.io/yaml v1.3.0
)
