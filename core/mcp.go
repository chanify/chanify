package core

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/logic"
	"github.com/chanify/chanify/mcpnotify"
	"github.com/chanify/chanify/model"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpAuthKey struct{}
type mcpAuthorization struct {
	raw    string
	record *model.MCPTokenRecord
}

func (c *Core) initMCP(listen bool) http.Handler {
	var mu sync.Mutex
	servers := map[string]*mcp.Server{}
	transport := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		auth, _ := req.Context().Value(mcpAuthKey{}).(*mcpAuthorization)
		if auth == nil {
			return nil
		}
		mu.Lock()
		defer mu.Unlock()
		if server := servers[auth.record.ID]; server != nil {
			return server
		}
		server := c.newMCPServer(listen)
		servers[auth.record.ID] = server
		return server
	}, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true, MaxRequestBodyBytes: 6 << 20, PropagateRequestCancellation: true})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.mcpMu.Lock()
		if c.mcpStopped {
			c.mcpMu.Unlock()
			http.Error(w, "server shutting down", http.StatusServiceUnavailable)
			return
		}
		c.mcpRequests.Add(1)
		c.mcpMu.Unlock()
		defer c.mcpRequests.Done()

		if origin := r.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)
			if err != nil || u.Host != r.Host || (u.Scheme != "http" && u.Scheme != "https") {
				http.Error(w, "invalid origin", http.StatusForbidden)
				return
			}
		}
		fields := strings.Fields(r.Header.Get("Authorization"))
		if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
			w.Header().Set("WWW-Authenticate", `Bearer realm="chanify-mcp"`)
			http.Error(w, "MCP bearer token required", http.StatusUnauthorized)
			return
		}
		record, err := c.logic.AuthenticateMCP(fields[1])
		if err != nil {
			http.Error(w, "invalid MCP credential", http.StatusUnauthorized)
			return
		}
		if !listen && r.Header.Get("Mcp-Method") == "subscriptions/listen" {
			http.Error(w, "MCP listening is disabled", http.StatusMethodNotAllowed)
			return
		}
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		stop := context.AfterFunc(c.mcpContext, cancel)
		defer stop()
		ctx = context.WithValue(ctx, mcpAuthKey{}, &mcpAuthorization{raw: fields[1], record: record})
		// Existing HTTP write timeouts must not kill an opted-in notification stream.
		if listen {
			_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
		}
		w.Header().Set("X-Accel-Buffering", "no")
		started, finished := c.beginAdminActivity(r, "mcp", record.ID, record.UserID)
		defer finished()
		writer := &mcpStreamWriter{ResponseWriter: w, onStream: started}
		// Covers expiry and revocation performed by a separate local management process.
		done := make(chan struct{})
		monitorDone := make(chan struct{})
		defer func() { close(done); <-monitorDone }()
		go func() {
			defer close(monitorDone)
			ticks := 0
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-done:
					return
				case <-ctx.Done():
					return
				case <-ticker.C:
					ticks++
					if listen && ticks%15 == 0 {
						writer.heartbeat()
					}
					if _, err := c.logic.AuthenticateMCP(fields[1]); err != nil {
						cancel()
						return
					}
				}
			}
		}()
		transport.ServeHTTP(writer, r.WithContext(ctx))
	})
}
func (c *Core) newMCPServer(listen bool) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "chanify", Version: c.logic.Version}, &mcp.ServerOptions{Capabilities: &mcp.ServerCapabilities{Tools: &mcp.ToolCapabilities{ListChanged: listen}}})
	for _, tool := range mcpnotify.Tools(c.mcpCapabilities()) {
		server.AddTool(tool, c.mcpTool(tool.Name))
	}

	return server
}
func (c *Core) currentMCPAuth(ctx context.Context) (*mcpAuthorization, error) {
	auth, _ := ctx.Value(mcpAuthKey{}).(*mcpAuthorization)
	if auth == nil {
		return nil, errors.New("missing MCP authorization")
	}
	record, err := c.logic.AuthenticateMCP(auth.raw)
	if err != nil {
		return nil, errors.New("MCP credential expired or revoked")
	}
	return &mcpAuthorization{raw: auth.raw, record: record}, ctx.Err()
}
func (c *Core) mcpTool(name string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		auth, err := c.currentMCPAuth(ctx)
		if err != nil {
			return mcpFailure("unauthorized", err.Error()), nil
		}
		args, err := mcpnotify.Parse(name, req.Params.Arguments, c.mcpCapabilities())
		if err != nil {
			return mcpnotify.ValidationFailure(err), nil
		}

		channel, err := logic.MCPChannelData(args.Channel)
		if err != nil {
			return mcpFailure("invalid_channel", "invalid configured channel"), nil
		}
		node, _ := crypto.Base32Encode.DecodeString(c.logic.NodeID)
		msg := model.NewMessageDestination(node, channel, nil)
		format := "text"
		if name == "send_text" {
			if len(args.Text)+len(args.Title) < 2000 {
				msg.TextContent(args.Text, args.Title, "", "")
			} else {
				if !c.mcpCanFiles() {
					return mcpFailure("unsupported", "long text storage is unavailable"), nil
				}
				path, err := c.logic.SaveFile("files", []byte(args.Text))
				if err != nil {
					return mcpFailure("storage_failed", "could not save text attachment"), nil
				}
				summary := mcpnotify.Preview(args)
				msg.TextFileContent(path, "message.txt", args.Title, summary, len([]byte(args.Text)), nil)
				format = "text_attachment"
			}
		} else {
			if !c.logic.CanFileStore() || (name == "send_file" && !c.mcpCanFiles()) || (name == "send_image" && !slices.Contains(c.logic.Features, "msg.image")) {
				return mcpFailure("unsupported", "attachment storage is unavailable"), nil
			}
			data := args.Bytes
			if name == "send_image" {
				path, err := c.logic.SaveFile("images", data)
				if err != nil {
					return mcpFailure("storage_failed", "could not save image"), nil
				}
				msg.ImageContent(path, createThumbnail(data), len(data))
				format = "image"
			} else {
				path, err := c.logic.SaveFile("files", data)
				if err != nil {
					return mcpFailure("storage_failed", "could not save attachment"), nil
				}
				msg.FileContent(path, args.Filename, args.Summary, len(data), nil)
				format = "file"
			}
		}
		if name != "send_text" {
			msg.AttachmentCaption(args.Title, args.Summary)
		}

		// Recheck after storage work before producing an external notification.
		if _, err := c.currentMCPAuth(ctx); err != nil {
			return mcpFailure("unauthorized", err.Error()), nil
		}
		result := &mcpSendResult{}
		c.sendDirectForUser(result, auth.record.UserID, msg)
		if result.code != 200 {
			return mcpFailure("send_failed", result.message()), nil
		}
		return mcpResult(map[string]interface{}{"status": "submitted", "format": format, "request_id": result.data["request-uid"]}, false), nil
	}
}
func mcpResult(data map[string]interface{}, failed bool) *mcp.CallToolResult {
	return mcpnotify.Result(data, failed)
}
func mcpFailure(code, message string) *mcp.CallToolResult {
	return mcpResult(map[string]interface{}{"code": code, "message": message}, true)
}

type mcpSendResult struct {
	code int
	data map[string]interface{}
}

func (r *mcpSendResult) JSON(code int, obj interface{}) {
	r.code = code
	body, _ := json.Marshal(obj)
	_ = json.Unmarshal(body, &r.data)
}
func (r *mcpSendResult) DataFromReader(code int, _ int64, _ string, reader io.Reader, _ map[string]string) {
	r.code = code
	_ = json.NewDecoder(reader).Decode(&r.data)
}
func (r *mcpSendResult) message() string {
	if text, ok := r.data["msg"].(string); ok {
		return text
	}
	return "notification submission failed"
}

func (c *Core) mcpCanFiles() bool {
	return c.logic.CanFileStore() && slices.Contains(c.logic.Features, "msg.file")
}

func (c *Core) mcpCapabilities() mcpnotify.Capabilities {
	return mcpnotify.Capabilities{SelectableChannel: true, Files: c.mcpCanFiles(), Images: c.logic.CanFileStore() && slices.Contains(c.logic.Features, "msg.image")}
}
