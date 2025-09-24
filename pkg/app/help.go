package app

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	flagHelp          = "help"
	flagHelpShorthand = "h"
)

// helpCommand creates a help command for the application with the specified name.
func helpCommand(name string) *cobra.Command {
	hc := &cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command.",
		Long: `Help provides help for any command in the application.
Simply type ` + name + ` help [path to command] for full details.`,
		Run: func(c *cobra.Command, args []string) {
			cmd, _, err := c.Root().Find(args)
			if cmd == nil || err != nil {
				c.Printf("Unknown help topic %#q\n", args)
				_ = c.Root().Usage()
			} else {
				cmd.InitDefaultHelpCmd()
				_ = cmd.Help()
			}
		},
	}
	return hc
}

// addHelpCommandFlag adds flags for a specific command of application to the specified FlagSet object.
func addHelpCommandFlag(usage string, fs *pflag.FlagSet) {
	fs.BoolP(
		flagHelp,
		flagHelpShorthand,
		false,
		fmt.Sprintf("Help for the %s command", color.GreenString(strings.Split(usage, " ")[0])),
	)
}
