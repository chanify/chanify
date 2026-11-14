//go:build !test
// +build !test

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var configLoadErr error
var executionStarted time.Time
var rootCmd = &cobra.Command{
	Use:   "chanify",
	Short: "Chanify CLI",
	Long:  `Chanify command line tools`,
}

// Execute command
func Execute() {
	executionStarted = time.Now()
	command, err := rootCmd.ExecuteC()
	if err != nil {
		if isAgentSend(command) {
			os.Exit(writeAgentError(command.OutOrStdout(), err))
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chanify.yml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "Make the operation more talkative")
	viper.BindPFlag("config.verbose", rootCmd.PersistentFlags().Lookup("verbose")) // nolint: errcheck
}

func initConfig() {
	configLoadErr = nil
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			configLoadErr = err
			fmt.Fprintln(os.Stderr, "Cannot locate the configuration directory.")
			return
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".chanify")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok || cfgFile != "" {
			configLoadErr = err
			fmt.Fprintln(os.Stderr, "Cannot read config file:", viper.ConfigFileUsed())
		}
	} else if viper.GetBool("config.verbose") {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
