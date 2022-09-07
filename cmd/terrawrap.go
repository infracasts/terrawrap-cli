package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile, cfgPath string

	rootCmd = &cobra.Command{
		Use:   "terrawrap",
		Short: "A generator of terraform modules from terraform resources",
		Long: `Terrawrap is a CLI tool that extracts useful information from
hashicorp-maintained terraform providers' resources, relying on their
standardized documentation, as this information isn't available via the
Terraform Registry API.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.terrawrap/config.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
		cfgPath = path.Dir(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		cfgPath = path.Join(home, ".terrawrap")
		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(cfgPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

