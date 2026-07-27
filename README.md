# AccessCore License SDK

Release **1.0.0**. A Rust core signs licenses, verifies them, and decides what a deployment
may do; Java, Python, and Go products take that decision through a thin binding. The core
travels inside each package, so nothing has to be installed beside your application.

## Java

Published to the GitHub Packages registry of this repository. The registry asks every consumer
to authenticate, public package or not, so the repository declaration carries a GitHub username
and a token with the `read:packages` scope — a classic token from
<https://github.com/settings/tokens/new?scopes=read:packages>, kept out of the build file.

```groovy
repositories {
    maven {
        url = uri('https://maven.pkg.github.com/accesscore-dev/accesscore-license-sdk-dist')
        credentials {
            username = providers.gradleProperty('gpr.user').orElse(providers.environmentVariable('GITHUB_USERNAME')).get()
            password = providers.gradleProperty('gpr.token').orElse(providers.environmentVariable('GITHUB_TOKEN')).get()
        }
        content { includeGroup 'dev.accesscore' }
    }
}
dependencies {
    implementation 'dev.accesscore:license-sdk:1.0.0'
}
```

Put `gpr.user` and `gpr.token` in `~/.gradle/gradle.properties`, or set `GITHUB_USERNAME`
and `GITHUB_TOKEN` in the environment. The full walkthrough, Maven included, is in
[the manual](docs/manual/en/java.md).

## Python

No token — the wheel is attached to [the release](https://github.com/accesscore-dev/accesscore-license-sdk-dist/releases/tag/v1.0.0).

```
pip install https://github.com/accesscore-dev/accesscore-license-sdk-dist/releases/download/v1.0.0/accesscore_license-1.0.0-py3-none-any.whl
```

## Go

No token.

```
go get github.com/accesscore-dev/accesscore-license-sdk-dist/packages/go@v1.0.0
```

## Or take the file

Every language is attached to [the release](https://github.com/accesscore-dev/accesscore-license-sdk-dist/releases/tag/v1.0.0): the jar, the wheel, the Go module, and the prebuilt cores for binding the C ABI directly.

## What is here

| Path | What it is |
| --- | --- |
| `packages/go` | The Go module, with the static core it links |
| `native/include` | The C ABI, for binding a language the SDK does not |
| `native/lib` | The prebuilt core, one directory per platform |
| `docs/manual/en`, `docs/manual/ko` | The manual, one folder per language |
| `conformance/vectors` | The cases every binding answers the same way |

## Documents

| | |
| --- | --- |
| **English** | [docs/manual/en/README.md](docs/manual/en/README.md) |
| **한국어** | [docs/manual/ko/README.md](docs/manual/ko/README.md) |

Start with `concepts` in either folder — what the core decides and what your application
decides — then the guide for your language.

## What is not here

The Rust core, the signing paths, and the issuing side. A deployment verifies licenses; only an
issuer signs them, and signing needs a private key that has never been in any repository.
