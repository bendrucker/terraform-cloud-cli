package configwrite

import (
	"testing"
)

func TestTerraformWorkspace(t *testing.T) {
	testStepChanges(t, stepTests{
		{
			name: "incomplete",
			step: &TerraformWorkspace{Variable: "environment"},
			in: map[string]string{
				"outputs.tf": `
					output "attribute" {
						value = terraform.workspace
					}
					
					output "interpolated" {
						value = "The workspace is ${terraform.workspace}"
					}
					
					output "function" {
						value = lookup({}, terraform.workspace, false)
					}				
				`,
				"variables.tf": `
					variable "foo" {}

					variable "bar" {}
					
					variable "baz" {}				
				`,
			},
			expected: map[string]string{
				"outputs.tf": `
					output "attribute" {
						value = var.environment
					}
					
					output "interpolated" {
						value = "The workspace is ${var.environment}"
					}
					
					output "function" {
						value = lookup({}, var.environment, false)
					}
				`,
				"variables.tf": `
					variable "environment" {
						type        = string
						description = "The environment where the module will be deployed"
					}
					
					variable "foo" {}
					
					variable "bar" {}
					
					variable "baz" {}
				`,
			},
		},
		{
			name: "incomplete",
			step: &TerraformWorkspace{Variable: "environment"},
			in: map[string]string{
				"outputs.tf": `
					output "attribute" {
						value = var.environment
					}
					
					output "interpolated" {
						value = "The workspace is ${var.environment}"
					}
					
					output "function" {
						value = lookup({}, var.environment, false)
					}				
				`,
			},
			expected: map[string]string{},
		},
	})
}
