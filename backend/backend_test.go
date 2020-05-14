package backend

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/configs"
)

func TestIsRemote(t *testing.T) {
	type test struct {
		Name   string
		Remote bool
	}

	tests := []test{
		{
			Name:   "empty",
			Remote: false,
		},
		{
			Name:   "s3",
			Remote: false,
		},
		{
			Name:   "remote/name",
			Remote: true,
		},
		{
			Name:   "remote/prefix",
			Remote: true,
		},
		{
			Name:   "remote/hostname",
			Remote: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			parser := configs.NewParser(nil)
			module, diags := parser.LoadConfigDir(fmt.Sprintf("testdata/%s", test.Name))
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			if remote := IsRemote(module.Backend); test.Remote != remote {
				t.Fatalf("expected %t, got %t", test.Remote, remote)
			}
		})
	}
}

func TestDecodeConfig(t *testing.T) {
	type test struct {
		Name               string
		Backend            *RemoteBackend
		MultipleWorkspaces bool
	}

	tests := []test{
		{
			Name: "remote/name",
			Backend: &RemoteBackend{
				Organization: String("org"),
				Workspaces: Workspaces{
					Name: String("ws-name"),
				},
			},
		},
		{
			Name: "remote/prefix",
			Backend: &RemoteBackend{
				Organization: String("org"),
				Workspaces: Workspaces{
					Prefix: String("ws-"),
				},
			},
			MultipleWorkspaces: true,
		},
		{
			Name: "remote/hostname",
			Backend: &RemoteBackend{
				Hostname:     String("host.name"),
				Organization: String("org"),
				Workspaces: Workspaces{
					Name: String("ws-name"),
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			parser := configs.NewParser(nil)
			module, diags := parser.LoadConfigDir(fmt.Sprintf("testdata/%s", test.Name))
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			rb, err := DecodeConfig(module.Backend)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}

			if !cmp.Equal(test.Backend, rb) {
				t.Errorf("backends are not equal: %s", cmp.Diff(test.Backend, rb))
			}

			if m := rb.Workspaces.Multiple(); test.MultipleWorkspaces != m {
				t.Errorf("expected multiple workspaces = %t, got %t", test.MultipleWorkspaces, m)
			}
		})
	}
}
