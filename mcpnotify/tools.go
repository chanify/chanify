package mcpnotify

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const FileLimit = 4 << 20

type Capabilities struct {
	Images            bool
	Files             bool
	SelectableChannel bool
	FixedChannel      string
}
type Arguments struct {
	Channel  string `json:"channel"`
	Text     string `json:"text"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Data     string `json:"data_base64"`
	Filename string `json:"filename"`
	Bytes    []byte `json:"-"`
}

func Tools(caps Capabilities) []*mcp.Tool {
	tools := []*mcp.Tool{}
	add := func(tool *mcp.Tool) { tools = append(tools, tool) }
	schema := func(properties map[string]interface{}, required ...string) map[string]interface{} {
		result := map[string]interface{}{"type": "object", "properties": properties, "additionalProperties": false}
		if len(required) > 0 {
			result["required"] = required
		}
		return result
	}
	destination := "Omitting channel sends to sys.none."
	if !caps.SelectableChannel {
		destination = "The destination is fixed by the configured sending token (" + caps.FixedChannel + "). No channel override is accepted."
	}
	channelField := func(fields map[string]interface{}) map[string]interface{} {
		if !caps.SelectableChannel {
			delete(fields, "channel")
		}
		return fields
	}
	channel := map[string]interface{}{"type": "string", "minLength": 1, "maxLength": 64, "default": "sys.none", "description": "Optional channel code; omitted sends to Uncategorized (sys.none). An explicit empty or invalid code is an error. User codes contain letters/digits/underscore/dot/hyphen, beginning with a letter or digit."}
	shortString := func(max int) map[string]interface{} {
		return map[string]interface{}{"type": "string", "maxLength": max}
	}
	maxText := 400
	if caps.Files {
		maxText = 100000
	}
	description := fmt.Sprintf("Send a text notification or complete report to the user. text accepts at most %d Unicode characters; title at most 80. %s Long text is automatically attached in full when needed; it is never silently truncated. summary optionally controls the notification preview. Success means submitted, not read. Retrying an ambiguous failure may duplicate a notification.", maxText, destination)
	textFields := map[string]interface{}{"channel": channel, "text": map[string]interface{}{"type": "string", "minLength": 1, "maxLength": maxText}, "title": shortString(80), "summary": shortString(200)}
	additive := false
	sendingHints := &mcp.ToolAnnotations{DestructiveHint: &additive, OpenWorldHint: &additive}
	add(&mcp.Tool{Name: "send_text", Annotations: sendingHints, Description: description, InputSchema: schema(channelField(textFields), "text")})

	if caps.Images || caps.Files {
		for _, name := range []string{"send_image", "send_file"} {
			if (name == "send_image" && !caps.Images) || (name == "send_file" && !caps.Files) {
				continue
			}
			fields := map[string]interface{}{"channel": channel, "title": shortString(80), "summary": shortString(200), "data_base64": map[string]interface{}{"type": "string", "minLength": 1, "maxLength": base64.StdEncoding.EncodedLen(FileLimit), "description": "Base64-encoded file bytes, at most 4 MiB decoded. Local paths and remote URLs are not accepted."}, "filename": shortString(128)}
			required := []string{"data_base64"}
			desc := "Send a file attachment with an optional notification summary; at most 4 MiB decoded. Provide filename."
			if name == "send_image" {
				desc = "Send a PNG or JPEG image, at most 4 MiB and 32 million pixels, with an optional description. Provide image bytes as data_base64."
			} else {
				required = append(required, "filename")
			}
			add(&mcp.Tool{Name: name, Annotations: sendingHints, Description: desc + " " + destination, InputSchema: schema(channelField(fields), required...)})
		}
	}
	return tools
}

type ValidationError struct{ Code, Message string }

func (e *ValidationError) Error() string { return e.Message }
func invalid(code, message string) error { return &ValidationError{code, message} }
func Parse(name string, raw json.RawMessage, caps Capabilities) (*Arguments, error) {
	var supplied map[string]json.RawMessage
	if len(raw) > 6<<20 || json.Unmarshal(raw, &supplied) != nil || supplied == nil {
		return nil, invalid("invalid_arguments", "provide an arguments object of at most 6 MiB")
	}
	allowed := map[string]bool{"title": true, "summary": true}
	if caps.SelectableChannel {
		allowed["channel"] = true
	}
	switch name {
	case "send_text":
		allowed["text"] = true
	case "send_image", "send_file":
		if (name == "send_image" && !caps.Images) || (name == "send_file" && !caps.Files) {
			return nil, invalid("unsupported", "attachment capability is unavailable")
		}
		allowed["data_base64"] = true
		allowed["filename"] = true
	default:
		return nil, invalid("unsupported", "unknown sending tool")
	}
	for key, value := range supplied {
		if !allowed[key] {
			return nil, invalid("invalid_arguments", "arguments do not match this tool; fixed-channel tokens do not accept channel overrides")
		}
		if string(value) == "null" {
			return nil, invalid("invalid_arguments", "explicit null arguments are not accepted")
		}
	}
	args := &Arguments{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(args) != nil {
		return nil, invalid("invalid_arguments", "invalid tool arguments")
	}
	if caps.SelectableChannel {
		if _, exists := supplied["channel"]; !exists {
			args.Channel = "sys.none"
		}
		if _, err := ChannelData(args.Channel); err != nil {
			return nil, invalid("invalid_channel", "provide a valid channel code")
		}
	}
	if utf8.RuneCountInString(args.Title) > 80 || utf8.RuneCountInString(args.Summary) > 200 {
		return nil, invalid("too_large", "title accepts 80 and summary 200 Unicode characters")
	}
	if name == "send_text" {
		limit := 400
		if caps.Files {
			limit = 100000
		}
		if strings.TrimSpace(args.Text) == "" {
			return nil, invalid("invalid_text", fmt.Sprintf("text must contain 1 to %d Unicode characters", limit))
		}
		if utf8.RuneCountInString(args.Text) > limit {
			return nil, invalid("too_large", fmt.Sprintf("text accepts at most %d Unicode characters", limit))
		}
		return args, nil
	}
	if len(args.Data) > base64.StdEncoding.EncodedLen(FileLimit) {
		return nil, invalid("too_large", "attachment exceeds 4 MiB")
	}
	data, err := base64.StdEncoding.Strict().DecodeString(args.Data)
	if err != nil || len(data) == 0 || len(data) > FileLimit {
		return nil, invalid("invalid_data", "provide valid Base64 bytes, at most 4 MiB")
	}
	args.Bytes = data
	if name == "send_image" {
		cfg, kind, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil || (kind != "png" && kind != "jpeg") || cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > 32000000/cfg.Height {
			return nil, invalid("invalid_image", "provide a PNG or JPEG of at most 32 million pixels")
		}
		if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
			return nil, invalid("invalid_image", "provide a complete, decodable PNG or JPEG")
		}
	} else if args.Filename == "" || args.Filename == "." || args.Filename == ".." || strings.ContainsAny(args.Filename, "/\\\x00") || utf8.RuneCountInString(args.Filename) > 128 {
		return nil, invalid("invalid_filename", "provide a filename without directory components")
	}
	return args, nil
}
func Preview(args *Arguments) string {
	if args.Summary != "" {
		return args.Summary
	}
	runes := []rune(args.Text)
	if len(runes) > 160 {
		return string(runes[:160]) + "…"
	}
	return args.Text
}
func Result(data map[string]interface{}, failed bool) *mcp.CallToolResult {
	body, _ := json.Marshal(data)
	return &mcp.CallToolResult{IsError: failed, StructuredContent: data, Content: []mcp.Content{&mcp.TextContent{Text: string(body)}}}
}
func Failure(code, message string) *mcp.CallToolResult {
	return Result(map[string]interface{}{"code": code, "message": message}, true)
}
func ValidationFailure(err error) *mcp.CallToolResult {
	var validation *ValidationError
	if errors.As(err, &validation) {
		return Failure(validation.Code, validation.Message)
	}
	return Failure("invalid_arguments", "invalid tool arguments")
}
