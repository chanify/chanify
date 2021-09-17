//go:build !test
// +build !test

package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var rootCmd = &cobra.Command{
	Use:   "chanify",
	Short: "Chanify CLI",
	Long:  `Chanify command line tools`,
}

// Execute command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
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
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".chanify")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil && viper.GetBool("config.verbose") {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
