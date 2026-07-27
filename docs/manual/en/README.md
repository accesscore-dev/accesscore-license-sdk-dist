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

Everything below is public — no account, no token. The core travels inside each package, so
nothing has to be installed beside your application.

**Java**

```groovy
repositories {
    maven { url 'https://accesscore-dev.github.io/accesscore-license-sdk-dist/packages/maven' }
}
dependencies {
    implementation 'dev.accesscore:license-sdk:1.0.0'
}
```

**Python**

```
pip install --extra-index-url https://accesscore-dev.github.io/accesscore-license-sdk-dist/packages/pypi/simple \
    accesscore-license==1.0.0
```

**Go**

```
go get github.com/accesscore-dev/accesscore-license-sdk-dist/packages/go@v1.0.0
```

| Language | Coordinate | How the core travels |
| --- | --- | --- |
| Java | `dev.accesscore:license-sdk` | inside the jar, extracted at run time |
| Python | `accesscore-license` | inside the wheel |
| Go | `github.com/accesscore-dev/accesscore-license-sdk-dist/packages/go` | linked statically |
