# Codebahn CLI

Command-line interface for [Codebahn](https://codebahn.net). Managed Git and CI for European teams.

```
codebahn auth login
codebahn pr list
codebahn issue create --title "ship it"
codebahn ci dispatch --workflow ci.yml
```

One binary. OAuth login. No config files.

## Install

Download the latest binary for your platform from
[releases.codebahn.net](https://releases.codebahn.net/cli/latest.json),
or install from source:

```bash
go install github.com/codebahn/codebahn-cli/cmd/codebahn@latest
```

### Updates

The CLI checks for new versions once every 24 hours and prints a notice
when one is available. To update in place:

```bash
codebahn update
```

This downloads the new binary from `releases.codebahn.net`, verifies its
GPG signature and SHA256 checksum, and replaces itself. No external tools
required.

To check without installing: `codebahn update --check`.
To silence the periodic check: set `CODEBAHN_NO_UPDATE_CHECK=1` or add
`"check_updates": false` to `~/.config/codebahn/config.json`.

## Authentication

Log in through the browser. Tokens are saved to `~/.config/codebahn/config.json`.

```bash
codebahn auth login
codebahn auth status
```

For CI, set `CODEBAHN_TOKEN` instead.

## Commands

| Group    | What                                          |
|----------|-----------------------------------------------|
| `repo`   | Create, list, read files, branches, commits   |
| `issue`  | Create, list, comment, labels, milestones     |
| `pr`     | Create, list, merge, diff, reviews            |
| `ci`     | Dispatch workflows, list runs, read logs      |
| `search` | Code, repos, issues                           |

When you run inside a Git repo, `--owner` and `--repo` are detected from the remote.

```bash
codebahn pr merge --index 42 --style squash
codebahn repo cat --ref main --filePath README.md
codebahn ci logs --run_id 7
```

Every subcommand accepts `--help`.

## MCP and AI agents

Codebahn has a built-in [MCP server](https://codebahn.net/docs/guides/mcp/) that connects AI coding agents to your repos, issues, and CI.

The `tools/` package in this repo is the shared source of truth for both this CLI and the MCP endpoint. 54 tools, one set of types, zero drift.

```go
import "github.com/codebahn/codebahn-cli/tools"
import "github.com/codebahn/codebahn-cli/tools/schema"

td := tools.ByName("create_issue")
jsonSchema := schema.For(td)
```

## Release infrastructure

Codebahn is an EU-hosted Git platform. We apply that principle to our own
tooling: the CLI's release artifacts are hosted on our own EU infrastructure,
not on GitHub or any third-party CDN.

**How releases work:**

1. Source code lives here on GitHub (public, for discoverability).
2. When we tag a release, GitHub Actions cross-compiles binaries for
   Linux and macOS (amd64 and arm64).
3. The workflow uploads the binaries, a SHA256 checksums file, and a
   detached GPG signature to **Scaleway Object Storage** in `fr-par`
   (Paris, France), served at `releases.codebahn.net`.
4. `codebahn update` fetches from `releases.codebahn.net`. It never
   contacts GitHub.

Release artifacts are hosted on our own EU infrastructure so the CLI's
update check stays within the EU. Source code is on GitHub because it's
the standard home for open-source projects. GitHub Actions compiles the
public source; no customer data is involved in the build.

## Verifying releases

Every release includes a `checksums.txt` file signed with our GPG
release key. The CLI verifies both the signature and the checksum
automatically during `codebahn update`.

To verify manually:

```bash
# Import the public key
curl -sS https://codebahn.net/release-signing-key.asc | gpg --import

# Download and verify
curl -O https://releases.codebahn.net/cli/v0.1.0/checksums.txt
curl -O https://releases.codebahn.net/cli/v0.1.0/checksums.txt.asc
gpg --verify checksums.txt.asc checksums.txt

# Check the binary
sha256sum -c checksums.txt --ignore-missing
```

The public key is also embedded in the CLI binary itself, so
`codebahn update` does not need to fetch it separately.

## Cutting a release

Tag and push:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The GitHub Actions workflow runs tests, builds with
[GoReleaser](https://goreleaser.com/) (parallel cross-compilation, GPG
signing), and uploads everything to `releases.codebahn.net`. The
`latest.json` manifest is updated so existing installs pick up the new
version within 24 hours.

For a local release (bypassing CI): `./scripts/release.sh v0.1.0`.
Requires `goreleaser`, `gpg` (with the release key), and `aws` CLI
configured for Scaleway.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
