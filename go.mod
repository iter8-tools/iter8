module github.com/iter8-tools/iter8

go 1.16

require (
	fortio.org/fortio v1.19.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/antonmedv/expr v1.9.0
	github.com/go-playground/validator/v10 v10.9.0
	github.com/hashicorp/go-getter v1.5.9
	github.com/jarcoal/httpmock v1.0.8
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	helm.sh/helm/v3 v3.7.2
	k8s.io/api v0.22.4
	k8s.io/apimachinery v0.22.4
	k8s.io/cli-runtime v0.22.4
	k8s.io/client-go v0.22.4
	sigs.k8s.io/yaml v1.2.0
)

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
