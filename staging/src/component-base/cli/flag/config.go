package flag

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const configFlagName = "config"

var cfgFile string

func init() {
	pflag.StringVarP(&cfgFile, configFlagName, "c", "", "Path to the configuration file, support only YAML.")
}

// AddConfigFlag adds a config flag to the specified FlagSet object and binds it to viper.
// basename is the binary name of the application without space, like "apiserver".
// It will be used as the prefix of environment variables and the name of configuration file by viper.
func AddConfigFlag(basename string, fs *pflag.FlagSet) {
	// register the config flag to the specified FlagSet object
	fs.AddFlag(pflag.Lookup(configFlagName))

	viper.AutomaticEnv()
	viper.SetEnvPrefix(strings.ReplaceAll(strings.ToUpper(basename), "-", "_"))
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	cobra.OnInitialize(func() {
		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else {
			viper.AddConfigPath(".")
			viper.AddConfigPath("$HOME/.siam")
			viper.AddConfigPath("/etc/siam")

			// if the binary name contains "-", use the part after "-" as the config file name
			// like "siam-apiserver" will use "apiserver" as the config file name
			filename := basename
			if strings.Contains(basename, "-") {
				filename = strings.Split(basename, "-")[1]
			}
			viper.SetConfigName(filename)
		}

		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to read configuration file(%s): %v\n", cfgFile, err)
			os.Exit(1)
		}
	})
}
