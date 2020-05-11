package migrate

import (
	"github.com/bendrucker/terraform-cloud-cli/migrate/configwrite"
	"github.com/hashicorp/go-tfe"
)

type Config struct {
	Backend           configwrite.RemoteBackendConfig
	WorkspaceVariable string
	TfvarsFilename    string
	ModulesDir        string

	Client *tfe.Client
}

type RemoteBackendConfig = configwrite.RemoteBackendConfig
type WorkspaceConfig = configwrite.WorkspaceConfig
