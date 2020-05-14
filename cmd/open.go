package cmd

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform/configs"
	"github.com/skratchdot/open-golang/open"
	"github.com/zclconf/go-cty/cty"
)

type OpenCommand struct {
	*Meta
}

func (c *OpenCommand) Run(args []string) int {
	if len(args) > 1 {
		c.UI.Error("1 argument supported")
		return 1
	}

	var path string
	if len(args) != 0 {
		path = args[0]
	}

	c.LoadConfig(args[0])

	parser := configs.NewParser(nil)
	module, diags := parser.LoadConfigDir(path)
	if diags.HasErrors() {
		c.UI.Error(diags.Error())
		return 1
	}

	if module.Backend == nil || module.Backend.Type != "remote" {
		c.UI.Error("Remote backend not found")
		c.UI.Info("\nTo open a Terraform Cloud workspace, your configuration must include backend configuration:")
		c.UI.Info(exampleRemoteBackend)
		return 1
	}

	content, _, diags := module.Backend.Config.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: "hostname",
			},
			{
				Name:     "organization",
				Required: true,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "workspaces",
			},
		},
	})
	if diags.HasErrors() {
		c.UI.Error(diags.Error())
		return 1
	}

	host, _ := content.Attributes["hostname"].Expr.Value(&hcl.EvalContext{})
	org, _ := content.Attributes["organization"].Expr.Value(&hcl.EvalContext{})

	content, _, diags = content.Blocks.OfType("workspaces")[0].Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "name"},
			{Name: "prefix"},
		},
	})
	if diags.HasErrors() {
		c.UI.Error(diags.Error())
		return 1
	}

	name, _ := content.Attributes["name"].Expr.Value(&hcl.EvalContext{})

	if err := open.Run(fmt.Sprintf("https://%s/app/%s/workspaces/%s", host.AsString(), org.AsString(), name.AsString())); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	return 0
}

func (c *OpenCommand) LoadConfig(dir string) error {
	parser := configs.NewParser(nil)
	module, diags := parser.LoadConfigDir(dir)
	if diags.HasErrors() {
		return diags
	}

	if module.Backend == nil || module.Backend.Type != "remote" {
		return errors.New("remote backend not found")
	}

	content, _, diags := module.Backend.Config.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: "hostname",
			},
			{
				Name:     "organization",
				Required: true,
			},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "workspaces",
			},
		},
	})
	if diags.HasErrors() {
		return diags
	}

	var host, organization, name, prefix cty.Value

	if attr, ok := content.Attributes["hostname"]; ok {
		host, diags = attr.Expr.Value(&hcl.EvalContext{})
		if diags.HasErrors() {
			return diags
		}
	}

	if attr, ok := content.Attributes["organization"]; ok {
		organization, diags = attr.Expr.Value(&hcl.EvalContext{})
		if diags.HasErrors() {
			return diags
		}
	}

	content, _, diags = content.Blocks.OfType("workspaces")[0].Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "name"},
			{Name: "prefix"},
		},
	})
	if diags.HasErrors() {
		return diags
	}

	if attr, ok := content.Attributes["name"]; ok {
		name, diags = attr.Expr.Value(&hcl.EvalContext{})
		if diags.HasErrors() {
			return diags
		}
	}

	if attr, ok := content.Attributes["prefix"]; ok {
		prefix, diags = attr.Expr.Value(&hcl.EvalContext{})
		if diags.HasErrors() {
			return diags
		}
	}

	if prefix == cty.NilVal {

	}

	fmt.Println(host.AsString(), organization.AsString(), name.AsString())
	return nil
}

func (c *OpenCommand) Synopsis() string {
	return ""
}

func (c *OpenCommand) Help() string {
	return ""
}

const exampleRemoteBackend = `
terraform {
  backend "remote" {
    // ...
  }
}
`
