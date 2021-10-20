package core

// Run an experiment
func (e *Experiment) Run() error {
	var err error
	for _, t := range e.Tasks {
		err = t.Run(e)
	}
	return err
}
