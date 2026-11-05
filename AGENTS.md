# codebahn-cli

The `codebahn` command-line client for Codebahn (Go, cobra). The `tools/` module is the shared source of truth for the CLI and for the hosted MCP endpoint at `https://codebahn.net/mcp`; the server side is not in this repo. Product decisions live in the sibling repo `codebahn-workspace`.

## Commands

- Build: `go build -o codebahn ./cmd/codebahn` (the binary is gitignored).
- Test: `go test -race ./...`. `tools/` is its own module, so also `cd tools && go test -race ./... && go vet ./...`.
- Vet: `go vet ./...`.
- Lint: `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run ./...` runs gosec only (`.golangci.yml`); the version is pinned in `.github/workflows/ci.yml`. Add `-c ../.golangci.yml` when running inside `tools/`.
- Regenerate: `go run ./cmd/gen` rewrites `internal/gen/*.go` and `tools.json` from `tools.All`.
- Consolidated MCP registry as JSON: `go run ./cmd/gen-mcp-tools`.
- `.claude/check` runs exactly what CI runs (build, race tests, vet, lint, generated-file drift, tools module). `--quick` is the same, it takes about a minute.

## Layout

- `tools/`: one `ToolDef` per API operation (name, group, method, path template, arg struct with `json`, `required`, `desc`, `default` and `api` tags). `tools.All` is the registry. `mcp.go` groups them into consolidated `MCPTool`s that dispatch on an `action` argument. Add or change a tool here first.
- `cmd/gen` turns `tools.All` into cobra commands in `internal/gen/` and writes `tools.json`. Never hand-edit those outputs.
- `client/`: the HTTP client. `internal/config`: `config.json` with the OAuth tokens (mode 0600). `internal/oauth`, `internal/update` (signed self-update), `internal/migrate` (GitHub to Codebahn import), `internal/format` and `internal/output` (rendering).
- `cmd/codebahn`: the binary entry point; the version is injected by goreleaser.

## Release

1. Push a `vX.Y.Z` tag (`scripts/release.sh vX.Y.Z` is the local goreleaser and upload path).
2. Run the `Release` workflow (`workflow_dispatch`); it releases the newest `v*` tag.
3. It tests, runs goreleaser with a GPG-signed checksum, uploads to S3 and publishes `server.json` to the MCP registry.

## Conventions

- Every GitHub Action is pinned to a full commit SHA with the version in a trailing comment. Keep that for anything you add.
- CI fails when `internal/gen` or `tools.json` drift from `go run ./cmd/gen`; regenerate and commit.
- Dependencies are few on purpose (cobra, go-crypto, x/term, x/mod); the `tools/` module has none.

## Security

Path-scoped OWASP ASVS 5.0 rules are generated into `.claude/rules/security/` from `.claude/asvs.map.yaml`; regenerate them with `.claude/render-rules`. The constraints below apply everywhere. For MCP-specific guidance see the OWASP MCP Security cheat sheet: https://cheatsheetseries.owasp.org/cheatsheets/MCP_Security_Cheat_Sheet.html

Derived from the OpenSSF Security-Focused Guide for AI Code Assistant Instructions (2025-08-01, https://best.openssf.org/Security-Focused-Guide-for-AI-Code-Assistant-Instructions) and from securityblog.omegapoint.se (CC BY-SA 4.0); the reasoning is in research/agent-context-and-memory-tooling-2026-11.

- Never put secrets, tokens or credentials in code, config, tests, fixtures, logs or commit history. Read them from the environment or a vault reference.
- Never disable a security feature to make something work: TLS verification, CSRF, CSP, audience or signature checks, type checks in deserialization, lint or scanner rules. Ask instead.
- Never build SQL, shell commands, file paths or HTML from untrusted input by concatenation. Use parameterized queries, argument arrays, path validation and context-aware output encoding. Never sanitize HTML by regex rejection.
- Never accept an ID token as an access token. Every JWT check verifies an algorithm allow-list, issuer, audience, type and expiry, and signing keys are never shared across token purposes.
- Never match redirect URIs by prefix, wildcard or regex: exact string match against a registered list, after one decode.
- Never make a route anonymous by omission. New routes are authenticated by default; anonymous routes are allow-listed in one place with a comment saying why.
- Never take identity from a header a proxy could rewrite, and never enforce authorization only at a gateway or in client code; the service layer decides.
- Data-level access failures answer 404 for not-my-data, 403 for a missing permission and 401 when unauthenticated; never reveal that another tenant's object exists.
- Never log tokens, secrets, passwords or personal data. Log security-relevant denials and failures as warnings with enough context to investigate.
- Never add a dependency without a reason, a known maintainer and a pinned version. Prefer the standard library.
- Never compare secrets, tokens or hashes with ordinary equality; use constant-time comparison.
- Never leave security-relevant code without negative tests: unauthenticated, wrong permission and not-my-data for every endpoint, and a deny test for every permission mapping.

Before finishing a change that touches auth, sessions, tokens, permissions, uploads, SQL, shell, crypto or config, review your own diff against this list and fix what you find.
