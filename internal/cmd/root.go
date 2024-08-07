package cmd

import (
	"fmt"
	"strings"

	"github.com/aprimetechnology/derisk-sql/internal/cmd/check"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DefaultConfigFileName  = "settings"
	DefaultConfigFilePath  = "."
	DefaultConfigEnvPrefix = "DERISK_SQL"
)

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:          "derisk-sql",
	SilenceUsage: true,
	Short:        "CLI for linting SQL migration files to prevent easy-to-miss mistakes",
	// PersistentPreRunE will be inherited by child subcommands
	// initializeConfig here sets up viper for using config files
	// each relevant child subcommand's RunE function should then call viper.Unmarshal(&someSubcommandConfigStruct)
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig(cmd)
	},
}

func initializeConfig(cmd *cobra.Command) error {
	viper.SetConfigName(DefaultConfigFileName)
	viper.AddConfigPath(DefaultConfigFilePath)
	configMessage := fmt.Sprintf(
		"config file starting with %q in directory %q",
		DefaultConfigFileName,
		DefaultConfigFilePath,
	)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf(
				"Failure reading %s: %w",
				configMessage,
				err,
			)
		}
		fmt.Printf("%s not found! Moving on.\n", configMessage)
	} else {
		fmt.Printf("Using %s\n", configMessage)
	}

	// also check environment variables, prefixed with envPrefix
	viper.SetEnvPrefix(DefaultConfigEnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	return nil
}

func init() {
	RootCmd.AddCommand(check.CheckCmd)
}
