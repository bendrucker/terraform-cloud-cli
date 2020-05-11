package configwrite

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

func NewFile(path string, hcl *hclwrite.File) *File {
	return &File{
		hcl:          hcl,
		Dir:          filepath.Dir(path),
		OriginalName: filepath.Base(path),
	}
}

type File struct {
	hcl *hclwrite.File

	Dir          string
	OriginalName string
	NewName      string
}

func (f *File) Destination() string {
	name := f.OriginalName
	if f.NewName != "" {
		name = f.NewName
	}

	return filepath.Join(f.Dir, name)
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

// Changes is map of paths to file objects that should be written to prepare the module for Terraform Cloud
type Changes map[string]*File

func (c Changes) Write() error {
	for _, change := range c {
		if err := change.Write(); err != nil {
			return err
		}
	}

	return nil
}
