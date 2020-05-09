package configwrite

import (
	"testing"
)

func TestRemoteBackend(t *testing.T) {
	testStepChanges(t, stepTests{
		{
			name: "incomplete",
			step: &RemoteBackend{
				Config: RemoteBackendConfig{
					Hostname:     "host.name",
					Organization: "org",
					Workspaces: WorkspaceConfig{
						Name: "ws",
					},
				},
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
			},
			expected: map[string]string{
				"backend.tf": `
					terraform {
						backend "remote" {
							hostname     = "host.name"
							organization = "org"
					
							workspaces {
								name = "ws"
							}
						}
					}
				`,
			},
		},
		{
			name: "incomplete/prefix",
			step: &RemoteBackend{
				Config: RemoteBackendConfig{
					Hostname:     "host.name",
					Organization: "org",
					Workspaces: WorkspaceConfig{
						Prefix: "ws-",
					},
				},
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
			},
			expected: map[string]string{
				"backend.tf": `
					terraform {
						backend "remote" {
							hostname     = "host.name"
							organization = "org"
					
							workspaces {
								prefix = "ws-"
							}
						}
					}
				`,
			},
		},
		{
			name: "complete",
			step: &RemoteBackend{
				Config: RemoteBackendConfig{
					Hostname:     "host.name",
					Organization: "org",
					Workspaces: WorkspaceConfig{
						Prefix: "ws-",
					},
				},
			},
			in: map[string]string{
				"backend.tf": `
					terraform {
						backend "remote" {
							hostname     = "host.name"
							organization = "org"
					
							workspaces {
								prefix = "ws-"
							}
						}
					}
				`,
			},
			expected: map[string]string{},
		},
	})
}
