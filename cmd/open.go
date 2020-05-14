package cmd

import (
	"flag"
	"fmt"

	"github.com/bendrucker/terraform-cloud-cli/backend"
	"github.com/hashicorp/terraform/configs"
	"github.com/skratchdot/open-golang/open"
)

type OpenCommand struct {
	*Meta

	Workspace string
}

func (c *OpenCommand) Run(args []string) int {
	flags := c.flags()

	if err := flags.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	if flags.NArg() > 1 {
		c.UI.Error("1 argument supported")
		return 1
	}

	var path string
	if flags.NArg() != 0 {
		path = flags.Arg(0)
	}

	parser := configs.NewParser(nil)
	module, diags := parser.LoadConfigDir(path)
	if diags.HasErrors() {
		c.UI.Error(diags.Error())
		return 1
	}

	if !backend.IsRemote(module.Backend) {
		c.UI.Error("Remote backend not found")
		c.UI.Info("\nTo open a Terraform Cloud workspace, your configuration must include backend configuration:")
		c.UI.Info(exampleRemoteBackend)
		return 1
	}

	remote, err := backend.DecodeConfig(module.Backend)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	if c.Workspace != "" && !remote.Workspaces.Multiple() {
		c.UI.Error("--workspace can only be used when a prefix is set")
		return 1
	}

	url := c.url(remote)
	c.UI.Output(fmt.Sprintf(`Opening "%s"`, url))

	if err := open.Run(url); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	return 0
}

func (c *OpenCommand) url(remote *backend.RemoteBackend) string {
	url := fmt.Sprintf("https://%s/app/%s/workspaces", *remote.Hostname, *remote.Organization)

	if remote.Workspaces.Multiple() {
		if c.Workspace == "" {
			return fmt.Sprintf("%s?search=%s", url, *remote.Workspaces.Prefix)
		}

		return fmt.Sprintf("%s/%s%s", url, *remote.Workspaces.Prefix, c.Workspace)
	}

	return fmt.Sprintf("%s/%s", url, *remote.Workspaces.Name)
}

func (c *OpenCommand) Synopsis() string {
	return "Opens the Terraform Cloud UI to the workspace"
}

func (c *OpenCommand) flags() *flag.FlagSet {
	f := c.flagSet("open")

	f.StringVar(&c.Workspace, "workspace", "", "The workspace to open. If the configuration does not set a prefix, setting this is an error.")

	return f
}

func (c *OpenCommand) Help() string {
	return `
Usage: terraform-cloud open [DIR] [options]
  Opens the Terraform Cloud UI for the provided (or current) directory. If a single workspace is defined,
  it will be opened directly. If the module allow multiple workspaces (by setting "prefix"), the workspace
  list will be opened with the prefix set as the search, unless --workspace is set.

Options:
` + flagUsage(c.flags())
}

const exampleRemoteBackend = `
terraform {
  backend "remote" {
    // ...
  }
}
`
