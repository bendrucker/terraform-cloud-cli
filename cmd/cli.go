package cmd

import (
	"os"

	"github.com/mitchellh/cli"
)

func NewCLI() *cli.CLI {
	c := cli.NewCLI("terraform-cloud", "")
	c.Args = os.Args[1:]

	meta := &Meta{
		UI: &cli.ColoredUi{
			OutputColor: cli.UiColorNone,
			InfoColor:   cli.UiColorNone,
			ErrorColor:  cli.UiColorRed,
			WarnColor:   cli.UiColorYellow,
			Ui: &cli.BasicUi{
				Reader:      os.Stdin,
				Writer:      os.Stdout,
				ErrorWriter: os.Stderr,
			},
		},
	}

	c.Commands = map[string]cli.CommandFactory{
		"migrate": func() (cli.Command, error) {
			return &RunCommand{Meta: meta}, nil
		},
	}

	return c
}
