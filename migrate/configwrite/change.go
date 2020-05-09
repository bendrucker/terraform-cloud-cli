package configwrite

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

type Change struct {
	File   *hclwrite.File
	Rename string
}

func (c *Change) Destination(path string) string {
	if c.Rename == "" {
		return path
	}

	return filepath.Join(filepath.Dir(path), c.Rename)
}

func (c *Change) WriteFile(path string) error {
	file, err := os.OpenFile(c.Destination(path), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = c.File.WriteTo(file)
	if err != nil {
		return err
	}

	if c.Rename != "" {
		return os.Remove(path)
	}

	return nil
}

// Changes is a map of changed file objects that should be written to prepare the module for Terraform Cloud
type Changes map[string]*Change

func (c Changes) Add(path string, change *Change) error {
	if existing, ok := c[path]; ok && existing.Rename != "" && change.Rename != "" {
		return &renameCollisionError{Existing: existing.Rename, Proposed: change.Rename}
	}

	c[path] = change

	return nil
}

func (c Changes) WriteFiles() error {
	for path, change := range c {
		if err := change.WriteFile(path); err != nil {
			return err
		}
	}

	return nil
}

type renameCollisionError struct {
	Existing string
	Proposed string
}

func (e *renameCollisionError) Error() string {
	return fmt.Sprintf("cannot rename to '%s', already renamed to '%s'", e.Proposed, e.Existing)
}
