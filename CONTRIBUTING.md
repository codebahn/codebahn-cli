# Contributing

Thanks for your interest in contributing to the Codebahn CLI.

## Reporting issues

Open an issue on [GitHub](https://github.com/codebahn/codebahn-cli/issues). Include the output of `codebahn --version` and the command that failed.

## Development

```bash
git clone https://github.com/codebahn/codebahn-cli.git
cd codebahn-cli
go build -o codebahn ./cmd/codebahn
go test ./...
```

The `tools/` directory is a separate Go module with zero dependencies. Run its tests separately:

```bash
cd tools && go test ./...
```

## Code generation

The cobra commands in `internal/gen/` are generated from the tool definitions in `tools/`. After changing a tool definition, regenerate and commit the output:

```bash
go run ./cmd/gen
```

## Pull requests

Keep PRs small and focused. One change per PR. Tests are expected for new functionality.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
