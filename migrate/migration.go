package migrate

import (
	"github.com/bendrucker/terraform-cloud-cli/migrate/configwrite"
	"github.com/hashicorp/hcl/v2"
)

func New(path string, config Config) (*Migration, hcl.Diagnostics) {
	writer, diags := configwrite.New(path)
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

	return &Migration{steps}, diags
}

type Migration struct {
	steps configwrite.Steps
}

func (m *Migration) Changes() (configwrite.Changes, hcl.Diagnostics) {
	return m.steps.Changes()
}
