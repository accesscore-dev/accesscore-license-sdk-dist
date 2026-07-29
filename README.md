# AccessCore License SDK

Release **1.0.2**. A Rust core signs licenses, verifies them, and decides what a deployment
may do; Java, Python, and Go products take that decision through a thin binding. The core
travels inside each package, so nothing has to be installed beside your application.

![One core signs, the same core judges](docs/manual/en/images/architecture.png)

## The whole idea in three calls

Your product never parses a license. It asks three questions:

```
gate.state()                      is this deployment licensed right now, and until when
gate.requireFeature("ATTENDANCE") may it use this module — or refuse, with a message to show
gate.checkQuota("USERS", 199)     is there room for one more, or is the ceiling reached
```

Everything else — signatures, machine binding, expiry, grace periods, release ceilings,
heartbeats — happens behind those three, in one place, the same way in every language. Read
[concepts](docs/manual/en/concepts.md) for what each check does and the order they run in.

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
    implementation 'dev.accesscore:license-sdk:1.0.2'
}
```

Put `gpr.user` and `gpr.token` in `~/.gradle/gradle.properties`, or set `GITHUB_USERNAME`
and `GITHUB_TOKEN` in the environment. The full walkthrough, Maven included, is in
[the manual](docs/manual/en/java.md).

## Python

No token — the wheel is attached to [the release](https://github.com/accesscore-dev/accesscore-license-sdk-dist/releases/tag/v1.0.2).

```
pip install https://github.com/accesscore-dev/accesscore-license-sdk-dist/releases/download/v1.0.2/accesscore_license-1.0.2-py3-none-any.whl
```

## Go

No token.

```
go get github.com/accesscore-dev/accesscore-license-sdk-dist/packages/go@v1.0.2
```

## Or take the file

Every language is attached to [the release](https://github.com/accesscore-dev/accesscore-license-sdk-dist/releases/tag/v1.0.2): the jar, the wheel, the Go module, and the prebuilt cores for binding the C ABI directly.

## What is here

| Path | What it is |
| --- | --- |
| `packages/go` | The Go module, with the static core it links |
| `native/include` | The C ABI, for binding a language the SDK does not |
| `native/lib` | The prebuilt core, one directory per platform |
| `docs/manual/en`, `docs/manual/ko` | The manual, one folder per language |
| `conformance/vectors` | The cases every binding answers the same way |

## Documents

| | English | 한국어 |
| --- | --- | --- |
| Start here | [manual](docs/manual/en/README.md) | [매뉴얼](docs/manual/ko/README.md) |
| What the core decides, and in what order | [concepts](docs/manual/en/concepts.md) | [개념](docs/manual/ko/concepts.md) |
| Java | [java](docs/manual/en/java.md) | [java](docs/manual/ko/java.md) |
| Python | [python](docs/manual/en/python.md) | [python](docs/manual/ko/python.md) |
| Go | [go](docs/manual/en/go.md) | [go](docs/manual/ko/go.md) |
| Every refusal, for your translations | [messages](docs/manual/en/messages.md) | [메시지 키](docs/manual/ko/messages.md) |

Start with concepts — it is short, and every language guide assumes it.

## When something refuses

| What you see | Usually means |
| --- | --- |
| `NOT_ACTIVATED` | nothing is registered yet — activate with a product key |
| `FINGERPRINT_UNAVAILABLE` | a container without the host machine identifiers mounted |
| `MACHINE_MISMATCH` | the license belongs to a different host — move the seat on the server |
| `SCHEMA_UNSUPPORTED` | the license is newer than this build — upgrade, or have it reissued |
| `HEARTBEAT_OVERDUE` | the deployment has not checked in within the tolerated window |
| `RELEASE_NOT_ENTITLED` | the deployment was upgraded past what the contract covers |

## What is not here

The Rust core, the signing paths, and the issuing side. A deployment verifies licenses; only an
issuer signs them, and signing needs a private key that has never been in any repository.
