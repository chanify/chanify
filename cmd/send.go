//+build !test

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"

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
			sound := viper.GetString("client.sound")
			priority := viper.GetInt("client.priority")
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
			data := url.Values{
				"text":  {text},
				"token": {token},
			}
			if len(sound) > 0 {
				data.Add("sound", sound)
			}
			if priority > 0 {
				data.Add("sound", strconv.Itoa(priority))
			}
			resp, err := http.PostForm(viper.GetString("client.endpoint")+"/v1/sender", data)
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusOK {
				x, _ := ioutil.ReadAll(resp.Body)
				defer resp.Body.Close()
				return fmt.Errorf("Send failed: %d, %s", resp.StatusCode, string(x))
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
	sendCmd.Flags().String("sound", "1", "Message sound.")
	sendCmd.Flags().String("text", "", "Text message content.")
	sendCmd.Flags().Int("priority", 0, "Message priority.")
	viper.BindPFlag("client.token", sendCmd.Flags().Lookup("token"))       // nolint: errcheck
	viper.BindPFlag("client.sound", sendCmd.Flags().Lookup("sound"))       // nolint: errcheck
	viper.BindPFlag("client.priority", sendCmd.Flags().Lookup("priority")) // nolint: errcheck
	viper.BindPFlag("client.endpoint", sendCmd.Flags().Lookup("endpoint")) // nolint: errcheck
}
