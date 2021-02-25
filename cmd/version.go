//+build !test

package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

var (
	// Version number
	Version = "unknown-version"
	// GitCommit code
	GitCommit = "unknown-gitcommit"
	// BuildTime RFC3339 UTC
	BuildTime = "unknown-buildtime"
)

const versionShortFormat = `chanify version %s, build %s
`
const versionTemplate = `{{.Client.Name}} version {{.Client.Version}}
Go version: {{.Client.GoVersion}}
Git commit: {{.Client.GitCommit}}
Built:      {{.Client.BuildTime.Format "Mon Jan _2 15:04:05 2006"}}
OS/Arch:    {{.Client.OS}}/{{.Client.Arch}}
`

func init() {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of chanify command line tools",
		Long:  "Show the chanify command line tools version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			tmpl, err := cmd.Flags().GetString("format")
			if err != nil || len(tmpl) <= 0 {
				tmpl = versionTemplate
			} else {
				tmpl += "\n"
			}
			t := template.New("").Funcs(template.FuncMap{"json": jsonMarshal})
			if t, err = t.Parse(tmpl); err != nil {
				return err
			}
			var data struct {
				Client struct {
					Name      string
					Version   string
					GoVersion string
					GitCommit string
					BuildTime time.Time
					OS, Arch  string
				}
			}
			buildTime, _ := time.Parse(time.RFC3339, BuildTime)
			data.Client.Name = "chanify"
			data.Client.Version = Version
			data.Client.GoVersion = runtime.Version()
			data.Client.GitCommit = GitCommit
			data.Client.BuildTime = buildTime
			data.Client.OS = runtime.GOOS
			data.Client.Arch = runtime.GOARCH
			return t.Execute(cmd.OutOrStdout(), data)
		},
	}
	versionCmd.Flags().StringP("format", "f", "", "Format the output using the given Go template")
	rootCmd.SetVersionTemplate(fmt.Sprintf(versionShortFormat, Version, GitCommit))
	rootCmd.AddCommand(versionCmd)
}

func jsonMarshal(data interface{}) string {
	res, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(res)
}
