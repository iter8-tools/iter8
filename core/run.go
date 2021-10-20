package core

// Run an experiment
func (e *Experiment) Run() error {
	var err error
	for i, t := range e.Tasks {
		err = t.Run(i)
	}
	return err
}
