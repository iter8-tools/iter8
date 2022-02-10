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
	fortio.org/fortio v1.20.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/antonmedv/expr v1.9.0
	github.com/bojand/ghz v0.106.1
	github.com/jarcoal/httpmock v1.1.0
	github.com/jinzhu/copier v0.3.5
	github.com/mcuadros/go-defaults v1.2.0
	github.com/montanaflynn/stats v0.6.6
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
	helm.sh/helm/v3 v3.8.0
	k8s.io/cli-runtime v0.23.3 // indirect
	k8s.io/client-go v0.23.3
	sigs.k8s.io/yaml v1.3.0
)
