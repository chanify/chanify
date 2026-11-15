//go:build !test

package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chanify/chanify/agentcli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func commandForTest(t *testing.T, path ...string) *cobra.Command {
	t.Helper()
	c, _, err := rootCmd.Find(path)
	if err != nil {
		t.Fatal(err)
	}
	c.Flags().AddFlagSet(c.InheritedFlags())
	originalContext := c.Context()
	c.SetContext(context.Background())
	t.Cleanup(func() { c.SetContext(originalContext) })
	// Restore flag values and Changed bits, including inherited flag objects.
	c.Flags().VisitAll(func(f *pflag.Flag) {
		value, changed := f.Value.String(), f.Changed
		t.Cleanup(func() { _ = f.Value.Set(value); f.Changed = changed })
	})
	args := append([]string(nil), c.Flags().Args()...)
	t.Cleanup(func() { _ = c.Flags().Parse(args) })
	out, errOut := c.OutOrStdout(), c.ErrOrStderr()
	t.Cleanup(func() { c.SetOut(out); c.SetErr(errOut) })
	return c
}
func resetConfig(t *testing.T) { t.Helper(); viper.Reset(); t.Cleanup(viper.Reset) }
func setFlag(t *testing.T, c *cobra.Command, name, value string) {
	t.Helper()
	if err := c.Flags().Set(name, value); err != nil {
		t.Fatal(err)
	}
}

func TestServerOptions(t *testing.T) {
	resetConfig(t)
	for _, tc := range []struct{ name, hostname, host, want string }{{"node", "legacy", "host", "node"}, {"", "legacy", "host", "legacy"}, {"", "", "host", "host"}} {
		viper.Set("server.name", tc.name)
		viper.Set("server.hostname", tc.hostname)
		viper.Set("server.host", tc.host)
		if got := getName(); got != tc.want {
			t.Fatal(got)
		}
	}
	viper.Set("server.hostname", "")
	viper.Set("server.host", "node.example")
	for _, tc := range []struct {
		port int
		want string
	}{{80, "http://node.example"}, {443, "https://node.example"}, {8080, "http://node.example:8080"}} {
		viper.Set("server.port", tc.port)
		if got := getEndpoint(); got != tc.want {
			t.Fatal(got)
		}
	}
	viper.Set("server.endpoint", "https://public.example/push")
	if getEndpoint() != "https://public.example/push" {
		t.Fatal(getEndpoint())
	}
	viper.Set("server.endpoint", "")
	viper.Set("server.host", "")
	if getEndpoint() != "" {
		t.Fatal(getEndpoint())
	}
	viper.Set("server.hostname", "legacy.example")
	viper.Set("server.port", 80)
	if getEndpoint() != "http://legacy.example" {
		t.Fatal(getEndpoint())
	}
	for _, tc := range []struct {
		in   string
		want time.Duration
	}{{"2m", 2 * time.Minute}, {"bad", 10 * time.Second}, {"0", 0}} {
		if got := parseDuration(tc.in); got != tc.want {
			t.Fatal(got)
		}
	}
	viper.Set("path", "")
	if getExpandPath("path") != "" {
		t.Fatal("empty path changed")
	}
	viper.Set("path", t.TempDir())
	if getExpandPath("path") != viper.GetString("path") {
		t.Fatal("absolute path changed")
	}
	viper.Set("path", "~other/unsupported")
	if getExpandPath("path") != "~other/unsupported" {
		t.Fatal("unsupported path changed")
	}
}

func TestRegisterWhitelist(t *testing.T) {
	resetConfig(t)
	c := &cobra.Command{}
	c.Flags().String("whitelist", "cli-one,cli-two", "")
	viper.Set("server.register.enable", true)
	if enabled, list := getUserWhitlist(c); !enabled || list != nil {
		t.Fatal(enabled, list)
	}
	viper.Set("server.register.enable", false)
	viper.Set("server.register.whitelist", []string{"config-user"})
	enabled, list := getUserWhitlist(c)
	if enabled || strings.Join(list, ",") != "cli-one,cli-two,config-user" {
		t.Fatal(enabled, list)
	}
	if _, list := getUserWhitlist(&cobra.Command{}); len(list) != 1 {
		t.Fatal(list)
	}
}

func TestVersionOutput(t *testing.T) {
	c := commandForTest(t, "version")
	for _, tc := range []struct {
		format, want string
		fail         bool
	}{{"", "chanify version " + Version, false}, {"{{.Client.Name}}", "chanify\n", false}, {"{{json .Client.Version}}", `"` + Version + `"`, false}, {"{{", "", true}, {"{{.Missing}}", "", true}} {
		setFlag(t, c, "format", tc.format)
		var out bytes.Buffer
		c.SetOut(&out)
		err := c.RunE(c, nil)
		if (err != nil) != tc.fail {
			t.Fatal(err)
		}
		if !tc.fail && !strings.Contains(out.String(), tc.want) {
			t.Fatal(out.String())
		}
	}
	if jsonMarshal(make(chan int)) != "" || jsonMarshal("text") != `"text"` {
		t.Fatal("JSON formatting changed")
	}
}

func TestInitConfigErrorsAndVerbose(t *testing.T) {
	resetConfig(t)
	previous := cfgFile
	previousErr := configLoadErr
	t.Cleanup(func() { cfgFile = previous; configLoadErr = previousErr })
	dir := t.TempDir()
	for _, tc := range []struct {
		name, body    string
		missing, fail bool
	}{{"valid.yml", "client:\n  token: synthetic\n", false, false}, {"bad.yml", "client: [", false, true}, {"missing.yml", "", true, true}} {
		cfgFile = filepath.Join(dir, tc.name)
		if !tc.missing {
			if err := os.WriteFile(cfgFile, []byte(tc.body), 0600); err != nil {
				t.Fatal(err)
			}
		}
		viper.Set("config.verbose", true)
		initConfig()
		if (configLoadErr != nil) != tc.fail {
			t.Fatal(configLoadErr)
		}
	}
}

type outputFailure struct{}

func (outputFailure) Write([]byte) (int, error) { return 0, errors.New("output unavailable") }

func TestAgentErrorOutput(t *testing.T) {
	for _, tc := range []struct {
		err  error
		exit int
	}{
		{errors.New("argument error"), 2},
		{&agentCommandError{exit: 3, result: agentcli.Result{SchemaVersion: 1, Status: "failed", Error: &agentcli.Detail{Code: "unauthorized", Message: "rejected"}}}, 3},
		{&agentOutputError{}, 1},
	} {
		var out bytes.Buffer
		if got := writeAgentError(&out, tc.err); got != tc.exit {
			t.Fatal(got)
		}
		if tc.err.Error() == "" {
			t.Fatal("empty error description")
		}
	}
	if got := writeAgentError(outputFailure{}, errors.New("argument error")); got != 1 {
		t.Fatal(got)
	}
	if isAgentSend(nil) || isAgentSend(&cobra.Command{}) {
		t.Fatal("ordinary command classified as agent send")
	}
}

func TestExecuteVersionCommand(t *testing.T) {
	c := commandForTest(t, "version")
	setFlag(t, c, "format", "{{.Client.Name}}")
	previousStarted := executionStarted
	previousFile := cfgFile
	previousErr := configLoadErr
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		executionStarted = previousStarted
		cfgFile = previousFile
		configLoadErr = previousErr
	})
	cfgFile = filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(cfgFile, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	resetConfig(t)
	var out bytes.Buffer
	c.SetOut(&out)
	rootCmd.SetArgs([]string{"version"})
	Execute()
	if executionStarted.IsZero() || out.String() != "chanify\n" {
		t.Fatal(out.String())
	}
}

func TestDefaultConfigurationDirectory(t *testing.T) {
	resetConfig(t)
	oldFile, oldError := cfgFile, configLoadErr
	cfgFile = ""
	t.Cleanup(func() { cfgFile = oldFile; configLoadErr = oldError })
	dir := t.TempDir()
	loadConfig(func() (string, error) { return dir, nil })
	if configLoadErr != nil {
		t.Fatal("missing default configuration must be optional", configLoadErr)
	}
	want := errors.New("home unavailable")
	loadConfig(func() (string, error) { return "", want })
	if !errors.Is(configLoadErr, want) {
		t.Fatal(configLoadErr)
	}
}

func TestWebhookMixedKeyConfiguration(t *testing.T) {
	resetConfig(t)
	viper.Set("server.plugin", map[string]interface{}{"webhook": []interface{}{map[interface{}]interface{}{"name": "hook", "file": "hook.lua", 42: "ignored"}}})
	hooks := getWebhooks()
	if len(hooks) != 1 || hooks[0]["name"] != "hook" || hooks[0]["file"] != "hook.lua" || len(hooks[0]) != 2 {
		t.Fatal(hooks)
	}
}
