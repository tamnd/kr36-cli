---
title: "Installation"
description: "Install kr36 from a release, with go install, or from source."
weight: 20
---

## Prebuilt binaries

Every [release](https://github.com/tamnd/kr36-cli-cli/releases) carries archives for Linux, macOS,
and Windows on amd64 and arm64, plus deb, rpm, and apk packages for Linux.
Download, unpack, put `kr36` on your `PATH`, done. The `checksums.txt`
on each release is signed with keyless [cosign](https://docs.sigstore.dev/) if
you want to verify before running.

## With Go

```bash
go install github.com/tamnd/kr36-cli-cli/cmd/kr36@latest
```

That puts `kr36` in `$(go env GOPATH)/bin`, which is `~/go/bin` unless
you moved it. Make sure that directory is on your `PATH`.

## From source

```bash
git clone https://github.com/tamnd/kr36-cli-cli
cd kr36-cli-cli
make build        # produces ./bin/kr36
./bin/kr36 version
```

## Container image

```bash
docker run --rm ghcr.io/tamnd/kr36:latest --help
```

## Checking the install

```bash
kr36 version
```

prints the version and exits.
