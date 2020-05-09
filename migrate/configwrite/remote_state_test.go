package configwrite

import (
	"testing"
)

func TestRemoteState(t *testing.T) {
	testStepChanges(t, stepTests{
		{
			name: "incomplete",
			step: &RemoteState{
				RemoteBackend: RemoteBackendConfig{
					Hostname:     "host.name",
					Organization: "org",
					Workspaces: WorkspaceConfig{
						Name: "ws",
					},
				},
				Path: "dependent/",
			},
			in: map[string]string{
				"backend.tf": `
					terraform {
						backend "s3" {
							key    = "terraform.tfstate"
							bucket = "terraform-state"
							region = "us-east-1"
						}
					}
				`,
				"./dependent/a/backend.tf": `
					data "terraform_remote_state" "match" {
						backend = "s3"
					
						config = {
							key    = "terraform.tfstate"
							bucket = "terraform-state"
							region = "us-east-1"
						}
					}
					
					data "terraform_remote_state" "wrong_type" {
						backend = "remote"
					
						config = {}
					}
					
					data "terraform_remote_state" "wrong_config" {
						backend = "s3"
					
						config = {
							key    = "a-different-terraform.tfstate"
							bucket = "terraform-state"
							region = "us-east-1"
						}
					}
				`,
			},
			expected: map[string]string{
				"dependent/a/backend.tf": `
					data "terraform_remote_state" "match" {
						backend = "remote"
					
						config = {
							hostname     = "host.name"
							organization = "org"
					
							workspaces = {
								name = "ws"
							}
						}
					}
					
					data "terraform_remote_state" "wrong_type" {
						backend = "remote"
					
						config = {}
					}
					
					data "terraform_remote_state" "wrong_config" {
						backend = "s3"
					
						config = {
							key    = "a-different-terraform.tfstate"
							bucket = "terraform-state"
							region = "us-east-1"
						}
					}
				`,
			},
		},
		{
			name: "incomplete/prefix",
			step: &RemoteState{
				RemoteBackend: RemoteBackendConfig{
					Hostname:     "host.name",
					Organization: "org",
					Workspaces: WorkspaceConfig{
						Prefix: "ws-",
					},
				},
				Path: "dependent/",
			},
			in: map[string]string{
				"backend.tf": `
					terraform {
						backend "s3" {
							key    = "terraform.tfstate"
							bucket = "terraform-state"
							region = "us-east-1"
						}
					}
				`,
				"./dependent/a/backend.tf": `
					data "terraform_remote_state" "match" {
						backend   = "s3"
						workspace = terraform.workspace
					
						config = {
							key    = "terraform.tfstate"
							bucket = "terraform-state"
							region = "us-east-1"
						}
					}
					
					data "terraform_remote_state" "wrong_type" {
						backend = "remote"
					
						config = {}
					}
					
					data "terraform_remote_state" "wrong_config" {
						backend = "s3"
					
						config = {
							key    = "a-different-terraform.tfstate"
							bucket = "terraform-state"
							region = "us-east-1"
						}
					}
				`,
			},
			expected: map[string]string{
				"dependent/a/backend.tf": `
					data "terraform_remote_state" "match" {
						backend = "remote"
					
						config = {
							hostname     = "host.name"
							organization = "org"
					
							workspaces = {
								name = "ws-${terraform.workspace}"
							}
						}
					}
					
					data "terraform_remote_state" "wrong_type" {
						backend = "remote"
					
						config = {}
					}
					
					data "terraform_remote_state" "wrong_config" {
						backend = "s3"
					
						config = {
							key    = "a-different-terraform.tfstate"
							bucket = "terraform-state"
							region = "us-east-1"
						}
					}
				`,
			},
		},
	})
}
