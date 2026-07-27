# AccessCore License SDK — manual

For putting the SDK into a product.

A Rust core signs licenses, verifies them, and decides what a deployment may do. Products
written in Java, Python, and Go take that decision through a thin binding. The core touches no
file, no clock, and no socket — your application supplies the values and the core returns a
conclusion. Because every language calls the same core, two deployments reading the same license
reach the same answer, down to the message key of a refusal.

| | |
| --- | --- |
| [Concepts](concepts.md) | What the core decides, what your application decides, and the order the checks run in |
| [Java](java.md) | Taking it, the interfaces to implement, the gate, the framework adapters |
| [Python](python.md) | The same, in Python |
| [Go](go.md) | The same, in Go |
| [Message keys](messages.md) | Every refusal the SDK can return, and its argument order |

Start with concepts, then the guide for your language.

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
