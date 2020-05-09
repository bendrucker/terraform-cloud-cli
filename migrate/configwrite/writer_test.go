package configwrite

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/lithammer/dedent"
	"github.com/spf13/afero"
)

func newTestWriter(t *testing.T, path string, setup func(afero.Fs)) *Writer {
	fs := afero.NewMemMapFs()
	setup(fs)
	writer, diags := newWriter(path, fs)
	if len(diags) != 0 {
		for _, diag := range diags {
			t.Error(diag.Error())
		}
		t.FailNow()
	}
	return writer
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
