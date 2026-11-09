//go:build !test

package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestWebhookConfigCompatibility(t *testing.T) {
	for _, format := range []string{"yaml", "json"} {
		t.Run(format, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)
			config := `{"server":{"secret":"legacy + /?&=测试","plugin":{"webhook":[{"name":"legacy","file":"legacy.lua","env":{"token":"test-token"}}]}}}`
			if format == "yaml" {
				config = `server:
  secret: "legacy + /?&=测试"
  plugin:
    webhook:
      - name: legacy
        file: legacy.lua
        env:
          token: test-token
`
			}
			viper.SetConfigType(format)
			if err := viper.ReadConfig(strings.NewReader(config)); err != nil {
				t.Fatal(err)
			}
			if secret := viper.GetString("server.secret"); secret != "legacy + /?&=测试" {
				t.Fatal("configured secret changed")
			}
			hooks := getWebhooks()
			if len(hooks) != 1 || hooks[0]["name"] != "legacy" || hooks[0]["file"] != "legacy.lua" {
				t.Fatalf("webhook config lost: %#v", hooks)
			}
			env, ok := hooks[0]["env"].(map[string]interface{})
			if !ok || env["token"] != "test-token" {
				t.Fatalf("webhook environment changed: %#v", hooks[0]["env"])
			}
		})
	}
}
