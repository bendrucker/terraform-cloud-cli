package migrate

import "github.com/bendrucker/terraform-cloud-cli/migrate/configwrite"

type Config struct {
	Backend           configwrite.RemoteBackendConfig
	WorkspaceVariable string
	TfvarsFilename    string
	ModulesDir        string
}

type RemoteBackendConfig = configwrite.RemoteBackendConfig
type WorkspaceConfig = configwrite.WorkspaceConfig
