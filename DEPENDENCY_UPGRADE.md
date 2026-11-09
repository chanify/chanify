# Go 1.27.3 dependency upgrade review

Reviewed on 2026-11-08. Exact versions and checksums are recorded in `go.mod`
and `go.sum`.

## Changes

- Gin 1.9.0 -> 1.12.0; Protobuf 1.28.1 -> 1.36.12.
- APNS 0.23.0 -> 0.25.0; JWT v4 4.4.1 -> 4.5.2.
- MySQL driver 1.7.0 -> 1.10.1; SQLite 1.21.0 -> 1.62.0.
- Viper 1.15.0 -> 1.21.0; Cobra 1.6.1 -> 1.10.2.
- fsnotify 1.6.0 -> 1.10.1; GopherLua 1.1.0 -> 1.1.2.
- Update x/image, networking dependencies, and Sonic. Pin quic-go 0.59.1
  to fix GO-2026-5676 introduced through Gin's HTTP/3 dependency.
- Accept both old and new map representations when reading webhook settings.
- Wait for the file watcher to exit before clearing its pointer on shutdown;
  handle a closed error channel without spinning.
- Reject unreadable or invalid existing node keys at startup instead of
  silently replacing them. Only a missing option initializes a new node key.
- Correct the account deletion test fixture to include push tokens and check
  that unrelated device keys remain intact.

## Compatibility evidence

- Go 1.27.3 builds and `go vet ./...` pass on macOS arm64.
- Race tests pass with `go test -race -skip '^TestSendForward$' ./...`,
  including command/config tests excluded by the historical `test` build tag.
  `TestSendForward` was also run separately in the user's normal terminal and
  passed with the race detector enabled.
- Legacy private scalars, public keys, node IDs, signatures, and ciphertexts
  retain their existing regression vectors. Configured secrets retain their
  original bytes. APNS ES256 JWT generation and signature verification pass.
- A database created with SQLite driver 1.21.0 was read with 1.62.0: stored
  node, user, and device keys remained unchanged. Persisted node keys also
  survive service restarts without being rewritten.
- CGO-disabled cross-builds pass for Linux amd64, arm64, 386, ARMv7;
  macOS amd64; and Windows amd64.
- `govulncheck -show verbose ./...` finds no affected called symbols or
  imported packages. It reports GO-2026-5932 for the unused
  `golang.org/x/crypto/openpgp` package at module scope; this service does not
  import that package.

## Deployment checks

The isolated tool environment cannot listen on a local TCP port. The forwarding
integration test passed in a normal terminal with:

```sh
go test -race -tags test ./core -run '^TestSendForward$' -count=1
```

Live MySQL, central forwarding, and APNS delivery were not tested against
production backends. Validate the existing MySQL DSN and charset/collation
before deployment: the driver changed database-name escaping and default
charset negotiation after 1.7. Back up the database and compare the public node
ID before and after rollout; keep the configured secret unchanged.

This change does not migrate protobuf field definitions, JSON APIs, key formats,
or legacy derivation rules. The old GoReleaser/lint/CodeQL workflow configuration
is a separate deployment-tooling follow-up; application dependency tests do not
validate those tools.
