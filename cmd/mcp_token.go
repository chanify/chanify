//go:build !test

package cmd

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/chanify/chanify/logic"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	parent := &cobra.Command{Use: "mcp-token", Short: "Manage independent MCP credentials using local node database access"}
	rootCmd.AddCommand(parent)
	parent.PersistentFlags().String("user", "", "Account user ID")
	parent.PersistentFlags().String("dburl", "", "Node database URL (defaults to server.dburl or server.datapath)")
	for _, action := range []string{"create", "list", "rename", "revoke"} {
		command := &cobra.Command{Use: action, Short: action + " MCP credentials", RunE: func(cmd *cobra.Command, args []string) error {
			uid, _ := cmd.Flags().GetString("user")
			if uid == "" {
				return errors.New("--user is required")
			}
			dsn, _ := cmd.Flags().GetString("dburl")
			if dsn == "" {
				dsn = viper.GetString("server.dburl")
			}
			l, err := logic.NewLogic(&logic.Options{DBUrl: dsn, DataPath: getExpandPath("server.datapath"), Secret: viper.GetString("server.secret")})
			if err != nil {
				return err
			}
			defer l.Close()
			var result interface{}
			switch cmd.Name() {
			case "create":
				name, _ := cmd.Flags().GetString("name")
				ttl, _ := cmd.Flags().GetDuration("expires-in")
				if ttl <= 0 || ttl > 5*366*24*time.Hour {
					return errors.New("invalid --expires-in")
				}
				record, raw, err := l.CreateMCPToken(uid, name, time.Now().Add(ttl).Unix())
				if err != nil {
					return err
				}
				result = map[string]interface{}{"credential": record, "token": raw}
			case "list":
				result, err = l.ListMCPTokens(uid)
			default:
				id, _ := cmd.Flags().GetString("id")
				name, _ := cmd.Flags().GetString("name")
				err = l.UpdateMCPToken(uid, id, name, cmd.Name() == "revoke")
				result = map[string]bool{"success": err == nil}
			}
			if err != nil {
				return err
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
		}}
		command.Flags().String("name", "", "Credential name")
		command.Flags().String("id", "", "Credential ID")
		if action == "create" {
			command.Flags().Duration("expires-in", 90*24*time.Hour, "Credential lifetime")
		}
		parent.AddCommand(command)
	}
}
