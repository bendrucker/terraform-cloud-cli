package configwrite

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
)

type stepTest struct {
	name     string
	step     Step
	in       map[string]string
	expected map[string]string
	diags    hcl.Diagnostics
}

type stepTests []stepTest

func testStepChanges(t *testing.T, tests stepTests) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer := newTestModule(t, test.in)
			changes, diags := test.step.WithWriter(writer).Changes()

			out := make(map[string]string)
			for _, file := range changes {
				out[file.Destination()] = string(file.hcl.Bytes())
			}

			expected := make(map[string]string)
			for path, content := range test.expected {
				expected[path] = trimTestConfig(content)
			}

			assert.Equal(t, expected, out)
			assert.Equal(t, test.diags, diags)
		})
	}
}
