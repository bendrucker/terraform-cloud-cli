package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bendrucker/terraform-cloud-cli/migrate"
	"github.com/bendrucker/terraform-cloud-cli/migrate/configwrite"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2"
)

type MigrateCommand struct {
	*Meta

	WorkspaceName     string
	WorkspacePrefix   string
	WorkspaceVariable string
	TfvarsFilename    string
	ModulesDir        string
}

func (c *MigrateCommand) flags() *flag.FlagSet {
	f := c.flagSet("migrate")

	f.StringVar(&c.WorkspaceName, "workspace-name", "", "The name of the Terraform Cloud workspace (conflicts with --workspace-prefix)")
	f.StringVar(&c.WorkspacePrefix, "workspace-prefix", "", "The prefix of the Terraform Cloud workspaces (conflicts with --workspace-name)")
	f.StringVar(&c.ModulesDir, "modules", "", "A directory where other Terraform modules are stored. If set, it will be scanned recursively for terraform_remote_state references.")
	f.StringVar(&c.WorkspaceVariable, "workspace-variable", "environment", "Variable that will replace terraform.workspace")
	f.StringVar(&c.TfvarsFilename, "tfvars-filename", configwrite.TfvarsAlternateFilename, "New filename for terraform.tfvars")

	return f
}

func (c *MigrateCommand) Run(args []string) int {
	flags := c.flags()

	if err := flags.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	if flags.NArg() != 1 {
		c.UI.Error(fmt.Sprintf("A single module path is required, received %d arguments", flags.NArg()))
		return 1
	}

	path := flags.Args()[0]
	abspath, err := filepath.Abs(path)
	if err != nil {
		c.UI.Error(fmt.Sprintf("failed to resolve path: %s", path))
		return 1
	}

	c.UI.Info(fmt.Sprintf("Upgrading Terraform module %s", abspath))

	if c.WorkspaceName == "" && c.WorkspacePrefix == "" {
		c.UI.Error("workspace name or prefix is required")
		return 1
	}

	if c.WorkspaceName != "" && c.WorkspacePrefix != "" {
		c.UI.Error("workspace cannot have a name and prefix")
		return 1
	}

	if err := c.Meta.LoadConfig(c.config.Hostname); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	migration, err := migrate.New(path, migrate.Config{
		Client: c.Meta.API,

		Backend: migrate.RemoteBackendConfig{
			Hostname:     c.config.Hostname,
			Organization: c.config.Organization,
			Workspaces: migrate.WorkspaceConfig{
				Prefix: c.WorkspacePrefix,
				Name:   c.WorkspaceName,
			},
		},
		WorkspaceVariable: c.WorkspaceVariable,
		TfvarsFilename:    c.TfvarsFilename,
		ModulesDir:        c.ModulesDir,
	})

	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	if migration.MultipleWorkspaces() {
		c.UI.Output("Checking for existing Terraform Cloud workspaces...")

		list, err := c.Meta.API.Workspaces.List(context.TODO(), c.config.Organization, tfe.WorkspaceListOptions{
			Search: tfe.String(c.WorkspacePrefix),
		})
		if err != nil {
			c.UI.Error("Failed to list workspaces: " + err.Error())
			c.UI.Info(`Credentials may be expired or invalid. Re-run "terraform login".`)
			return 1
		}

		if len(list.Items) == 0 {
			c.UI.Warn(fmt.Sprintf("No workspaces found with prefix '%s'", c.WorkspacePrefix))
			fmt.Println()
			c.UI.Info(strings.TrimSpace(`
When "terraform init" runs with the new backend configuration, it will attempt to create new workspaces.
If you are using the "tfe" provider and "tfe_workspace" you should create workspaces via Terraform before proceeding.
`))

			fmt.Println()
			if _, err := c.UI.Ask("Press enter to proceed:"); err != nil {
				c.UI.Error(err.Error())
				return 1
			}
		}
	} else {
		c.UI.Output("Checking for an existing Terraform Cloud workspace...")

		_, err := c.Meta.API.Workspaces.Read(context.Background(), c.config.Organization, c.WorkspaceName)
		if err != nil && err != tfe.ErrResourceNotFound {
			c.UI.Error("Failed to get workspace: " + err.Error())
			c.UI.Info(`Credentials may be expired or invalid. Re-run "terraform login".`)
			return 1
		}

		if err == tfe.ErrResourceNotFound {
			c.UI.Warn(fmt.Sprintf("Workspace named '%s' was not found", c.WorkspaceName))
			fmt.Println()
			c.UI.Info(`When "terraform init" runs with the new backend configuration, it will attempt to create this workspace.`)
			c.UI.Info(`If you are using the "tfe" provider and "tfe_workspace" you should create a workspace via Terraform before proceeding.`)

			fmt.Println()
			if _, err := c.UI.Ask("Press enter to proceed:"); err != nil {
				c.UI.Error(err.Error())
				return 1
			}
		}
	}

	c.UI.Output("Running 'terraform init'...")
	fmt.Println()

	if code := c.terraformInit(abspath); code != 0 {
		return code
	}

	changes, diags := migration.Changes()
	if diags.HasErrors() {
		c.printDiags(diags)
		return 1
	}

	if err := changes.WriteFiles(); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	for path, change := range changes {
		str := path
		if change.Rename != "" {
			str = fmt.Sprintf("%s -> %s", path, change.Destination(path))
		}

		fmt.Println(str)
	}

	c.UI.Info("Running 'terraform init' to copy state")
	c.UI.Info("When prompted, type 'yes' to confirm")
	fmt.Println()

	if code := c.terraformInit(abspath); code != 0 {
		return code
	}

	c.UI.Info("Migration complete!")
	c.UI.Info("If your workspace is VCS-enabled, commit these changes and push to trigger a run.")
	c.UI.Info("If not, you can now call 'terraform plan' and 'terraform apply' locally.")

	return 0
}

func (c *MigrateCommand) Help() string {
	return strings.TrimSpace(`
Usage: terraform-cloud migrate [DIR] [options]
  Migrate a Terraform module to Terraform Cloud

Options:
` + flagUsage(c.flags()))
}

func (c *MigrateCommand) Synopsis() string {
	return "Migrate a Terraform module from an existing backend to Terraform Cloud"
}

func (c *MigrateCommand) printDiags(diags hcl.Diagnostics) {
	for _, diag := range diags {
		switch diag.Severity {
		case hcl.DiagError:
			c.UI.Error(diag.Summary)
		case hcl.DiagWarning:
			c.UI.Warn(diag.Summary)
		}
		c.UI.Info(diag.Detail)
		if diag.Subject != nil {
			c.UI.Info(diag.Subject.String())
		}
	}
}

func (c *MigrateCommand) terraformInit(path string) int {
	cmd := exec.Command("terraform", "init", path)

	cmd.Dir = path

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			return err.ExitCode()
		}

		c.UI.Error(fmt.Sprintf("failed to terraform init: %v", err))
	}

	return 0
}
