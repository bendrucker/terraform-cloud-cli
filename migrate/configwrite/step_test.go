package configwrite

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type stepTest struct {
	name     string
	step     Step
	in       map[string]string
	expected map[string]string
}

type stepTests []stepTest

func testStepChanges(t *testing.T, tests stepTests) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writer := newTestModule(t, test.in)
			changes, err := test.step.WithWriter(writer).Changes()
			if !assert.NoError(t, err) {
				return
			}

			out := make(map[string]string)
			for _, file := range changes {
				out[file.Destination()] = string(file.hcl.Bytes())
			}

			expected := make(map[string]string)
			for path, content := range test.expected {
				expected[path] = trimTestConfig(content)
			}

			assert.Equal(t, expected, out)
		})
	}
}
