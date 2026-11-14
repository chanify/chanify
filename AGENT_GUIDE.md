# Chanify for coding agents

This is a standalone operating guide for Chanify 1.5.0. Give this file to your coding agent to set up notifications through CLI or MCP. Examples contain placeholders, never usable credentials. Verify the installed binary exposes the commands below; older published binaries may not include them.

## Instructions to the agent

1. Identify the user's available connection: an installed `chanify` binary plus a private client configuration, an existing MCP connection, or a remote MCP URL plus a secret reference. Ask only for missing setup details. Do not ask the user to paste secrets into the conversation.
2. Prefer `chanify agent send` for one-shot shell tasks and local attachments. Use local stdio MCP when the agent supports MCP subprocesses; use remote MCP when a remote connection is already provisioned. Do not configure all three routes or start a server unnecessarily.
3. Use the user's intended destination and notification policy. If unspecified, send one concise completion or failure summary when the requested work ends; do not send progress spam. Sending a setup test requires the user's authorization to notify them.
4. Keep credentials in the user's private configuration or the MCP client's secret facility. Never print or commit secrets, include them in notification text, tool arguments, URLs, screenshots, or diagnostic reports. Do not read an entire secret configuration into the conversation. Do not change endpoint or credential to work around a rejection.
5. Discover actual command/tool capabilities. Use absolute paths for the executable, configuration, and attachments. Do not invent tools, channels, flags, delivery receipts, or retry guarantees.
6. Report `submitted` as submitted, never delivered/read. Do not automatically repeat an ambiguous send, even with the same JSON-RPC ID. Explain the duplicate risk before a user-requested retry.
7. Do not create/revoke credentials, change node configuration, or delete media unless that administration task is authorized. Sending credentials do not grant administration access.

## Setup inputs and credential types

| Route | Required setup | Credential | Needs server `/mcp`? |
| --- | --- | --- | --- |
| CLI `agent send` | Binary, private client YAML | Channel sending token or registered server-wide `cst1:` sending token | No |
| Local stdio `mcp` | Same binary/YAML, MCP subprocess support | Same sending token | No |
| Remote Streamable HTTP | Self-hosted node URL and MCP client header support | Independent `cmcp1.` MCP token | Yes |

A management password is only for `/admin`. A remote MCP token cannot be used in the CLI/local stdio configuration. A sending token cannot authenticate to remote `/mcp`.

The user obtains a channel token in the Apple client, or creates a server-wide sending token in iOS node details → Sending Tokens. For the latter, the node must support `auth.server-token` and be bound with **Store device info** enabled. The token must be registered on that same node. Never generate a private account key or fabricate a sending token.

Use an existing private file, or have the user save the following outside their repository (restrict permissions, for example `chmod 600` on Unix):

```yaml
client:
  endpoint: https://your-node.example
  token: '<sending-token>'
```

The endpoint is the node base URL, not `/mcp` or `/v1/sender`. Default endpoint is `https://api.chanify.net`. Use HTTPS remotely; loopback HTTP is suitable for local validation. Keep the endpoint paired with its intended credential.

## CLI: one notification per invocation

Check the installed command first:

```sh
chanify version
chanify agent send --help
chanify mcp --help
```

If unavailable, install a release containing these commands or build this source with the Go version declared in `go.mod` and `make build`. Do not assume the currently published `latest` binary matches this checkout.

Examples below assume `chanify` is in PATH. Replace paths with the user's actual absolute paths. Choose **exactly one** content flag:

```sh
chanify --config /absolute/path/client.yml agent send --text 'Build and tests passed' --title 'Task complete'
chanify --config /absolute/path/client.yml agent send --text-file /absolute/path/report.md --title 'Report'
chanify --config /absolute/path/client.yml agent send --stdin < /absolute/path/report.md
chanify --config /absolute/path/client.yml agent send --image /absolute/path/screenshot.png --summary 'UI validation'
chanify --config /absolute/path/client.yml agent send --file /absolute/path/report.pdf --summary 'Full report'
```

`--text` is literal; `@file` does not read a file. Paths refer to local regular files, not remote URLs. `--text-file` and `--stdin` require UTF-8 text and preserve its full contents. Use `--summary` for an attachment caption, rather than combining `--file` and `--text`. Optional `--timeout 30s` is the total input/discovery/send budget, not a separate budget per stage. `--endpoint` overrides the configured node only when the user intends that change. There is no `--token`, `--json`, or `--dry-run` on `agent send`; use the private configuration. Legacy `chanify send` remains available, but its flags/output differ.

Stdout contains one JSON object (except `--help`); stderr is diagnostic. Parse the JSON and process exit status. Example:

```json
{"schema_version":1,"status":"submitted","format":"text","request_id":"<request-id>"}
```

| Exit | Meaning | Action |
| --- | --- | --- |
| 0 | `submitted`, server confirmation with request ID | Report submission; do not resend |
| 1 | Internal preparation/output failure | Inspect diagnostics without exposing secrets |
| 2 | Invalid arguments/configuration/local input | Correct input before retrying |
| 3 | Invalid or rejected credential | Ask user to check expiry, node, registration or revocation |
| 4 | Unsupported capability, discovery failure, or timeout before send | Check endpoint/capabilities/network |
| 5 | `unknown`, submission may already have happened | Do not automatically retry |
| 130 / 143 | Interrupted | Inspect JSON status; interruption after send may be `unknown` |

Failures contain `error.code` and `error.message`; use stable codes rather than parsing prose. `format` can be `text`, `text_attachment`, `image`, or `file`. A 400/404/429, 5xx, timeout during send, or malformed confirmation may have an uncertain outcome. No automatic retries or idempotency keys are provided.

## Local stdio MCP

Configure your MCP client to launch the executable as a subprocess with these arguments. This is a common configuration shape, **not a client-specific installation command**; adapt field names to the host's supported format:

```json
{
  "mcpServers": {
    "chanify": {
      "command": "/absolute/path/chanify",
      "args": ["--config", "/absolute/path/client.yml", "mcp"]
    }
  }
}
```

Restart/reconnect that MCP connection after changing its setup. The host owns stdin/stdout and performs initialization/tool discovery; do not run `chanify mcp` as a one-shot shell send or write human text to stdin. No local listening port, remote MCP credential, or server MCP switch is required. Stdout is reserved for protocol messages; diagnostics use stderr.

## Remote MCP (self-hosted)

The node operator merges this into the existing server configuration and restarts the node, preserving its database, identity and other settings:

```yaml
server:
  mcp:
    enabled: true
    listen: false
```

Obtain a dedicated MCP token via `/admin` → MCP Tokens → Create, selecting a user already bound with device storage enabled. Alternatively, an authorized node operator can use local database access:

```sh
chanify --config /absolute/path/server.yml mcp-token create --user '<user-id>' --name coding-agent --expires-in 2160h
chanify --config /absolute/path/server.yml mcp-token list --user '<user-id>'
chanify --config /absolute/path/server.yml mcp-token rename --user '<user-id>' --id '<credential-id>' --name new-name
chanify --config /absolute/path/server.yml mcp-token revoke --user '<user-id>' --id '<credential-id>'
```

These commands open the node database; they are not remote admin commands. Creation returns the full secret once; provision it directly into a private secret store, never a committed file or shared log. Names are mutable, expiry is immutable. Revoke only credentials selected by the user.

Configure the agent host using these connection parameters:

```text
Transport: Streamable HTTP
URL: https://your-node.example/mcp
Authorization header: Bearer <mcp-token-from-secret-store>
```

The concrete configuration syntax depends on the host. OAuth login and legacy GET SSE/HTTP+SSE are not implemented. Normal tool calls work with `listen: false`; do not require optional SSE subscriptions for sending. A 404 at `/mcp` usually means the switch is off, the path is wrong, or a proxy route is missing. A 401 means credentials were rejected.

## MCP tools and shared content limits

Initialize and discover the tools first; do not assume every node supports attachments. Remote MCP and local stdio expose the following according to node capabilities:

| Tool | Required arguments | Optional arguments |
| --- | --- | --- |
| `send_text` | `text` | `title`, `summary`, and `channel` when offered |
| `send_image` | `data_base64` | `filename`, `title`, `summary`, and `channel` when offered |
| `send_file` | `filename`, `data_base64` | `title`, `summary`, and `channel` when offered |

Example `send_text` arguments (not a complete JSON-RPC request):

```json
{"text":"Build and tests passed","title":"Task complete"}
```

For attachments, provide standard Base64 of the actual bytes. MCP does not read local paths or fetch URLs. Prefer the CLI file flags when sending local artifacts from a shell; avoid manually placing large Base64 strings in conversation history.

| Content | Limit |
| --- | --- |
| Text on default service or node without file storage | 400 Unicode characters |
| Text on self-hosted node with file storage | 100,000 Unicode characters |
| Title / summary | 80 / 200 Unicode characters |
| Image | PNG/JPEG, 4 MiB decoded, 32 million pixels |
| File | 4 MiB decoded; filename ≤128 Unicode characters, without directory components |
| Remote MCP request | 6 MiB |

Long text is attached in full as UTF-8 `message.txt` when supported; only the preview may be shortened. Never silently truncate a report or split it into unsolicited notifications. The default service supports short text through CLI/local stdio, not these self-hosted attachment workflows.

Ordinary channel tokens bind a fixed destination: **omit `channel` entirely**; even specifying the same channel is rejected. Server-wide sending tokens and remote MCP accept readable channel codes, defaulting to `sys.none` (Uncategorized). User codes are 1–64 ASCII letters/digits/underscore/dot/hyphen, starting with a letter or digit. `sys.timesets` means Timeline; device-specific channels are unavailable. Do not supply an encoded protobuf ID, empty string or null. There is no `list_channels` tool or iOS channel-directory synchronization; ask the user for a code when necessary.

Inspect MCP tool results, including `isError` and structured/text content, even if HTTP or JSON-RPC succeeded. Success includes `status: submitted`, `format`, and `request_id`. A transport success is not proof of notification success. A `send_failed`, `outcome_unknown`, disconnect or timeout must not trigger an automatic repeat: delivery may already have occurred for some devices.

## Delivery troubleshooting

1. Confirm endpoint, credential type, expiry and intended channel. Do not expose secrets while checking them.
2. Check that the Apple client is bound to this node, device storage is enabled where required, notifications are authorized, and a push token is registered.
3. `no devices found` means no usable registered device was selected. `no devices send success` means no attempted device submission was counted as successful; inspect server `Send apns failed` / `APNs rejected notification` logs for network or APNs status/reason.
4. Localhost reachability only tests the agent-to-node connection. The node separately needs outbound APNs access. A sandbox may permit local HTTP while blocking APNs. Run the same local node outside that sandbox for a controlled comparison; never start a second instance on the same port or change its database identity.
5. Server acceptance is not a read receipt. For an end-to-end setup test, verify the exact test message in the Apple client. Do not claim that a simulator can never receive remote pushes.

## Copyable user handoff

> Read AGENT_GUIDE.md and use Chanify to notify me when the requested task completes or fails. The installed binary is `<absolute-binary-path>` and the private client configuration is `<absolute-config-path>`; alternatively use my already configured Chanify MCP connection. Keep credentials private. Use the configured destination unless I specify a channel. Send one concise summary, inspect submission results, and do not retry an uncertain send automatically. If setup is incomplete, tell me exactly which non-secret information or local configuration is missing. Send a setup test only when I authorize it.
