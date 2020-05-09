package configwrite

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform/configs"
	"github.com/spf13/afero"
)

func New(path string) (*Writer, hcl.Diagnostics) {
	return newWriter(path, nil)
}

func newWriter(path string, fs afero.Fs) (*Writer, hcl.Diagnostics) {
	if fs == nil {
		fs = afero.NewOsFs()
	}
	parser := configs.NewParser(fs)

	if !parser.IsConfigDir(path) {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Not a module directory",
				Detail:   fmt.Sprintf("Directory %s does not contain Terraform configuration files.", path),
				Subject: &hcl.Range{
					Filename: path,
				},
			},
		}
	}

	module, diags := parser.LoadConfigDir(path)

	return &Writer{
		fs:     fs,
		parser: parser,
		module: module,
		files:  make(map[string]*hclwrite.File),
	}, diags
}

// Writer provides access to information about the Terraform module structure and the ability to update its files
type Writer struct {
	fs     afero.Fs
	parser *configs.Parser
	module *configs.Module
	files  map[string]*hclwrite.File
}

// Dir returns the module directory
func (w *Writer) Dir() string {
	return w.module.SourceDir
}

// Backend returns the backend, or nil if none is defined
func (w *Writer) Backend() *configs.Backend {
	return w.module.Backend
}

// HasBackend returns true if the module has a backend configuration
func (w *Writer) HasBackend() bool {
	return w.Backend() != nil
}

// Variables returns the declared variables for the module
func (w *Writer) Variables() map[string]*configs.Variable {
	return w.module.Variables
}

// RemoteStateDataSources returns a list of remote state data sources defined for the module
func (w *Writer) RemoteStateDataSources() []*configs.Resource {
	resources := make([]*configs.Resource, 0)

	for _, resource := range w.module.DataResources {
		if resource.Type == "terraform_remote_state" {
			resources = append(resources, resource)
		}
	}

	return resources
}

// File returns an existing file object or creates and caches one
func (w *Writer) File(path string) (*hclwrite.File, hcl.Diagnostics) {
	file, ok := w.files[path]
	if ok {
		return file, nil
	}

	b, err := afero.ReadFile(w.fs, path)
	if err != nil && !os.IsNotExist(err) {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "file read error",
				Detail:   fmt.Sprintf("file %s could not be read: %v", path, err),
			},
		}
	}

	var diags hcl.Diagnostics
	if os.IsNotExist(err) {
		file = hclwrite.NewEmptyFile()
	} else {
		file, diags = hclwrite.ParseConfig(b, path, hcl.InitialPos)
	}

	if file != nil {
		w.files[path] = file
	}

	return file, diags
}
