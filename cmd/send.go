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
	sendCmd.Flags().String("audio", "", "Audio file path.")
	sendCmd.Flags().String("file", "", "File path.")
	sendCmd.Flags().String("title", "", "Message title.")
	sendCmd.Flags().String("copy", "", "Copy test for text message.")
	sendCmd.Flags().String("autocopy", "", "Auto copy text for text message.")
	sendCmd.Flags().StringArray("action", []string{}, "Action item for action message.")
	sendCmd.Flags().Int("priority", 0, "Message priority.")
	viper.BindPFlag("client.token", sendCmd.Flags().Lookup("token"))       // nolint: errcheck
	viper.BindPFlag("client.sound", sendCmd.Flags().Lookup("sound"))       // nolint: errcheck
	viper.BindPFlag("client.autocopy", sendCmd.Flags().Lookup("autocopy")) // nolint: errcheck
	viper.BindPFlag("client.priority", sendCmd.Flags().Lookup("priority")) // nolint: errcheck
	viper.BindPFlag("client.endpoint", sendCmd.Flags().Lookup("endpoint")) // nolint: errcheck
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
	text, _ := cmd.Flags().GetString("text")
	link, _ := cmd.Flags().GetString("link")
	if len(text) > 1 && text[0] == '@' {
		txt, err := readInputFile(text[1:])
		if err != nil {
			return err
		}
		text = string(txt)
	}
	imagePath, _ := cmd.Flags().GetString("image")
	image, _ := readInputFile(imagePath)
	audioPath, _ := cmd.Flags().GetString("audio")
	audio, _ := readInputFile(audioPath)
	filePath, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}
	file, filename, err := readFile(filePath)
	if err != nil {
		return err
	}
	if len(text) <= 0 && len(image) <= 0 && len(link) <= 0 && len(file) <= 0 && len(audio) <= 0 {
		return errors.New("no message content")
	}
	var data bytes.Buffer
	w := multipart.NewWriter(&data)
	if len(token) <= 0 {
		return errors.New("no token")
	}
	setFieldValue(w, "token", []byte(token))
	setFieldValue(w, "text", []byte(text))
	setFieldValue(w, "link", []byte(link))
	setFieldFile(w, "image", "image", image)
	setFieldFile(w, "audio", "audio", audio)
	setFieldFile(w, "file", filename, file)
	title, _ := cmd.Flags().GetString("title")
	setFieldValue(w, "title", []byte(title))
	copytext, _ := cmd.Flags().GetString("copy")
	setFieldValue(w, "copy", []byte(copytext))
	setFieldValue(w, "autocopy", []byte(autocopy))
	setFieldValue(w, "sound", []byte(sound))
	setFieldValueInt(w, "priority", priority)
	actions, _ := cmd.Flags().GetStringArray("action")
	setFieldValues(w, "action", actions)
	w.Close()
	return sendMessage(&data, w.FormDataContentType())
}

func sendMessage(in *bytes.Buffer, content string) error {
	req, err := http.NewRequest("POST", viper.GetString("client.endpoint")+"/v1/sender", in)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", content)
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

func readFile(filePath string) ([]byte, string, error) {
	if len(filePath) > 0 {
		in, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, "", err
		}
		return in, filepath.Base(filePath), nil
	}
	return nil, "", nil
}

func readInputFile(path string) ([]byte, error) {
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

func setFieldValue(w *multipart.Writer, name string, value []byte) {
	if len(value) > 0 {
		fw, _ := w.CreateFormField(name)
		fw.Write(value) // nolint: errcheck
	}
}

func setFieldValues(w *multipart.Writer, name string, value []string) {
	for _, v := range value {
		if len(v) > 0 {
			fw, _ := w.CreateFormField(name)
			fw.Write([]byte(v)) // nolint: errcheck
		}
	}
}

func setFieldValueInt(w *multipart.Writer, name string, value int) {
	if value > 0 {
		fw, _ := w.CreateFormField(name)
		fw.Write([]byte(strconv.Itoa(value))) // nolint: errcheck
	}
}

func setFieldFile(w *multipart.Writer, name string, fname string, value []byte) {
	if len(value) > 0 && len(fname) > 0 {
		fw, _ := w.CreateFormFile(name, fname)
		fw.Write(value) // nolint: errcheck
	}
}
