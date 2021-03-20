//+build !test

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	sendCmd := &cobra.Command{
		Use:   "send",
		Short: "Send message cli",
		Long:  "A send message command line tool.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			token := viper.GetString("client.token")
			if len(token) <= 0 {
				return fmt.Errorf("Send token not found.")
			}
			text, err := cmd.Flags().GetString("text")
			if err != nil {
				return err
			}
			if len(text) > 1 {
				if text[0] == '@' {
					if text[1] == '-' {
						in, err := ioutil.ReadAll(os.Stdin)
						if err != nil {
							return err
						}
						text = string(in)
					} else {
						in, err := ioutil.ReadFile(text[1:])
						if err != nil {
							return err
						}
						text = string(in)
					}
				}
			}
			if len(text) <= 0 {
				return fmt.Errorf("No message content.")
			}
			resp, err := http.Post(viper.GetString("client.endpoint")+"/v1/sender/"+token, "text/plain", strings.NewReader(text))
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("Send failed: %d", resp.StatusCode)
			}
			defer resp.Body.Close()
			var res struct {
				Uid string `json:"request-uid,omitempty"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
				return err
			}
			fmt.Println("request-uid:", res.Uid)
			return nil
		},
	}
	rootCmd.AddCommand(sendCmd)
	sendCmd.Flags().String("endpoint", "https://api.chanify.net", "Node server endpoint.")
	sendCmd.Flags().String("token", "", "Send token.")
	sendCmd.Flags().String("text", "", "Text message content.")
	viper.BindPFlag("client.token", sendCmd.Flags().Lookup("token"))       // nolint: errcheck
	viper.BindPFlag("client.endpoint", sendCmd.Flags().Lookup("endpoint")) // nolint: errcheck
}
