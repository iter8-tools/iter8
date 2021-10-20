package core

// Experiment context provides the methods needed to store and retrieve experiment spec and results
type ExperimentContext interface {
	// ReadSpec reads experiment spec
	ReadSpec() (*ExperimentSpec, error)
	// ReadResult reads experiment result
	ReadResult() (*ExperimentResult, error)
	// WriteResult writes experiment result
	WriteResult() error
}

// File context enables file based experiments
type FileContext struct {
	// SpecFile is the full path to the experiment
	SpecFile string
	// ResultFile is the full path to the experiment result
	ResultFile string
}

// ReadSpec from file
func (fc *FileContext) ReadSpec() (*ExperimentSpec, error) {
	return nil, nil
}

// ReadResult from file
func (fc *FileContext) ReadResult() (*ExperimentResult, error) {
	return nil, nil
}

// WriteResult to file
func (fc *FileContext) WriteResult() error {
	return nil
}
