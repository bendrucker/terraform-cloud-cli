package configwrite

import (
	"github.com/hashicorp/terraform/configs"
	"github.com/spf13/afero"
)

func New(module *configs.Module) *Writer {
	return newWriter(module, afero.NewOsFs())
}

func newWriter(module *configs.Module, fs afero.Fs) *Writer {
	return &Writer{
		fs:     fs,
		module: module,
		files:  make(map[string]*File),
	}
}

// Writer provides access to information about the Terraform module structure and the ability to update its files
type Writer struct {
	fs     afero.Fs
	module *configs.Module
	files  map[string]*File
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
func (w *Writer) File(path string) (*File, error) {
	file, ok := w.files[path]
	if ok {
		return file, nil
	}

	exists, err := afero.Exists(w.fs, path)
	if err != nil {
		return nil, err
	}

	if !exists {
		file = NewFile(path)
	} else {
		file, err = ExistingFile(path, w.fs)
		if err != nil {
			return nil, err
		}
	}

	w.files[path] = file

	return file, nil
}
