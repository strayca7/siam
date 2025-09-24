package app

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"k8s.io/component-base/term"

	"github.com/strayca7/siam/internal/pkg/config"
	"github.com/strayca7/siam/pkg/logger"
	"github.com/strayca7/siam/pkg/serrors"
	cliflag "github.com/strayca7/siam/staging/src/component-base/cli/flag"
	// cliflag "k8s.io/component-base/cli/flag"
)

// App is the main structure of a cli application.
// It is recommended that an app be created with the app.NewApp() function.
type App struct {
	// name is the full name of the application, like "API Server".
	name string
	// basename is the binary name of the application without space, like "apiserver".
	// basename must use the constant in staging/src/api/name/v1/name.go
	basename    string
	description string
	options     CliOptions
	// runFunc is the application's startup callback function.
	// If runFunc is not nil, it will be called in RunE of cobra.Command by runCommand method.
	runFunc   RunFunc
	silence   bool
	noVersion bool
	noConfig  bool
	// App's commands is a list of sub commands of the application. Its sub commands may have their own sub commands.
	commands []*Command
	args     cobra.PositionalArgs
	// cmd is the root command of the application.
	// All of other fields are used to build cmd.
	cmd *cobra.Command
}

// Option defines optional parameters for initializing the application structure.
type Option func(*App)

// WithOptions to open the application's function to read from the command line
// or read parameters from the configuration file.
func WithOptions(opts CliOptions) Option {
	return func(a *App) {
		a.options = opts
	}
}

// RunFunc defines the application's startup callback function.
type RunFunc func(basename string) error

// WithRunFunc is used to set the application startup callback function option.
func WithRunFunc(runFunc RunFunc) Option {
	return func(a *App) {
		a.runFunc = runFunc
	}
}

// WithDescription is used to set the description of the application.
func WithDescription(desc string) Option {
	return func(a *App) {
		a.description = desc
	}
}

// WithSilence sets the application to silent mode, in which the program startup
// information, configuration information, and version information are not
// printed in the console.
func WithSilence() Option {
	return func(a *App) {
		a.silence = true
	}
}

// WithNoVersion set the application does not provide version flag.
func WithNoVersion() Option {
	return func(a *App) {
		a.noVersion = true
	}
}

// WithNoConfig set the application does not provide config flag.
func WithNoConfig() Option {
	return func(a *App) {
		a.noConfig = true
	}
}

// WithDefaultValidArgs set default validation function to valid non-flag arguments.
func WithDefaultValidArgs() Option {
	return func(a *App) {
		a.args = func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		}
	}
}

// NewApp creates a new application instance based on the given application name, binary name, and other options.
func NewApp(name string, basename string, opts ...Option) *App {
	a := &App{
		name:     name,
		basename: basename,
	}
	for _, o := range opts {
		o(a)
	}

	a.buildCommand()
	return a
}

// buildCommand builds the cobra.Command instance and assigns it to cmd based on the App structure.
func (a *App) buildCommand() {
	cmd := &cobra.Command{
		Use:           FormatBasename(a.basename),
		Short:         a.name,
		Long:          a.description,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          a.args,
	}
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.Flags().SortFlags = true
	cliflag.InitFlags(cmd.Flags())

	if len(a.commands) > 0 {
		// convert Command to cobra.Command and add to cmd
		for _, c := range a.commands {
			cmd.AddCommand(c.cobraCommand())
		}
		cmd.SetHelpCommand(helpCommand(FormatBasename(a.basename)))
	}
	if a.runFunc != nil {
		cmd.RunE = a.runCommand
	}

	var namedFlagSets cliflag.NamedFlagSets
	// add flags from options to cobra.Command
	if a.options != nil {
		namedFlagSets = a.options.Flags()
		fs := cmd.Flags()
		for _, f := range namedFlagSets.FlagSets {
			fs.AddFlagSet(f)
		}
	}
	if !a.noVersion {
		cliflag.AddVersionFlag(namedFlagSets.FlagSet("global"))
	}
	if !a.noConfig {
		cliflag.AddConfigFlag(cmd.Name(), namedFlagSets.FlagSet("global"))
	}

	// now this AddFlagSet function here only adds a "help" flag to the command,
	// you can see this function's implementation in Kubernetes, k8s.io/component-base/cli/globalflag/globalflags.go
	cliflag.AddGlobalFlags(cmd.Name(), namedFlagSets.FlagSet("global"))

	cmd.Flags().AddFlagSet(namedFlagSets.FlagSet("global"))

	col, _, _ := term.TerminalSize(cmd.OutOrStdout())
	// format and set usage and help function
	cliflag.SetUsageAndHelpFunc(cmd, namedFlagSets, col)
	a.cmd = cmd
}

// runCommand is the callback function for executing the command.
func (a *App) runCommand(cmd *cobra.Command, args []string) error {
	printWorkingDir()
	cliflag.PrintFlags(cmd.Flags())
	if !a.noVersion {
		// display application version information
		// TODO: use a custom version method
		logger.L().Info(fmt.Sprintf("%s version: %s", a.name, cmd.Version))
	}
	if !a.noConfig {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		if err := viper.Unmarshal(a.options); err != nil {
			return err
		}
	}

	if !a.silence {
		logger.L().Info("Application is starting...", zap.String("name", a.name))
		if !a.noVersion {
			// TODO: use a custom version method
			logger.L().Info(fmt.Sprintf("%s version: %s", a.name, cmd.Version))
		}
		if !a.noConfig {
			logger.L().Info("Current configuration", zap.String("config file", viper.ConfigFileUsed()))
		}
		if a.options != nil {
			if err := a.applyOptionRules(); err != nil {
				return err
			}
		}
		// runFunc is the core of the runCommand function
		if a.runFunc != nil {
			return a.runFunc(a.basename)
		}
	}
	return nil
}

func (a *App) applyOptionRules() error {
	if completeableOptions, ok := a.options.(CompleteableOptions); ok {
		if err := completeableOptions.Complete(); err != nil {
			return err
		}
	}
	if errs := a.options.Validate(); errs != nil {
		return serrors.NewAggregate(errs)
	}
	if printableOptions, ok := a.options.(PrintableOptions); ok && !a.silence {
		logger.L().Info("Current configuration", zap.String("options", printableOptions.String()))
	}
	return nil
}

// Command returns the root cobra.Command instance of the application.
func (a *App) Command() *cobra.Command {
	return a.cmd
}

// Run launches the application and returns an exit code.
func Run(a *App) int {
	// TODO: logger should be initialized after parsing command line args successfully and before main logic execution
	logger.Init(context.Background(), config.GlobalOptions.Log, logger.WithName(a.basename))
	defer logger.L().Sync()
	if err := a.cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

// FormatBasename formats the binary name of the application according to the operating system.
// It will lowercase the name. If the OS is Windows, it will also remove the ".exe" suffix.
func FormatBasename(basename string) string {
	basename = strings.ToLower(basename)
	if runtime.GOOS == "windows" {
		basename = strings.TrimSuffix(basename, ".exe")
	}
	return basename
}

// printWorkingDir prints the current working directory to the log in debug level.
func printWorkingDir() {
	wd, _ := os.Getwd()
	logger.L().Debug("WorkingDir", zap.String("dir", wd))
}
