package migrate

import (
	"context"
	"errors"

	"github.com/bendrucker/terraform-cloud-cli/migrate/configwrite"
	"github.com/hashicorp/go-tfe"
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

var (
	ErrWorkspaceLocked  = tfe.ErrWorkspaceLocked
	ErrResourceNotFound = tfe.ErrResourceNotFound
)

func (m *Migration) MultipleWorkspaces() bool {
	return m.config.Backend.Workspaces.Prefix != ""
}

func (m *Migration) RemoteWorkspaces() ([]*tfe.Workspace, error) {
	if m.MultipleWorkspaces() {
		list, err := m.client.Workspaces.List(context.TODO(), m.config.Backend.Organization, tfe.WorkspaceListOptions{
			Search: tfe.String(m.config.Backend.Workspaces.Prefix),
		})
		if err != nil {
			return nil, err
		}

		return list.Items, nil
	}

	ws, err := m.client.Workspaces.Read(context.TODO(), m.config.Backend.Organization, m.config.Backend.Workspaces.Name)
	if err != nil {
		if errors.Is(err, tfe.ErrResourceNotFound) {
			return []*tfe.Workspace{}, nil
		}

		return nil, err
	}

	return []*tfe.Workspace{ws}, nil
}

func (m *Migration) Changes() (configwrite.Changes, error) {
	return m.steps.Changes()
}
