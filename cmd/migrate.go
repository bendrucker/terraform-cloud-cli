package cmd

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	migrate "github.com/bendrucker/terraform-cloud-migrate"
	"github.com/bendrucker/terraform-cloud-migrate/configwrite"
	"github.com/hashicorp/hcl/v2"
)

type MigrateCommand struct {
	*Meta

	WorkspaceName     string
	WorkspacePrefix   string
	WorkspaceVariable string
	TfvarsFilename    string
	ModulesDir        string
	NoInit            bool
}

func (c *MigrateCommand) flags() *flag.FlagSet {
	f := c.flagSet("migrate")

	f.StringVar(&c.WorkspaceName, "workspace-name", "", "The name of the Terraform Cloud workspace (conflicts with --workspace-prefix)")
	f.StringVar(&c.WorkspacePrefix, "workspace-prefix", "", "The prefix of the Terraform Cloud workspaces (conflicts with --workspace-name)")
	f.StringVar(&c.ModulesDir, "modules", "", "A directory where other Terraform modules are stored. If set, it will be scanned recursively for terraform_remote_state references.")
	f.StringVar(&c.WorkspaceVariable, "workspace-variable", "environment", "Variable that will replace terraform.workspace")
	f.StringVar(&c.TfvarsFilename, "tfvars-filename", configwrite.TfvarsAlternateFilename, "New filename for terraform.tfvars")

	f.BoolVar(&c.NoInit, "no-init", false, "Disable calling 'terraform init' before and after updating configuration to copy state.")

	return f
}

func (c *MigrateCommand) Run(args []string) int {
	flags := c.flags()

	if err := flags.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	if len(flags.Args()) != 1 {
		c.UI.Error("module path is required")
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

	migration, diags := migrate.New(path, migrate.Config{
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

	if diags.HasErrors() {
		c.printDiags(diags)
		return 1
	}

	changes, diags := migration.Changes()
	if diags.HasErrors() {
		c.printDiags(diags)
		return 1
	}

	if !c.NoInit {
		c.UI.Info("Running 'terraform init' prior to updating backend")
		c.UI.Info("This ensures that Terraform has persisted the existing backend configuration to local state")
		fmt.Println()

		if code := c.terraformInit(abspath); code != 0 {
			return code
		}
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

	if !c.NoInit {
		c.UI.Info("Running 'terraform init' to copy state")
		c.UI.Info("When prompted, type 'yes' to confirm")
		fmt.Println()

		if code := c.terraformInit(abspath); code != 0 {
			return code
		}
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
		c.UI.Info(diag.Subject.String())
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
