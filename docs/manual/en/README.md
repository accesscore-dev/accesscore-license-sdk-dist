# AccessCore License SDK — manual

For putting the SDK into a product.

A Rust core signs licenses, verifies them, and decides what a deployment may do. Products
written in Java, Python, and Go take that decision through a thin binding. The core touches no
file, no clock, and no socket — your application supplies the values and the core returns a
conclusion. Because every language calls the same core, two deployments reading the same license
reach the same answer, down to the message key of a refusal.

![One core signs, the same core judges](images/architecture.png)

## The whole idea in three calls

Your product never parses a license. It asks three questions:

```
gate.state()                      is this deployment licensed right now, and until when
gate.requireFeature("ATTENDANCE") may it use this module — or refuse, with a message to show
gate.checkQuota("USERS", 199)     is there room for one more, or is the ceiling reached
```

The names follow each language: `requireFeature` in Java and Go, `require_feature` in Python.

Everything else — signatures, machine binding, expiry, grace periods, release ceilings,
heartbeats — happens behind those three, in one place, the same way in every language.

## Where to start

| | |
| --- | --- |
| [Concepts](concepts.md) | What the core decides, what your application decides, and the order the checks run in |
| [Java](java.md) | Taking it, the interfaces to implement, the gate, the framework adapters |
| [Python](python.md) | The same, in Python |
| [Go](go.md) | The same, in Go |
| [Message keys](messages.md) | Every refusal the SDK can return, and its argument order |

Read [concepts](concepts.md) first — it is short, and the guides assume it. Then the guide for
your language.

## A first hour

1. **Take the package** — the guide for your language has the coordinate and, for Java, the
   `read:packages` token the registry asks for.
2. **Implement two interfaces** — `Store` (where the registration lives) and `ProductIdentity`
   (which product this is, and which release). Everything else is optional.
3. **Ship the verification key with the build** — a classpath resource, a packaged file, a
   compiled-in constant — and construct the manager at startup. A deployment with no license
   comes up in `NOT_ACTIVATED` and refuses every gated call; that is correct, not a failure.
4. **Give an operator a license screen** and leave its path open in the gate. A deployment has
   to reach its own license screen in order to stop being unlicensed.
5. **Activate once** against a license server, and watch `state()` turn `VALID`.
6. **Add `requireFeature` and `checkQuota`** where the rules actually live.

## Taking it

The core travels inside each package, so nothing has to be installed beside your application.

| Language | Where it comes from | Token |
| --- | --- | --- |
| Java | `dev.accesscore:license-sdk`, from the GitHub Packages registry | `read:packages` — see [the Java guide](java.md) |
| Python | the wheel on [the release](https://github.com/accesscore-dev/accesscore-license-sdk-dist/releases/tag/v1.0.0) | none |
| Go | `github.com/accesscore-dev/accesscore-license-sdk-dist/packages/go` | none |

Every language is also attached to the release as a file — the jar, the wheel, the Go module,
and the prebuilt cores for binding the C ABI from a language the SDK does not cover. A build
that cannot reach a registry, or would rather not, takes the file.

## When something refuses

| What you see | Usually means |
| --- | --- |
| `NOT_ACTIVATED` | nothing is registered yet — activate with a product key |
| `FINGERPRINT_UNAVAILABLE` | a container without the host machine identifiers mounted |
| `MACHINE_MISMATCH` | the license belongs to a different host — move the seat on the server |
| `SCHEMA_UNSUPPORTED` | the license is newer than this build — upgrade, or have it reissued |
| `HEARTBEAT_OVERDUE` | the deployment has not checked in within the tolerated window |
| `RELEASE_NOT_ENTITLED` | the deployment was upgraded past what the contract covers |

Every status, what it means, and what an operator should do: [concepts](concepts.md) → Statuses.
