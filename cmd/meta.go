package cmd

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-tfe"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform/command/cliconfig"
	"github.com/hashicorp/terraform/plugin/discovery"
	"github.com/mitchellh/cli"
)

// Meta stores values and methods needed by all commands
type Meta struct {
	UI  cli.Ui
	API *tfe.Client
}

func (m *Meta) flagSet(name string) *flag.FlagSet {
	f := flag.NewFlagSet(name, flag.ContinueOnError)

	return f
}

func (m *Meta) LoadConfig(host string) error {
	tfeConfig := tfe.DefaultConfig()
	tfeConfig.Address = fmt.Sprintf("https://%s", host)

	if tfeConfig.Token == "" {
		config, diags := cliconfig.LoadConfig()
		if diags.HasErrors() {
			return diags.Err()
		}

		source, err := config.CredentialsSource(discovery.FindPlugins("credentials", globalPluginDirs()))
		if err != nil {
			return err
		}

		creds, err := source.ForHost(svchost.Hostname(host))
		if err != nil {
			return err
		}

		tfeConfig.Token = creds.Token()
	}

	if tfeConfig.Token == "" {
		return fmt.Errorf("No Terraform Cloud API token set for %s", host)
	}

	client, err := tfe.NewClient(tfeConfig)
	m.API = client
	return err
}

func getToken(config *cliconfig.Config, host string) string {
	if api, ok := config.Credentials[host]; ok {
		if token, ok := api["token"]; ok {
			if str, ok := token.(string); ok {
				return str
			}
		}
	}

	return ""
}

func flagUsage(flags *flag.FlagSet) string {
	var out string

	flags.VisitAll(func(f *flag.Flag) {
		s := fmt.Sprintf("  --%s", f.Name) // Two spaces before -; see next two comments.
		name, usage := flag.UnquoteUsage(f)
		if len(name) > 0 {
			s += " " + name
		}
		s += "\n    \t"
		s += strings.ReplaceAll(usage, "\n", "\n    \t")

		out += s + "\n\n"
	})

	return out
}

// https://github.com/hashicorp/terraform/blob/da67d86d7f45b7985f875ad473e13571be7aec45/plugins.go#L18
func globalPluginDirs() []string {
	var ret []string
	// Look in ~/.terraform.d/plugins/ , or its equivalent on non-UNIX
	dir, err := cliconfig.ConfigDir()
	if err != nil {
		log.Printf("[ERROR] Error finding global config directory: %s", err)
	} else {
		machineDir := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
		ret = append(ret, filepath.Join(dir, "plugins"))
		ret = append(ret, filepath.Join(dir, "plugins", machineDir))
	}

	return ret
}
