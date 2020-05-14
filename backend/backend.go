package backend

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/terraform/configs"
)

// DefaultHostname is the hostname for the public version of Terraform Cloud
const DefaultHostname = "app.terraform.io"

// IsRemote returns whether the backend is defined and is of type "remote"
func IsRemote(backend *configs.Backend) bool {
	return backend != nil && backend.Type == "remote"
}

// DecodeConfig decodes a backend config into a RemoteBackend object
func DecodeConfig(backend *configs.Backend) (*RemoteBackend, error) {
	rb := &RemoteBackend{}
	if diags := gohcl.DecodeBody(backend.Config, &hcl.EvalContext{}, rb); diags.HasErrors() {
		return nil, diags
	}
	return rb, nil
}

// RemoteBackend is a Terraform remote backend configuration
type RemoteBackend struct {
	Hostname     *string `hcl:"hostname,attr"`
	Organization *string `hcl:"organization,attr"`

	Workspaces Workspaces `hcl:"workspaces,block"`
}

// Workspaces is a Terraform remote backend workspaces configuration
type Workspaces struct {
	Name   *string `hcl:"name,attr"`
	Prefix *string `hcl:"prefix,attr"`
}

// Multiple returns whether multiple workspaces are defined with a prefix
func (w *Workspaces) Multiple() bool {
	return w.Prefix != nil
}

// String returns a string pointer for a string value
func String(v string) *string {
	return &v
}
