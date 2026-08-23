# CLI Self-Update

Ship `codebahn update` and a periodic version check. Source code on GitHub,
builds on GitHub Actions, release artifacts on Scaleway Object Storage
(`releases.codebahn.net`). Users never touch GitHub.

## Files involved

### New (codebahn-cli)

| File | Purpose |
|---|---|
| `internal/update/check.go` | Version check: fetch `latest.json`, compare semver, cache timestamp |
| `internal/update/selfupdate.go` | Download binary, verify SHA256 + GPG signature, atomic replace |
| `internal/update/cache.go` | `CacheDir()` helper (`XDG_CACHE_HOME` / `~/.cache/codebahn`) |
| `cmd/codebahn/update.go` | `codebahn update` cobra command |
| `scripts/release.sh` | Cross-compile, checksum, GPG sign, upload to Scaleway bucket |
| `.github/workflows/release.yml` | On tag push: run `release.sh`, upload artifacts to bucket |
| `pgp/release-key.asc` | Embedded public key for signature verification (via `//go:embed`) |

### Modified (codebahn-cli)

| File | Change |
|---|---|
| `cmd/codebahn/main.go` | Add `update` command; hook periodic check into `PersistentPreRunE` |
| `go.mod` | Add `golang.org/x/mod` (for `semver.Compare`) |

### New (deploy repo: `codebahn/deploy`)

| File | Purpose |
|---|---|
| `terraform/modules/stack/releases.tf` | Scaleway Object Storage bucket (`codebahn-releases`), public-read ACL, IAM for CI uploads |
| `terraform/envs/production/dns.tf` | Add `releases` CNAME to bucket endpoint |

## Existing code to reuse

- **`internal/config.ConfigDir()`** / `XDG_CONFIG_HOME` pattern: copy for `CacheDir()` with `XDG_CACHE_HOME`.
- **`cmd/codebahn/main.go:28-33`**: version variable + ldflags pattern. `release.sh` sets `-X main.version=v0.x.x`.
- **`deploy/terraform/modules/stack/storage.tf`**: bucket + ACL + IAM pattern. The releases bucket follows the same structure minus versioning and encryption (public, static files).
- **Cobra command pattern** from `authCmd()` / `authLoginCmd()`: same shape for `updateCmd()`.

## Design

### Bucket layout

```
cli/latest.json
cli/v0.2.0/codebahn-linux-amd64
cli/v0.2.0/codebahn-linux-arm64
cli/v0.2.0/codebahn-darwin-amd64
cli/v0.2.0/codebahn-darwin-arm64
cli/v0.2.0/checksums.txt
cli/v0.2.0/checksums.txt.sig
```

### `latest.json`

```json
{"version": "0.2.0"}
```

### Periodic version check

- Runs in `PersistentPreRunE`, after auth but before command execution.
- Checks `~/.cache/codebahn/last-update-check` for a timestamp. Skips if <24h old.
- Fetches `https://releases.codebahn.net/cli/latest.json` (2s timeout, fail silently).
- If newer version exists, prints to stderr:
  `Update available: v0.2.0 (current: v0.1.0). Run "codebahn update" to upgrade.`
- Silenced by `CODEBAHN_NO_UPDATE_CHECK=1` env var or `"check_updates": false` in config.
- Skipped when `version == "dev"` (development builds).
- Skipped for the `update` command itself (it does its own check).

### `codebahn update` command

1. Fetch `latest.json`, determine target version.
2. If current == latest, print "Already up to date." and exit.
3. Detect Homebrew install (binary path contains `/Cellar/` or `/homebrew/`). Print
   "Installed via Homebrew. Run: brew upgrade codebahn" and exit.
4. Download binary for `runtime.GOOS`/`runtime.GOARCH`.
5. Download `checksums.txt` and `checksums.txt.sig`.
6. Verify GPG signature on `checksums.txt` using embedded public key.
7. Verify SHA256 of downloaded binary against `checksums.txt`.
8. Replace running binary: write to temp file in same directory, `os.Rename` (atomic on same filesystem).
9. Print "Updated to v0.2.0."

Flags:
- `--check`: only print whether an update is available, don't install.

### GPG signing

- Generate a dedicated release signing key: `Codebahn Releases <releases@codebahn.net>`.
- Private key stored as GitHub Actions secret (`GPG_PRIVATE_KEY`).
- Public key embedded in the binary via `//go:embed pgp/release-key.asc`.
- Public key also served at `https://codebahn.net/release-signing-key.asc` (static file via Caddy)
  for users who want to verify manually.
- Verification uses `golang.org/x/crypto/openpgp` (or its successor `github.com/ProtonMail/go-crypto`).

### `release.sh`

```
Usage: ./scripts/release.sh v0.2.0

1. Validate tag format (v\d+\.\d+\.\d+)
2. Cross-compile for 4 targets with ldflags
3. Generate checksums.txt (sha256sum)
4. GPG-sign checksums.txt
5. Upload all files + latest.json to Scaleway bucket via aws s3 cp
```

### GitHub Actions workflow

Trigger: push tag matching `v*`.
Steps:
1. Checkout code.
2. Set up Go.
3. Import GPG key from secret.
4. Run `./scripts/release.sh $TAG`.
5. (AWS CLI configured with Scaleway S3 credentials from secrets.)

Secrets needed: `SCW_ACCESS_KEY`, `SCW_SECRET_KEY`, `GPG_PRIVATE_KEY`.

### Terraform (releases bucket)

- Bucket: `codebahn-releases`, public-read ACL, no versioning, no encryption (public files).
- IAM: separate application + API key scoped to this bucket only (for CI uploads).
- DNS: `releases` CNAME record pointing to `codebahn-releases.s3.fr-par.scw.cloud`.
- Caddy: not involved; users hit the bucket directly via the CNAME.

## Out of scope

- **Windows builds**: not in the platform list.
- **Homebrew tap**: defer until there's demand.
- **Auto-update** (updating without user action): intentionally excluded.
- **Rollback**: if an update is bad, user re-runs `codebahn update` after a fix is released.
- **Delta updates**: always download the full binary.
- **Update channels** (stable/beta): single channel only.
- **Moving source code off GitHub**: separate decision.

## End-to-end check

1. `git tag v0.0.1 && git push --tags` triggers GitHub Actions.
2. Workflow cross-compiles, signs, uploads to `releases.codebahn.net`.
3. `curl https://releases.codebahn.net/cli/latest.json` returns `{"version":"0.0.1"}`.
4. Build a local binary with `go build -ldflags '-X main.version=v0.0.0' -o codebahn ./cmd/codebahn`.
5. `./codebahn repo list` prints update notice to stderr.
6. `./codebahn update --check` prints "Update available: v0.0.1 (current: v0.0.0)."
7. `./codebahn update` downloads, verifies signature + checksum, replaces binary.
8. `./codebahn --version` prints `v0.0.1`.
9. `./codebahn update` prints "Already up to date."
10. `CODEBAHN_NO_UPDATE_CHECK=1 ./codebahn repo list` prints no update notice.
