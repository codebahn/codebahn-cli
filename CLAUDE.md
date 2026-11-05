@AGENTS.md

Claude Code specifics:
- Auto-memory for this repo is redirected to `~/code/agent-knowledge/inbox/codebahn-cli` by `.claude/settings.json`. It is unverified scratch; promote anything worth keeping through the knowledge repo.
- `.claude/rules/security/*.md` are generated ASVS rules with a `paths:` frontmatter; they load when a matching file is read or edited. Change `.claude/asvs.map.yaml` and run `.claude/render-rules`; never edit the rendered files.
- Project settings deny reading `.env*` and the CLI's own `config.json` (it holds the user's OAuth tokens). Never read it, even when debugging auth.
