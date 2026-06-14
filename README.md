# kr36

Browse 36kr Chinese tech news via RSS (36氪)

`kr36` is a single pure-Go binary. It speaks to kr36-cli over plain
HTTPS, shapes the responses into clean records, and pipes into the rest of your
tools. No API key, nothing to run alongside it.

## Install

```bash
go install github.com/tamnd/kr36-cli-cli/cmd/kr36@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/kr36-cli-cli/releases), or run
the container image:

```bash
docker run --rm ghcr.io/tamnd/kr36:latest --help
```

## Usage

```bash
kr36 --help
kr36 version
```

This is a fresh scaffold. The command tree starts with `version`; build out the
real commands in `cli/` on top of the `kr36-cli` library package.

## Development

```
cmd/kr36/   thin main, wires cli.Root into fang
cli/                 the cobra command tree
kr36-cli/                the library: HTTP client and data models
docs/                tago documentation site
```

```bash
make build      # ./bin/kr36
make test       # go test ./...
make vet        # go vet ./...
```

## Releasing

Push a version tag and GitHub Actions runs GoReleaser, which builds the
archives, Linux packages, the multi-arch GHCR image, checksums, SBOMs, and a
cosign signature:

```bash
git tag v0.1.0
git push --tags
```

The Homebrew and Scoop steps self-disable until their tokens exist, so the first
release works with no extra secrets.

## License

Apache-2.0. See [LICENSE](LICENSE).
