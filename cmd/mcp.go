//go:build !test

package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/chanify/chanify/mcpnotify"
	"github.com/chanify/chanify/sender"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	command := &cobra.Command{Use: "mcp", Short: "Serve MCP over stdio using the existing node sending API", Args: cobra.NoArgs, SilenceErrors: true, SilenceUsage: true, RunE: runMCPCmd}
	command.Flags().String("endpoint", "", "Node endpoint; defaults to client.endpoint or https://api.chanify.net.")
	rootCmd.AddCommand(command)
}
func runMCPCmd(command *cobra.Command, _ []string) error {
	endpoint := viper.GetString("client.endpoint")
	if command.Flags().Changed("endpoint") {
		endpoint, _ = command.Flags().GetString("endpoint")
	}
	if endpoint == "" {
		endpoint = sender.DefaultEndpoint
	}
	ctx, cancel := signal.NotifyContext(command.Context(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	client, err := mcpnotify.NewClient(ctx, endpoint, viper.GetString("client.token"), Version, nil)
	if err != nil {
		return err
	}
	err = client.Server(Version).Run(ctx, &mcp.StdioTransport{MaxLineLength: 6 << 20})
	if ctx.Err() == context.Canceled {
		return nil
	}
	return err
}
