package migrate

import (
	"github.com/bendrucker/terraform-cloud-cli/migrate/configwrite"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform/configs"
)

func New(path string, config Config) (*Migration, error) {
	parser := configs.NewParser(nil)

	module, diags := parser.LoadConfigDir(path)
	if diags.HasErrors() {
		return nil, diags
	}

	writer := configwrite.New(module)
	steps := configwrite.NewSteps(writer, configwrite.Steps{
		&configwrite.RemoteBackend{Config: config.Backend},
		&configwrite.TerraformWorkspace{Variable: config.WorkspaceVariable},
		&configwrite.Tfvars{Filename: configwrite.TfvarsFilename},
	})

	if config.ModulesDir != "" {
		step := &configwrite.RemoteState{
			RemoteBackend: config.Backend,
			Path:          config.ModulesDir,
		}
		step.WithWriter(writer)
		steps = steps.Append(step)
	}

	return &Migration{
		config: config,
		client: config.Client,
		module: module,
		steps:  steps,
	}, nil
}

type Migration struct {
	config Config
	client *tfe.Client
	module *configs.Module
	steps  configwrite.Steps
}

func (m *Migration) MultipleWorkspaces() bool {
	return m.config.Backend.Workspaces.Prefix != ""
}

func (m *Migration) Changes() (configwrite.Changes, hcl.Diagnostics) {
	return m.steps.Changes()
}
