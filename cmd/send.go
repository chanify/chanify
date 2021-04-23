//+build !test

package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	sendCmd := &cobra.Command{
		Use:   "send",
		Short: "Send message cli",
		Long:  "A send message command line tool.",
		RunE:  runSendCmd,
	}
	rootCmd.AddCommand(sendCmd)
	sendCmd.Flags().String("endpoint", "https://api.chanify.net", "Node server endpoint.")
	sendCmd.Flags().String("token", "", "Send token.")
	sendCmd.Flags().String("sound", "", "Message sound.")
	sendCmd.Flags().String("text", "", "Text message content.")
	sendCmd.Flags().String("link", "", "Link message content.")
	sendCmd.Flags().String("image", "", "Image file path.")
	sendCmd.Flags().String("file", "", "File path.")
	sendCmd.Flags().String("title", "", "Message title.")
	sendCmd.Flags().String("copy", "", "Copy test for text message.")
	sendCmd.Flags().String("autocopy", "", "Auto copy text for text message.")
	sendCmd.Flags().Int("priority", 0, "Message priority.")
	viper.BindPFlag("client.token", sendCmd.Flags().Lookup("token"))       // nolint: errcheck
	viper.BindPFlag("client.sound", sendCmd.Flags().Lookup("sound"))       // nolint: errcheck
	viper.BindPFlag("client.autocopy", sendCmd.Flags().Lookup("autocopy")) // nolint: errcheck
	viper.BindPFlag("client.priority", sendCmd.Flags().Lookup("priority")) // nolint: errcheck
	viper.BindPFlag("client.endpoint", sendCmd.Flags().Lookup("endpoint")) // nolint: errcheck
}

func readFile(path string) ([]byte, error) {
	var data []byte
	if len(path) > 0 {
		if path[0] == '-' {
			in, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return nil, err
			}
			data = in
		} else {
			in, err := ioutil.ReadFile(path)
			if err != nil {
				return nil, err
			}
			data = in
		}
	}
	return data, nil
}

func runSendCmd(cmd *cobra.Command, args []string) error {
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	sound := viper.GetString("client.sound")
	autocopy := viper.GetString("client.autocopy")
	priority := viper.GetInt("client.priority")
	token := viper.GetString("client.token")
	if len(token) <= 0 {
		return errors.New("send token not found")
	}
	text, err := cmd.Flags().GetString("text")
	if err != nil {
		return err
	}
	link, err := cmd.Flags().GetString("link")
	if err != nil {
		return err
	}
	if len(text) > 1 && text[0] == '@' {
		txt, err := readFile(text[1:])
		if err != nil {
			return err
		}
		text = string(txt)
	}
	imagePath, err := cmd.Flags().GetString("image")
	if err != nil {
		return err
	}
	image, err := readFile(imagePath)
	if err != nil {
		return err
	}
	var file []byte
	var filename string
	filePath, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}
	if len(filePath) > 0 {
		in, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}
		file = in
		filename = filepath.Base(filePath)
	}
	if len(text) <= 0 && len(image) <= 0 && len(link) <= 0 && len(file) <= 0 {
		return errors.New("no message content")
	}
	var data bytes.Buffer
	w := multipart.NewWriter(&data)
	if len(token) <= 0 {
		return errors.New("no token")
	}
	fw, _ := w.CreateFormField("token")
	fw.Write([]byte(token)) // nolint: errcheck
	if len(text) > 0 {
		fw, _ = w.CreateFormField("text")
		fw.Write([]byte(text)) // nolint: errcheck
	}
	if len(link) > 0 {
		fw, _ = w.CreateFormField("link")
		fw.Write([]byte(link)) // nolint: errcheck
	}
	if len(image) > 0 {
		fw, _ = w.CreateFormFile("image", "image")
		fw.Write(image) // nolint: errcheck
	}
	if len(file) > 0 && len(filename) > 0 {
		fw, _ = w.CreateFormFile("file", filename)
		fw.Write(file) // nolint: errcheck
	}
	if title, err := cmd.Flags().GetString("title"); err == nil && len(title) > 0 {
		fw, _ = w.CreateFormField("title")
		fw.Write([]byte(title)) // nolint: errcheck
	}
	if copytext, err := cmd.Flags().GetString("copy"); err == nil && len(copytext) > 0 {
		fw, _ = w.CreateFormField("copy")
		fw.Write([]byte(copytext)) // nolint: errcheck
	}
	if len(autocopy) > 0 {
		fw, _ = w.CreateFormField("autocopy")
		fw.Write([]byte(autocopy)) // nolint: errcheck
	}
	if len(sound) > 0 {
		fw, _ = w.CreateFormField("sound")
		fw.Write([]byte(sound)) // nolint: errcheck
	}
	if priority > 0 {
		fw, _ = w.CreateFormField("priority")
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
		return fmt.Errorf("send failed: %d, %s", resp.StatusCode, string(x))
	}
	defer resp.Body.Close()
	var res struct {
		UID string `json:"request-uid,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}
	fmt.Println("request-uid:", res.UID)
	return nil
}
