package configwrite

import (
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const (
	BackendTypeRemote = "remote"
)

type RemoteBackend struct {
	writer *Writer
	Config RemoteBackendConfig
}

type RemoteBackendConfig struct {
	Hostname     string
	Organization string
	Workspaces   WorkspaceConfig
}

type WorkspaceConfig struct {
	Name   string
	Prefix string
}

func (b *RemoteBackend) WithWriter(w *Writer) Step {
	b.writer = w
	return b
}

func (b *RemoteBackend) Name() string {
	return "Remote Backend"
}

// Description returns a description of the step
func (b *RemoteBackend) Description() string {
	return `A "remote" backend should be configured for Terraform Cloud (https://www.terraform.io/docs/backends/types/remote.html)`
}

// MultipleWorkspaces returns whether the remote backend will be configured for multiple prefixed workspaces
func (b *RemoteBackend) MultipleWorkspaces() bool {
	return b.Config.Workspaces.Prefix != ""
}

// Changes updates the configured backend
func (b *RemoteBackend) Changes() (Changes, hcl.Diagnostics) {
	if b.writer.module.Backend.Type == "remote" {
		return Changes{}, nil
	}

	var path string
	var file *File
	var diags hcl.Diagnostics

	if b.writer.HasBackend() {
		path = b.writer.Backend().DeclRange.Filename
		file, diags = b.writer.File(path)
	} else {
		path = filepath.Join(b.writer.Dir(), "backend.tf")
		file, diags = b.writer.File(path)
		tf := file.hcl.Body().AppendBlock(hclwrite.NewBlock("terraform", []string{}))
		tf.Body().AppendBlock(hclwrite.NewBlock("backend", []string{"remote"}))
	}

	for _, block := range file.hcl.Body().Blocks() {
		if block.Type() != "terraform" {
			continue
		}

		for _, child := range block.Body().Blocks() {
			if child.Type() != "backend" {
				continue
			}

			block.Body().RemoveBlock(child)

			remote := block.Body().AppendBlock(hclwrite.NewBlock("backend", []string{"remote"})).Body()
			remote.SetAttributeValue("hostname", cty.StringVal(b.Config.Hostname))
			remote.SetAttributeValue("organization", cty.StringVal(b.Config.Organization))
			remote.AppendNewline()

			workspaces := remote.AppendBlock(hclwrite.NewBlock("workspaces", nil)).Body()
			if b.MultipleWorkspaces() {
				workspaces.SetAttributeValue("prefix", cty.StringVal(b.Config.Workspaces.Prefix))
			} else {
				workspaces.SetAttributeValue("name", cty.StringVal(b.Config.Workspaces.Name))
			}
		}

	}

	return Changes{path: file}, diags
}

var _ Step = (*RemoteBackend)(nil)
