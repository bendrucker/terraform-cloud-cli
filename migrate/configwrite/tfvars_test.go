package configwrite

import (
	"testing"
)

func TestTfvars(t *testing.T) {
	testStepChanges(t, stepTests{
		{
			name: "incomplete",
			step: &Tfvars{Filename: "terraform.auto.tfvars"},
			in: map[string]string{
				"main.tf": "",
				"terraform.tfvars": `
					foo = "bar"
					baz = "qux"
				`,
			},
			expected: map[string]string{
				"terraform.auto.tfvars": `
					foo = "bar"
					baz = "qux"
				`,
			},
		},
		{
			name: "incomplete/custom_name",
			step: &Tfvars{Filename: "default.auto.tfvars"},
			in: map[string]string{
				"main.tf": "",
				"terraform.tfvars": `
					foo = "bar"
					baz = "qux"
				`,
			},
			expected: map[string]string{
				"default.auto.tfvars": `
					foo = "bar"
					baz = "qux"
				`,
			},
		},
		{
			name: "complete",
			step: &Tfvars{Filename: "terraform.auto.tfvars"},
			in: map[string]string{
				"main.tf": "",
				"terraform.auto.tfvars": `
					foo = "bar"
					baz = "qux"
				`,
			},
			expected: map[string]string{},
		},
	})
}
