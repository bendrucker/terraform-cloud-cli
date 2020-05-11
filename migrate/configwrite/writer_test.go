package configwrite

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/configs"
	"github.com/lithammer/dedent"
	"github.com/spf13/afero"
)

func newTestWriter(t *testing.T, path string, setup func(afero.Fs)) *Writer {
	fs := afero.NewMemMapFs()
	setup(fs)
	parser := configs.NewParser(fs)

	module, diags := parser.LoadConfigDir(path)
	if diags.HasErrors() {
		t.Fatalf("failed to load module: %v", diags)
	}

	return newWriter(module, fs)
}

func newTestModule(t *testing.T, files map[string]string) *Writer {
	return newTestWriter(t, "", func(fs afero.Fs) {
		for name, content := range files {
			if err := fs.MkdirAll(filepath.Dir(name), 0600); err != nil {
				t.Error(err)
			}
			if err := afero.WriteFile(fs, name, []byte(trimTestConfig(content)), 0644); err != nil {
				t.Error(err)
			}
		}
	})
}

func trimTestConfig(config string) string {
	config = dedent.Dedent(config)
	config = strings.ReplaceAll(config, "\t", "  ")
	config = strings.TrimLeft(config, "\n")

	return config
}
