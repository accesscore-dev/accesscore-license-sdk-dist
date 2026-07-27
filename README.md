# AccessCore License SDK

Release **1.0.0**. A Rust core signs licenses, verifies them, and decides what a deployment
may do; Java, Python, and Go products take that decision through a thin binding. The core
travels inside each package, so nothing has to be installed beside your application.

Everything below is public — no account, no token.

## Java

```groovy
repositories {
    maven { url 'https://accesscore-dev.github.io/accesscore-license-sdk-dist/packages/maven' }
}
dependencies {
    implementation 'dev.accesscore:license-sdk:1.0.0'
}
```

## Python

```
pip install --extra-index-url https://accesscore-dev.github.io/accesscore-license-sdk-dist/packages/pypi/simple accesscore-license==1.0.0
```

## Go

```
go get github.com/accesscore-dev/accesscore-license-sdk-dist/packages/go@v1.0.0
```

## What is here

| Path | What it is |
| --- | --- |
| `packages/maven` | The Java artifact, as a Maven repository |
| `packages/pypi` | The Python wheel, as a simple index |
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
