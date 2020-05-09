package cmd

import (
	"flag"
	"fmt"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/mitchellh/cli"
)

// Meta stores values and methods needed by all commands
type Meta struct {
	config MetaConfig
	UI     cli.Ui
	API    *tfe.Client
}

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

func flagUsage(flags *flag.FlagSet) string {
	var out string

	flags.VisitAll(func(f *flag.Flag) {
		s := fmt.Sprintf("  -%s", f.Name) // Two spaces before -; see next two comments.
		name, usage := flag.UnquoteUsage(f)
		if len(name) > 0 {
			s += " " + name
		}
		s += "\n    \t"
		s += strings.ReplaceAll(usage, "\n", "\n    \t")

		out += s + "\n"
	})

	return out
}
