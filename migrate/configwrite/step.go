package configwrite

import (
	"fmt"
)

// Step is a step required to prepare a module to run in Terraform Cloud
type Step interface {
	Name() string

	// Description returns a description of the step
	Description() string

	// Changes returns a list of file changes and an error if changes could not be completed
	Changes() (Changes, error)

	WithWriter(*Writer) Step
}

func NewSteps(w *Writer, steps Steps) Steps {
	for _, step := range steps {
		step.WithWriter(w)
	}
	return steps
}

type Steps []Step

func (s Steps) Append(steps ...Step) Steps {
	return append(s, steps...)
}

func (s Steps) Changes() (Changes, error) {
	result := make(Changes)

	for _, step := range s {
		changes, err := step.Changes()
		if err != nil {
			return Changes{}, err
		}

		for path, file := range changes {
			if existing, ok := result[path]; ok {
				if existing.NewName != "" && file.NewName != "" {
					return Changes{}, fmt.Errorf(`conflict: step '%s' attempted to rename '%s' to '%s', but a previous step already renamed this file to '%s'`, step.Name(), path, file.NewName, existing.NewName)
				}
			}

			result[path] = file
		}
	}

	return result, nil
}
