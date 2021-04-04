//+build !test

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
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
			link, err := cmd.Flags().GetString("link")
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
			var image []byte
			imagePath, err := cmd.Flags().GetString("image")
			if err != nil {
				return err
			}
			if len(imagePath) > 0 {
				if imagePath[0] == '-' {
					in, err := ioutil.ReadAll(os.Stdin)
					if err != nil {
						return err
					}
					image = in
				} else {
					in, err := ioutil.ReadFile(imagePath)
					if err != nil {
						return err
					}
					image = in
				}
			}
			if len(text) <= 0 && len(image) <= 0 && len(link) <= 0 {
				return fmt.Errorf("No message content.")
			}
			var data bytes.Buffer
			w := multipart.NewWriter(&data)
			if len(token) <= 0 {
				return fmt.Errorf("No token.")
			} else {
				fw, _ := w.CreateFormField("token")
				fw.Write([]byte(token)) // nolint: errcheck
			}
			if len(text) > 0 {
				fw, _ := w.CreateFormField("text")
				fw.Write([]byte(text)) // nolint: errcheck
			}
			if len(link) > 0 {
				fw, _ := w.CreateFormField("link")
				fw.Write([]byte(link)) // nolint: errcheck
			}
			if len(image) > 0 {
				fw, _ := w.CreateFormFile("image", "image")
				fw.Write(image) // nolint: errcheck
			}
			if title, err := cmd.Flags().GetString("title"); err == nil && len(title) > 0 {
				fw, _ := w.CreateFormField("title")
				fw.Write([]byte(title)) // nolint: errcheck
			}
			if len(sound) > 0 {
				fw, _ := w.CreateFormField("sound")
				fw.Write([]byte(sound)) // nolint: errcheck
			}
			if priority > 0 {
				fw, _ := w.CreateFormField("priority")
				fw.Write([]byte(strconv.Itoa(priority))) // nolint: errcheck
			}
			w.Close()
			req, err := http.NewRequest("POST", viper.GetString("client.endpoint")+"/v1/sender", &data)
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", w.FormDataContentType())
			resp, err := http.DefaultClient.Do(req)
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
	sendCmd.Flags().String("link", "", "Link message content.")
	sendCmd.Flags().String("image", "", "Image file path.")
	sendCmd.Flags().String("title", "", "Message title.")
	sendCmd.Flags().Int("priority", 0, "Message priority.")
	viper.BindPFlag("client.token", sendCmd.Flags().Lookup("token"))       // nolint: errcheck
	viper.BindPFlag("client.sound", sendCmd.Flags().Lookup("sound"))       // nolint: errcheck
	viper.BindPFlag("client.priority", sendCmd.Flags().Lookup("priority")) // nolint: errcheck
	viper.BindPFlag("client.endpoint", sendCmd.Flags().Lookup("endpoint")) // nolint: errcheck
}
