package base

var (
	settings = NewEnvSettings()
	kd       = NewKubeDriver(settings)
)
