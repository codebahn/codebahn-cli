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

```bash
go install github.com/codebahn/codebahn-cli/cmd/codebahn@latest
```

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

## License

MIT
