# Two repositories

The source and the release live apart.

| Repository | Visibility | What it holds |
| --- | --- | --- |
| `accesscore-dev/accesscore-license-sdk` | private | the Rust core, every binding's source, the conformance vectors, the guides |
| `accesscore-dev/accesscore-license-sdk-dist` | public | the prebuilt core, the C header, the Go binding, the guides, the vectors |

## Why a public repository exists at all

Go. `go get` resolves a module by fetching its source, so a Go consumer needs a tree they can
read. Java and Python resolve through package registries and never see either repository, and
Rust consumers inside the organization resolve from the private one.

Since a public tree has to exist for Go, it is the place everything consumer-facing goes: the
prebuilt libraries, the header for a language the SDK does not bind, the guides, and the
conformance vectors — a consumer verifying their own integration runs those, and a promise about
what a token means is not one they can hold anyone to if they cannot read it.

## What never crosses

The Rust core sources and the issuing paths stay private. Nothing in the public repository can
mint a license: signing needs a private key, and no private key has been in either repository.
The conformance keys travel because they sign nothing but the vectors.

## How a release moves

```
tag v1.2.3 on the private repository
  ├── build the core for six platforms          (one job per platform, in parallel)
  ├── package each binding around the same set  (so no two bindings ship different cores)
  ├── publish  dev.accesscore:license-sdk       → GitHub Packages
  ├── publish  accesscore-license               → the wheel, attached to the release
  ├── assemble the consumer tree                 scripts/assemble-dist.sh
  ├── push it to the public repository           a fresh commit, tagged v1.2.3
  └── create the release, platforms attached
```

The public repository carries releases, not the development that produced them: each release is
one commit rather than a transplanted history. The Go module path is rewritten during assembly,
so a consumer's `go.mod` names the public repository and never the private one.

## What CI needs

| Secret | Where | What for |
| --- | --- | --- |
| `GITHUB_TOKEN` | provided | publishing to GitHub Packages, creating the release |
| `DIST_REPO_TOKEN` | set by hand | write access to the public repository, and to nothing else |

`DIST_REPO_TOKEN` is scoped to one repository on purpose. A release that can write anywhere is
a release step that can be turned into something else.

## Version numbers

One version across all four bindings and the core, taken from the tag. A binding that could
release on its own would let a consumer pair a Java wrapper with a core that reads a different
schema, and the schema number is exactly the thing that must not disagree.
