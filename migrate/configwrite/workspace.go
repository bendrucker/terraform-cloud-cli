package configwrite

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
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

// Complete checks if any terraform.workspace replaces are proposed
func (s *TerraformWorkspace) Complete() bool {
	files, _ := s.files()
	for _, file := range files {
		if hasTerraformWorkspace(file.Body()) {
			return false
		}
	}

	return true
}

// Description returns a description of the step
func (s *TerraformWorkspace) Description() string {
	return `terraform.workpace will always be set to default and should not be used with Terraform Cloud (https://www.terraform.io/docs/state/workspaces.html#current-workspace-interpolation)`
}

func (s *TerraformWorkspace) files() (map[string]*hclwrite.File, hcl.Diagnostics) {
	files, _, diags := s.writer.parser.ConfigDirFiles(s.writer.Dir())
	out := make(map[string]*hclwrite.File, len(files))

	for _, path := range files {
		file, fDiags := s.writer.File(path)
		out[path] = file
		diags = append(diags, fDiags...)
	}

	return out, diags
}

// Changes determines changes required to remove terraform.workspace
func (s *TerraformWorkspace) Changes() (Changes, hcl.Diagnostics) {
	files, diags := s.files()

	changes := make(Changes)
	for path, file := range files {
		if hasTerraformWorkspace(file.Body()) {
			replaceTerraformWorkspace(file.Body(), s.Variable)
			changes[path] = &Change{File: file}
		}
	}

	if len(changes) == 0 {
		return changes, diags
	}

	if _, ok := s.writer.Variables()[s.Variable]; !ok {
		path := filepath.Join(s.writer.Dir(), "variables.tf")
		file, fDiags := s.writer.File(path)
		diags = append(diags, fDiags...)

		changes[path] = &Change{
			File: addWorkspaceVariable(file, s.Variable),
		}
	}

	return changes, diags
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

func changedFiles(sources map[string][]byte, changes Changes) (Changes, hcl.Diagnostics) {
	changed := make(Changes)

	for path, change := range changes {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, hcl.Diagnostics{
				&hcl.Diagnostic{
					Summary: "file read error",
					Detail:  fmt.Sprintf("could not read file %s", path),
				},
			}
		}

		if bytes.Equal(b, change.File.Bytes()) {
			continue
		}

		changed[path] = &Change{File: change.File}
	}

	return changed, nil
}

func addWorkspaceVariable(file *hclwrite.File, name string) *hclwrite.File {
	variable := hclwrite.NewBlock("variable", []string{name})
	variable.Body().SetAttributeRaw("type", hclwrite.Tokens{
		{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte("string"),
		},
	})
	variable.Body().SetAttributeValue("description", cty.StringVal(fmt.Sprintf("The %s where the module will be deployed", name)))

	body := file.Body()
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
