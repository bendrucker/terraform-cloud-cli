package configwrite

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform/configs"
	"github.com/zclconf/go-cty/cty"
)

type TerraformWorkspace struct {
	writer   *Writer
	Variable string
}

func (s *TerraformWorkspace) WithWriter(w *Writer) Step {
	s.writer = w
	return s
}

func (s *TerraformWorkspace) Name() string {
	return "Replace terraform.workspace"
}

// Complete checks if any terraform.workspace replaces are proposed.
func (s *TerraformWorkspace) Complete() bool {
	files, _ := s.files()
	for _, file := range files {
		if hasTerraformWorkspace(file.hcl.Body()) {
			return false
		}
	}

	return true
}

// Description returns a description of the step.
func (s *TerraformWorkspace) Description() string {
	return `terraform.workpace will always be set to default and should not be used with Terraform Cloud (https://www.terraform.io/docs/state/workspaces.html#current-workspace-interpolation)`
}

func (s *TerraformWorkspace) files() (map[string]*File, error) {
	parser := configs.NewParser(s.writer.fs)
	files, _, diags := parser.ConfigDirFiles(s.writer.Dir())
	out := make(map[string]*File, len(files))

	for _, path := range files {
		file, err := s.writer.File(path)
		if err != nil {
			return nil, err
		}

		out[path] = file
	}

	if diags.HasErrors() {
		return nil, diags
	}

	return out, nil
}

// Changes determines changes required to remove terraform.workspace.
func (s *TerraformWorkspace) Changes() (Changes, error) {
	files, err := s.files()
	if err != nil {
		return Changes{}, err
	}

	changes := make(Changes)

	for path, file := range files {
		if hasTerraformWorkspace(file.hcl.Body()) {
			replaceTerraformWorkspace(file.hcl.Body(), s.Variable)
			changes[path] = file
		}
	}

	if len(changes) == 0 {
		return changes, nil
	}

	if _, ok := s.writer.Variables()[s.Variable]; !ok {
		path := filepath.Join(s.writer.Dir(), "variables.tf")

		file, err := s.writer.File(path)
		if err != nil {
			return Changes{}, err
		}

		changes[path] = addWorkspaceVariable(file, s.Variable)
	}

	return changes, err
}

func hasTerraformWorkspace(body *hclwrite.Body) bool {
	for _, attr := range body.Attributes() {
		for _, traversal := range attr.Expr().Variables() {
			if tokensEqual(traversal.BuildTokens(nil), hclwrite.Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("terraform"),
				},
				{
					Type:  hclsyntax.TokenDot,
					Bytes: []byte("."),
				},
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte("workspace"),
				},
			}) {
				return true
			}
		}
	}

	for _, block := range body.Blocks() {
		if hasTerraformWorkspace(block.Body()) {
			return true
		}
	}

	return false
}

func replaceTerraformWorkspace(body *hclwrite.Body, variable string) {
	for _, attr := range body.Attributes() {
		attr.Expr().RenameVariablePrefix(
			[]string{"terraform", "workspace"},
			[]string{"var", variable},
		)
	}

	for _, block := range body.Blocks() {
		replaceTerraformWorkspace(block.Body(), variable)
	}
}

func addWorkspaceVariable(file *File, name string) *File {
	variable := hclwrite.NewBlock("variable", []string{name})
	variable.Body().SetAttributeRaw("type", hclwrite.Tokens{
		{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte("string"),
		},
	})
	variable.Body().SetAttributeValue("description", cty.StringVal(fmt.Sprintf("The %s where the module will be deployed", name)))

	body := file.hcl.Body()
	existing := body.BuildTokens(nil)

	body.Clear()
	body.AppendBlock(variable)
	body.AppendNewline()
	body.AppendUnstructuredTokens(existing)

	return file
}

func tokensEqual(a hclwrite.Tokens, b hclwrite.Tokens) bool {
	if len(a) != len(b) {
		return false
	}

	for i, at := range a {
		bt := b[i]
		if at.Type != bt.Type || !bytes.Equal(at.Bytes, bt.Bytes) {
			return false
		}
	}

	return true
}

var _ Step = (*TerraformWorkspace)(nil)
