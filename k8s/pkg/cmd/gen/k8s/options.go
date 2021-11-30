package k8s

const (
	// Path to go template file
	k8sTemplateFilePath = "k8s.tpl"
	experimentFilePath  = "experiment.yaml"
)

type Options struct {
}

func newOptions() *Options {
	return &Options{}
}
