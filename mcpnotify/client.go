package mcpnotify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/model"
	"github.com/chanify/chanify/pb"
	"github.com/chanify/chanify/sender"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	http      *sender.Client
	token     string
	parsed    *model.Token
	caps      Capabilities
	nodeID    string
	newWriter func(io.Writer) *multipart.Writer
}

func NewClient(ctx context.Context, endpoint, token, version string, transport *http.Client) (*Client, error) {
	return NewClientWithTimeout(ctx, endpoint, token, version, transport, 0)
}

// NewClientWithTimeout changes only this client's HTTP timeout; ctx carries the total deadline.
func NewClientWithTimeout(ctx context.Context, endpoint, token, version string, transport *http.Client, timeout time.Duration) (*Client, error) {
	parsed, err := model.ParseToken(token)
	if err != nil || parsed.IsExpires() {
		return nil, rejected("invalid_token", "Configure a valid, unexpired sending token.")
	}
	raw, err := crypto.Base64Encode.DecodeString(strings.Split(strings.TrimPrefix(token, "cst1:"), ".")[0])
	var payload pb.Token
	if err != nil || proto.Unmarshal(raw, &payload) != nil || len(payload.DataHash) != 0 || len(payload.DeviceId) != 0 || payload.UserId == "" {
		return nil, rejected("invalid_token", "The token must be a sending credential.")
	}
	if timeout > 0 {
		copy := http.Client{}
		if transport != nil {
			copy = *transport
		}
		copy.Timeout = timeout
		transport = &copy
	}
	client, err := sender.New(endpoint, version, transport)
	if err != nil {
		return nil, rejected("config_error", err.Error())
	}
	info := &sender.Info{}
	if !client.IsDefault() {
		info, err = client.Info(ctx)
	}
	if err != nil {
		var status *sender.HTTPError
		if errors.As(err, &status) && status.Status == http.StatusNotFound && !parsed.IsServerToken() {
			info = &sender.Info{}
		} else {
			return nil, preflightError(err)
		}
	}
	caps := Capabilities{Images: slices.Contains(info.Features, "msg.image"), Files: slices.Contains(info.Features, "msg.file"), SelectableChannel: parsed.IsServerToken()}
	if parsed.IsServerToken() {
		if !slices.Contains(info.Features, "auth.server-token") || payload.NodeId != info.NodeID {
			return nil, rejected("unsupported", "Server-wide tokens require the matching node with auth.server-token support.")
		}
	} else {
		var channel pb.Channel
		if proto.Unmarshal(parsed.GetChannel(), &channel) != nil || len(channel.ProtoReflect().GetUnknown()) != 0 {
			return nil, rejected("invalid_token", "invalid fixed token channel")
		}
		switch channel.Type {
		case pb.ChanType_User:
			if channel.Name == "" {
				return nil, rejected("invalid_token", "fixed user channel requires a name")
			}
			caps.FixedChannel = channel.Name
		case pb.ChanType_Sys:
			switch channel.Code {
			case pb.ChanCode_Uncategorized:
				caps.FixedChannel = "sys.none"
			case pb.ChanCode_TimeSets:
				caps.FixedChannel = "sys.timesets"
			default:
				return nil, rejected("invalid_token", "unsupported fixed system channel")
			}
		default:
			return nil, rejected("invalid_token", "device-specific sending tokens are not supported")
		}
	}
	return &Client{http: client, token: token, parsed: parsed, caps: caps, nodeID: info.NodeID, newWriter: multipart.NewWriter}, nil
}
func (c *Client) Server(version string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "chanify", Version: version}, nil)
	for _, tool := range Tools(c.caps) {
		name := tool.Name
		server.AddTool(tool, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return c.Call(ctx, name, req.Params.Arguments), nil
		})
	}
	return server
}
func (c *Client) Call(ctx context.Context, name string, raw json.RawMessage) *mcp.CallToolResult {
	result, err := c.Send(ctx, name, raw)
	if err != nil {
		return callFailure(err)
	}
	return Result(map[string]interface{}{"status": "submitted", "format": result.Format, "request_id": result.RequestID}, false)
}

func callFailure(err error) *mcp.CallToolResult {
	var failure *SendError
	if errors.As(err, &failure) {
		code := failure.Code
		if code == "invalid_token" {
			code = "unauthorized"
		}
		return Result(map[string]interface{}{"code": code, "message": failure.Message, "status": failure.Status}, true)
	}
	return Failure("send_failed", "Cannot submit notification.")
}

func (c *Client) Send(ctx context.Context, name string, raw json.RawMessage) (*Submission, error) {
	if err := ctx.Err(); err != nil {
		return nil, preflightError(err)
	}
	if c.parsed.IsExpires() {
		return nil, rejected("invalid_token", "Sending token has expired.")
	}
	args, err := Parse(name, raw, c.caps)
	if err != nil {
		return nil, sendingValidationError(err)
	}
	// Keep discovery stable for this MCP session, but recheck attachments before sending.
	// Older sending APIs may otherwise ignore file parts and submit only their preview.
	if name != "send_text" || len(args.Text)+len(args.Title) >= 2000 {
		info, err := c.http.Info(ctx)
		if err != nil {
			return nil, preflightError(err)
		}
		feature := "msg.file"
		if name == "send_image" {
			feature = "msg.image"
		}
		if !slices.Contains(info.Features, feature) {
			return nil, rejected("unsupported", "node attachment capability is no longer available; reconnect to refresh tools")
		}
		if c.caps.SelectableChannel && (info.NodeID != c.nodeID || !slices.Contains(info.Features, "auth.server-token")) {
			return nil, rejected("unauthorized", "node no longer matches the server-wide token")
		}
	}
	var body bytes.Buffer
	writer := c.newWriter(&body)
	format, err := encodeNotification(writer, name, args, c.caps.SelectableChannel)
	if err != nil {
		return nil, err
	}
	result, err := c.http.Submit(ctx, &body, writer.FormDataContentType(), c.token)
	if err != nil {
		return nil, sendingError(err)
	}
	return &Submission{Format: format, RequestID: result.RequestID}, nil
}

func encodeNotification(writer *multipart.Writer, name string, args *Arguments, selectableChannel bool) (string, error) {
	var err error
	field := func(key, value string) error {
		if value == "" {
			return nil
		}
		return writer.WriteField(key, value)
	}
	upload := func(key, filename string, data []byte) error {
		part, err := writer.CreateFormFile(key, filename)
		if err != nil {
			return err
		}
		_, err = part.Write(data)
		return err
	}
	if selectableChannel {
		if err := field("channel", args.Channel); err != nil {
			return "", rejected("internal_error", "cannot encode channel")
		}
	}
	format := "text"
	if err := field("title", args.Title); err != nil {
		return "", rejected("internal_error", "cannot encode title")
	}
	switch name {
	case "send_text":
		if len(args.Text)+len(args.Title) < 2000 {
			err = field("text", args.Text)
		} else {
			err = upload("file", "message.txt", []byte(args.Text))
			if err == nil {
				err = field("text", Preview(args))
			}
			format = "text_attachment"
		}
	case "send_image":
		err = upload("image", "image", args.Bytes)
		format = "image"
	case "send_file":
		err = upload("file", args.Filename, args.Bytes)
		format = "file"
	}
	if err == nil && name != "send_text" {
		err = field("text", args.Summary)
	}
	if err == nil {
		err = writer.Close()
	}
	if err != nil {
		return "", rejected("internal_error", "cannot encode notification")
	}
	return format, nil
}

func sendingValidationError(err error) error {
	var validation *ValidationError
	if errors.As(err, &validation) {
		code := validation.Code
		if code == "invalid_text" {
			code = "invalid_arguments"
		}
		return rejected(code, validation.Message)
	}
	return rejected("invalid_arguments", "Invalid sending arguments.")
}
