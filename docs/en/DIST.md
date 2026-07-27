# AccessCore License SDK — distribution

Release **1.0.0**. This repository carries what a consumer of the SDK needs. The core lives
in a private repository; nothing here can mint a license.

## What is here

| Path | What it is |
| --- | --- |
| `go/` | The Go binding, with the static core it links |
| `include/aclic.h` | The C ABI, for binding a language the SDK does not |
| `native/` | The prebuilt core, one directory per platform |
| `docs/en`, `docs/ko` | Every document, one folder per language |
| `conformance/` | The vectors every binding answers the same way |

## Taking it

**Java** — `dev.accesscore:license-sdk:1.0.0` from GitHub Packages.

**Python** — `accesscore-license==1.0.0`.

**Go** — `go get github.com/accesscore-dev/accesscore-license-sdk-dist/go@v1.0.0`.

**Rust** — `aclic = "1.0.0"`.

## What is not here

The Rust core, the signing paths, and the issuing side. A deployment verifies tokens; only an
issuer signs them, and signing needs a private key that has never been in any repository.
