package configwrite

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// Step is a step required to prepare a module to run in Terraform Cloud
type Step interface {
	Name() string

	// Description returns a description of the step
	Description() string

	// Changes returns a list of files changes and diagnostics if errors ocurred. If Complete() returns true, this should be empty.
	Changes() (Changes, hcl.Diagnostics)

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

func (s Steps) Changes() (Changes, hcl.Diagnostics) {
	result := make(Changes)
	var diags hcl.Diagnostics

	for _, step := range s {
		changes, diags := step.Changes()

		for path, change := range changes {
			if err, ok := result.Add(path, change).(*renameCollisionError); ok {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagWarning,
					Summary:  "Rename skipped due to conflict",
					Detail:   fmt.Sprintf(`The "%s" step attempted to rename %s to %s, but a previous step already renamed this file to %s.`, step.Name(), path, err.Proposed, err.Existing),
					Subject:  &hcl.Range{Filename: err.Proposed},
				})
			}
		}

		if diags.HasErrors() {
			return result, diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf(`Step "%s" returned error(s)`, step.Name()),
				Detail:   fmt.Sprintf(`The "%s" step returned %d error(s). It changed %d files. Check the results for accuracy.`, step.Name(), len(errorDiags(diags)), len(changes)),
			})
		}

	}

	return result, diags
}
