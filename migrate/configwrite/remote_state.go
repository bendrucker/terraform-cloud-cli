package configwrite

import (
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform/configs"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

type RemoteState struct {
	writer        *Writer
	Path          string
	RemoteBackend RemoteBackendConfig
}

func (s *RemoteState) WithWriter(w *Writer) Step {
	s.writer = w
	return s
}

func (s *RemoteState) Name() string {
	return "Update terraform_remote_state"
}

// Description returns a description of the step
func (s *RemoteState) Description() string {
	return `A "remote" backend should be configured for Terraform Cloud (https://www.terraform.io/docs/backends/types/remote.html)`
}

// Changes updates the configured backend
func (s *RemoteState) Changes() (Changes, error) {
	changes := Changes{}
	parser := configs.NewParser(s.writer.fs)

	return changes, afero.Walk(s.writer.fs, s.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() || !parser.IsConfigDir(path) {
			return nil
		}

		sources, diags := s.sources(path)
		if diags.HasErrors() {
			return diags
		}

		for _, source := range sources {
			filepath := source.DeclRange.Filename
			file, err := s.writer.File(source.DeclRange.Filename)
			if err != nil {
				return err
			}

			block := file.hcl.Body().FirstMatchingBlock("data", []string{
				source.Type,
				source.Name,
			})

			workspace := block.Body().RemoveAttribute("workspace")

			block.Body().SetAttributeValue("backend", cty.StringVal("remote"))
			block.Body().SetAttributeRaw("config", flattenTokens([]hclwrite.Tokens{
				{
					{
						Type:  hclsyntax.TokenOBrace,
						Bytes: []byte("{"),
					},
					{
						Type:  hclsyntax.TokenNewline,
						Bytes: []byte("\n"),
					},
					{
						Type:  hclsyntax.TokenIdent,
						Bytes: []byte("hostname"),
					},
					{
						Type:  hclsyntax.TokenEqual,
						Bytes: []byte("="),
					},
					{
						Type:  hclsyntax.TokenOQuote,
						Bytes: []byte(`"`),
					},
					{
						Type:  hclsyntax.TokenQuotedLit,
						Bytes: []byte(s.RemoteBackend.Hostname),
					},
					{
						Type:  hclsyntax.TokenCQuote,
						Bytes: []byte(`"`),
					},
					{
						Type:  hclsyntax.TokenNewline,
						Bytes: []byte("\n"),
					},
					{
						Type:  hclsyntax.TokenIdent,
						Bytes: []byte("organization"),
					},
					{
						Type:  hclsyntax.TokenEqual,
						Bytes: []byte("="),
					},
					{
						Type:  hclsyntax.TokenOQuote,
						Bytes: []byte(`"`),
					},
					{
						Type:  hclsyntax.TokenQuotedLit,
						Bytes: []byte(s.RemoteBackend.Organization),
					},
					{
						Type:  hclsyntax.TokenCQuote,
						Bytes: []byte(`"`),
					},
					{
						Type:  hclsyntax.TokenNewline,
						Bytes: []byte("\n\n"),
					},
					{
						Type:  hclsyntax.TokenIdent,
						Bytes: []byte("workspaces"),
					},
					{
						Type:  hclsyntax.TokenEqual,
						Bytes: []byte("="),
					},
					{
						Type:  hclsyntax.TokenOBrace,
						Bytes: []byte("{"),
					},
					{
						Type:  hclsyntax.TokenNewline,
						Bytes: []byte("\n"),
					},
					{
						Type:  hclsyntax.TokenStringLit,
						Bytes: []byte("name"),
					},
					{
						Type:  hclsyntax.TokenEqual,
						Bytes: []byte("="),
					},
				},
				s.workspaceNameTokens(workspace),
				{
					{
						Type:  hclsyntax.TokenNewline,
						Bytes: []byte("\n"),
					},
					{
						Type:  hclsyntax.TokenCBrace,
						Bytes: []byte("}"),
					},
					{
						Type:  hclsyntax.TokenNewline,
						Bytes: []byte("\n"),
					},
					{
						Type:  hclsyntax.TokenCBrace,
						Bytes: []byte("}"),
					},
				},
			}))

			changes[filepath] = file
		}

		if diags.HasErrors() {
			return diags
		}

		return nil
	})

	return changes, nil
}

// Changes updates the configured backend
func (s *RemoteState) sources(path string) ([]*configs.Resource, hcl.Diagnostics) {
	parser := configs.NewParser(s.writer.fs)
	module, diags := parser.LoadConfigDir(path)

	sources := make([]*configs.Resource, 0)
	writer := newWriter(module, s.writer.fs)

Source:
	for _, source := range writer.RemoteStateDataSources() {
		attrs, diags := source.Config.JustAttributes()

		backend, bDiags := attrs["backend"].Expr.Value(nil)
		diags = append(diags, bDiags...)

		if backend.AsString() != s.writer.Backend().Type {
			continue
		}

		config, cDiags := attrs["config"].Expr.Value(nil)
		// errors on interpolations
		if cDiags.HasErrors() {
			continue
		}
		diags = append(diags, cDiags...)

		remoteBackendConfigAttrs, rDiags := s.writer.Backend().Config.JustAttributes()
		// errors when workspaces is block
		if rDiags.HasErrors() {
			continue
		}
		diags = append(diags, rDiags...)

		for key, value := range config.AsValueMap() {
			// workspaces is a block
			if key == "workspaces" {
				continue Source
			}

			rbValue, rDiags := remoteBackendConfigAttrs[key].Expr.Value(nil)
			diags = append(diags, rDiags...)

			if value.AsString() != rbValue.AsString() {
				continue Source
			}
		}

		sources = append(sources, source)
	}

	if diags.HasErrors() {
		return nil, diags
	}

	return sources, nil
}

func (s *RemoteState) workspaceNameTokens(workspace *hclwrite.Attribute) hclwrite.Tokens {
	if s.RemoteBackend.Workspaces.Prefix == "" {
		return hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenOQuote,
				Bytes: []byte(`"`),
			},
			{
				Type:  hclsyntax.TokenStringLit,
				Bytes: []byte(s.RemoteBackend.Workspaces.Name),
			},
			{
				Type:  hclsyntax.TokenCQuote,
				Bytes: []byte(`"`),
			},
		}
	}

	return flattenTokens([]hclwrite.Tokens{
		{
			{
				Type:  hclsyntax.TokenOQuote,
				Bytes: []byte(`"`),
			},
			{
				Type:  hclsyntax.TokenStringLit,
				Bytes: []byte(s.RemoteBackend.Workspaces.Prefix),
			},
			{
				Type:  hclsyntax.TokenTemplateInterp,
				Bytes: []byte("${"),
			},
		},
		workspace.Expr().BuildTokens(nil),
		{
			{
				Type:  hclsyntax.TokenTemplateSeqEnd,
				Bytes: []byte("}"),
			},
			{
				Type:  hclsyntax.TokenCQuote,
				Bytes: []byte(`"`),
			},
		},
	})
}

func flattenTokens(in []hclwrite.Tokens) hclwrite.Tokens {
	out := make(hclwrite.Tokens, 0)
	for _, tokens := range in {
		out = append(out, tokens...)
	}
	return out
}

var _ Step = (*RemoteState)(nil)
