package cmd

import (
	"flag"
	"fmt"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform/command/cliconfig"
	"github.com/mitchellh/cli"
)

// Meta stores values and methods needed by all commands
type Meta struct {
	config MetaConfig
	UI     cli.Ui
	API    *tfe.Client
}

// MetaConfig stores configuration that can be set in all commands
type MetaConfig struct {
	Hostname     string
	Organization string
}

func (m *Meta) flagSet(name string) *flag.FlagSet {
	f := flag.NewFlagSet(name, flag.ContinueOnError)

	f.StringVar(&m.config.Hostname, "hostname", "app.terraform.io", "Hostname for Terraform Cloud")
	f.StringVar(&m.config.Organization, "organization", "", "Organization name in Terraform Cloud")

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

		diags = config.Validate()
		if diags.HasErrors() {
			return diags.Err()
		}

		tfeConfig.Token = getToken(config, host)
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
