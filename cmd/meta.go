package cmd

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
)

// Meta stores values and methods needed by all commands
type Meta struct {
	UI cli.Ui
}

func (m *Meta) flagSet(name string) *flag.FlagSet {
	return flag.NewFlagSet(name, flag.ContinueOnError)
}

func (m *Meta) flagUsage(flags *flag.FlagSet) string {
	var out string

	flags.VisitAll(func(f *flag.Flag) {
		s := fmt.Sprintf("  -%s", f.Name) // Two spaces before -; see next two comments.
		name, usage := flag.UnquoteUsage(f)
		if len(name) > 0 {
			s += " " + name
		}
		// Boolean flags of one ASCII letter are so common we
		// treat them specially, putting their usage on the same line.
		if len(s) <= 4 { // space, space, '-', 'x'.
			s += "\t"
		} else {
			// Four spaces before the tab triggers good alignment
			// for both 4- and 8-space tab stops.
			s += "\n    \t"
		}
		s += strings.ReplaceAll(usage, "\n", "\n    \t")

		out += s + "\n"
	})

	return out
}
