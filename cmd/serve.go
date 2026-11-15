//go:build !test
// +build !test

package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/chanify/chanify/core"
	"github.com/chanify/chanify/logic"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Launch chanify api server",
		Long:  `Launch service for chanify api server.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return runServe(ctx, cmd)
		},
	}
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().String("host", "0.0.0.0", "Http restful service hostname")
	serveCmd.Flags().Int("port", 80, "Http restful service port")
	serveCmd.Flags().String("endpoint", "", "Http restful service endpoint")
	serveCmd.Flags().String("name", "", "Http service name")
	serveCmd.Flags().String("datapath", "~/.chanify", "Data file path")
	serveCmd.Flags().String("filepath", "", "Store file path")
	serveCmd.Flags().String("pluginpath", "~/.chanify/plugin", "Plugin file path")
	serveCmd.Flags().String("dburl", "", "Databse dsn uri")
	serveCmd.Flags().String("secret", "", "Secret key for serverless mode")
	serveCmd.Flags().String("readtimeout", "10s", "Http restful service read timeout.")
	serveCmd.Flags().String("writetimeout", "10s", "Http restful service read timeout.")
	serveCmd.Flags().Bool("registerable", true, "Enable register user")
	serveCmd.Flags().String("whitelist", "", "Whitelist for register users")
	viper.BindPFlag("server.host", serveCmd.Flags().Lookup("host"))                      // nolint: errcheck
	viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))                      // nolint: errcheck
	viper.BindPFlag("server.endpoint", serveCmd.Flags().Lookup("endpoint"))              // nolint: errcheck
	viper.BindPFlag("server.name", serveCmd.Flags().Lookup("name"))                      // nolint: errcheck
	viper.BindPFlag("server.datapath", serveCmd.Flags().Lookup("datapath"))              // nolint: errcheck
	viper.BindPFlag("server.filepath", serveCmd.Flags().Lookup("filepath"))              // nolint: errcheck
	viper.BindPFlag("server.pluginpath", serveCmd.Flags().Lookup("pluginpath"))          // nolint: errcheck
	viper.BindPFlag("server.dburl", serveCmd.Flags().Lookup("dburl"))                    // nolint: errcheck
	viper.BindPFlag("server.secret", serveCmd.Flags().Lookup("secret"))                  // nolint: errcheck
	viper.BindPFlag("server.http.readtimeout", serveCmd.Flags().Lookup("readtimeout"))   // nolint: errcheck
	viper.BindPFlag("server.http.writetimeout", serveCmd.Flags().Lookup("writetimeout")) // nolint: errcheck
	viper.BindPFlag("server.register.enable", serveCmd.Flags().Lookup("registerable"))   // nolint: errcheck
}

// Initialize before accepting traffic, and keep the core alive until HTTP shutdown completes.
func runServe(ctx context.Context, cmd *cobra.Command) error {
	c := core.New()
	defer c.Close()
	endpoint := getEndpoint()
	opts := &logic.Options{
		Name: getName(), Version: Version, Endpoint: endpoint,
		DataPath: getExpandPath("server.datapath"), FilePath: getExpandPath("server.filepath"),
		PluginPath: getExpandPath("server.pluginpath"), DBUrl: viper.GetString("server.dburl"),
		Secret: viper.GetString("server.secret"), WebHooks: getWebhooks(),
		MCPEnabled: viper.GetBool("server.mcp.enabled"), MCPListen: viper.GetBool("server.mcp.listen"),
		AdminPassword: viper.GetString("server.web_admin.password"),
	}
	opts.Registerable, opts.RegUsers = getUserWhitlist(cmd)
	if err := c.Init(opts); err != nil {
		return fmt.Errorf("init service: %w", err)
	}
	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port")),
		ReadTimeout:    parseDuration(viper.GetString("server.http.readtimeout")),
		WriteTimeout:   parseDuration(viper.GetString("server.http.writetimeout")),
		MaxHeaderBytes: 1 << 20, Handler: c.APIHandler(),
	}
	srv.RegisterOnShutdown(c.StopMCP)
	log.Println("Launching service", srv.Addr)
	if file := viper.ConfigFileUsed(); file != "" {
		log.Println("Using config file:", file)
	}
	log.Println("Node server endpoint:", endpoint)
	return serveHTTP(ctx, srv)
}

type httpService interface {
	ListenAndServe() error
	Shutdown(context.Context) error
	Close() error
}

func serveHTTP(ctx context.Context, srv httpService) error {
	stopped := make(chan error, 1)
	go func() { stopped <- srv.ListenAndServe() }()
	var err error
	select {
	case err = <-stopped:
	case <-ctx.Done():
		log.Println("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			_ = srv.Close()
			<-stopped //nolint:errcheck // Return the shutdown error after the listener has exited.
			return fmt.Errorf("shutdown service: %w", err)
		}
		err = <-stopped
		log.Println("Shutdown service success.")
	}
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func getName() string {
	name := viper.GetString("server.name")
	if len(name) <= 0 {
		name = viper.GetString("server.hostname")
		if len(name) <= 0 {
			name = viper.GetString("server.host")
		}
	}
	return name
}

func getExpandPath(key string) string {
	path := viper.GetString(key)
	if len(path) > 0 {
		if p, err := homedir.Expand(path); err == nil {
			path = p
		}
	}
	return path
}

func getEndpoint() string {
	endpoint := viper.GetString("server.endpoint")
	if len(endpoint) <= 0 {
		hostname := viper.GetString("server.hostname")
		if len(hostname) <= 0 {
			hostname = viper.GetString("server.host")
		}
		if len(hostname) > 0 {
			port := viper.GetInt("server.port")
			switch port {
			case 80:
				endpoint = "http://" + hostname
			case 443:
				endpoint = "https://" + hostname
			default:
				endpoint = fmt.Sprintf("http://%s:%d", hostname, port)
			}
		}
	}
	return endpoint
}

func getUserWhitlist(cmd *cobra.Command) (bool, []string) {
	if viper.GetBool("server.register.enable") {
		return true, nil
	}
	users := []string{}
	if wl, err := cmd.Flags().GetString("whitelist"); err == nil && len(wl) > 0 {
		users = strings.Split(wl, ",")
	}
	return false, append(users, viper.GetStringSlice("server.register.whitelist")...)
}

func getWebhooks() []map[string]interface{} {
	plugin := viper.GetStringMap("server.plugin")
	if whs, ok := plugin["webhook"]; ok {
		if whss, ok := whs.([]interface{}); ok {
			ret := []map[string]interface{}{}
			for _, wh := range whss {
				ww := map[string]interface{}{}
				switch webhook := wh.(type) {
				case map[string]interface{}:
					ww = webhook
				case map[interface{}]interface{}:
					for k, v := range webhook {
						if key, ok := k.(string); ok {
							ww[key] = v
						}
					}
				}
				if len(ww) > 0 {
					ret = append(ret, ww)
				}
			}
			return ret
		}
	}
	return nil
}

func parseDuration(dur string) time.Duration {
	d, err := time.ParseDuration(dur)
	if err != nil {
		return 10 * time.Second
	}
	return d
}
