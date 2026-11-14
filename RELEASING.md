# Build and release

The workflows use the Go version declared in `go.mod`. CI checks module files,
runs lint and race tests, and builds a GoReleaser snapshot for every supported
release target. `make test` runs command tests without build tags, then the
historical `test` configuration with a single coverage profile. `gocovmerge` is
no longer required. Override `TEST_FLAGS` to change test options.

GoReleaser is pinned in `.github/workflows/ci.yml`; lint is pinned in the same
file. Update these versions deliberately and run the checks below.

```sh
make build
make test
goreleaser check
goreleaser release --snapshot --clean
```

Use GoReleaser v2.20.0 for local release checks. Snapshot archives are written
to `dist/` without publishing. Existing archive names and release platforms are
preserved. Docker additionally supports Linux 386.

After committing the intended changes, push a `v`-prefixed version tag such as
`v1.5.0` to publish. GitHub Release and Docker publishing run independently after
CI passes. The release job uses `GITHUB_TOKEN` with `contents: write`; Docker
requires the repository secrets `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN`.
Coverage uploads use `CODECOV_TOKEN` when configured.

Docker tag behavior is unchanged:

- `main` publishes `wizjin/chanify:latest`.
- `dev` publishes `wizjin/chanify:dev`.
- `v1.5.0` publishes `wizjin/chanify:1.5.0`.

Thus `latest` follows `main`, rather than the latest version tag. Docker builds
cross-compile on the build host with CGO disabled, without QEMU. CI passes the Go
version from `go.mod` as the `GO_VERSION` build argument. For local builds, the
Dockerfile has a default; override it when testing a different Go version:

```sh
docker buildx build --build-arg GO_VERSION=1.27.3 \
  --platform linux/amd64,linux/386,linux/arm64,linux/arm/v7 .
```

The source version in `cmd/version.go` is 1.5.0. `go run . version` and
`make build` use this version; `make build VERSION=...` can override it.
GoReleaser uses the release tag, which must match the source version.

## 1.4.4 review (2026-11-10)

No additional release-blocking application defect was found in the reviewed
working-tree changes (webhook configuration compatibility, watcher shutdown,
and persisted node-key error handling). This is not production approval.

Validated in this session:

- `make build` succeeds; the binary reports version 1.4.4.
- `go test -race ./cmd` passes (run by `make test`).
- `go test -race -tags test -skip '^TestSendForward$' ./...` passes.
- `GOCACHE=/private/tmp/chanify-go-cache go vet ./...` passes.
- `git diff --check` passes.

Outstanding release gates:

- Full `make test`: the installed local Go toolchain lacks `covdata` and its
  coverage build fails. The forwarding integration test also needs permission
  to bind a local port, unavailable in this sandbox. Do not remove either check
  from CI to accommodate these local limitations.
- GoReleaser check/snapshot, golangci-lint, and Docker builds were not executed:
  those executables are not installed in this environment.
- GitHub Actions run status and publishing credentials were not verified;
  the local network proxy is unreachable. Workflow inspection alone does not
  establish a successful build or release.
- Confirm CI passes for the final commit, including release snapshots, then
  validate MySQL, forwarding, and APNS in staging before publishing v1.4.4.
