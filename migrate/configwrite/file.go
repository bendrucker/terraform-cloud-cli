package configwrite

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/afero"
)

func ExistingFile(path string, fs afero.Fs) (*File, error) {
	b, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(path)

	f, diags := hclwrite.ParseConfig(b, filename, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, err
	}

	return &File{
		hcl:           f,
		Dir:           filepath.Dir(path),
		OriginalName:  filename,
		OriginalBytes: b,
	}, nil
}

func NewFile(path string) *File {
	return &File{
		hcl:     hclwrite.NewEmptyFile(),
		Dir:     filepath.Dir(path),
		NewName: filepath.Base(path),
	}
}

type File struct {
	hcl *hclwrite.File

	Dir           string
	OriginalName  string
	OriginalBytes []byte
	NewName       string
}

func (f *File) Destination() string {
	name := f.OriginalName
	if f.NewName != "" {
		name = f.NewName
	}

	return filepath.Join(f.Dir, name)
}

func (f *File) Diff() difflib.UnifiedDiff {
	var from string
	if f.OriginalName != "" {
		from = filepath.Join(f.Dir, f.OriginalName)
	}

	return difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(f.OriginalBytes)),
		B:        difflib.SplitLines(string(f.hcl.Bytes())),
		FromFile: from,
		ToFile:   f.Destination(),
		Context:  3,
	}
}

func (f *File) Write() error {
	file, err := os.OpenFile(f.Destination(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = f.hcl.WriteTo(file)
	if err != nil {
		return err
	}

	if f.NewName != "" {
		return os.Remove(filepath.Join(f.Dir, f.OriginalName))
	}

	return nil
}

// Changes is map of paths to file objects that should be written to prepare the module for Terraform Cloud.
type Changes map[string]*File

func (c Changes) Write() error {
	for _, change := range c {
		if err := change.Write(); err != nil {
			return err
		}
	}

	return nil
}
