package app

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Command is a sub command structure of a cli application.
// It is recommended that a command be created with the app.NewCommand() function.
type Command struct {
	usage   string
	desc    string
	options CliOptions
	command []*Command
	// runFunc is the command's startup callback function.
	// If runFunc is not nil, it will be called in Run of cobra.Command.
	runFunc RunCommandFunc
}

// CommandOption defines optional parameters for initializing the command structure.
type CommandOption func(*Command)

func WithCommandOptions(opts CliOptions) CommandOption {
	return func(c *Command) {
		c.options = opts
	}
}

// RunCommandFunc defines the application's command startup callback function.
type RunCommandFunc func(args []string) error

func WithCommandRunFunc(runFunc RunCommandFunc) CommandOption {
	return func(c *Command) {
		c.runFunc = runFunc
	}
}

// NewCommand creates a new sub command instance based on the given command name and other options.
func NewCommand(usage, desc string, opts ...CommandOption) *Command {
	c := &Command{
		usage: usage,
		desc:  desc,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// AddCommand adds one or more sub commands to the current command.
func (c *Command) AddCommand(cmd ...*Command) {
	c.command = append(c.command, cmd...)
}

// cobraCommand converts the Command structure to a cobra.Command structure.
func (c *Command) cobraCommand() *cobra.Command {
	cc := &cobra.Command{
		Use:   c.usage,
		Short: c.desc,
	}
	cc.SetOut(os.Stdout)
	cc.Flags().SortFlags = false
	// c has sub commands
	if len(c.command) > 0 {
		for _, cmd := range c.command {
			cc.AddCommand(cmd.cobraCommand())
		}
	}
	if c.runFunc != nil {
		cc.Run = c.runCommand
	}
	if c.options != nil {
		for _, f := range c.options.Flags().FlagSets {
			cc.Flags().AddFlagSet(f)
		}
	}
	addHelpCommandFlag(c.usage, cc.Flags())

	return cc
}

// runCommand is the callback function for executing the command.
func (c *Command) runCommand(cmd *cobra.Command, args []string) {
	if c.runFunc != nil {
		if err := c.runFunc(args); err != nil {
			fmt.Printf("%v %v\n", color.RedString("Error:"), err)
			os.Exit(1)
		}
	}
}
