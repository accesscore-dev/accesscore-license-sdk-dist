# Go

## Taking it

```
go get github.com/accesscore-dev/accesscore-license-sdk-dist/go@v1.0.0
```

The core is linked statically, so a Go binary carries it rather than needing a shared library
installed beside it. cgo is required; `CGO_ENABLED=0` will not build.

## What a product has to supply

Interfaces in the package. Only the first two are required, along with the verification keys.

| Interface | What it answers |
| --- | --- |
| `Store` | reading and writing the registration |
| `ProductIdentity` | which product this is, and which release |
| `FeatureCatalogue` | the feature keys this deployment gates on |
| `QuotaCounter` | the current count per quota code |
| `FeatureActivationChecker` | whether an administrator switched a feature on |
| `AuditRecorder` | where license events are recorded |
| `ActivationServerDirectory` | where the activation server is |

## Assembling it

```go
manager := aclic.NewLicenseManager(
    store,
    identity,
    []aclic.KeyMaterial{{Kid: "acps-2026-01", PublicKey: publicKeyPEM}},
    aclic.WithCatalogue(catalogue),
    aclic.WithCounter(userCounter),
    aclic.WithAudit(auditRecorder),
    aclic.WithConfiguredServerURL(config.LicenseServerURL),
    aclic.WithInstanceID(config.InstanceID),
)

if _, err := manager.Verify(time.Now(), integrityOK); err != nil {
    return err
}
```

`Verify` takes the instant and the integrity result from the caller. The core reads no clock,
and what to hash for an integrity check differs per language.

## Enforcing it

```go
// Every request, while the deployment is unlicensed.
handler := aclic.LicenseGateMiddleware(manager.Gate(), []string{"/license", "/activations"})(mux)

// One feature.
protected := aclic.RequireFeatureMiddleware(manager.Gate(), "ATTENDANCE")(attendanceHandler)

// One ceiling, before adding.
if err := manager.CheckQuota("USERS"); err != nil {
    aclic.WriteDenial(writer, err)
    return
}
```

gin, echo, and fiber are reached through `GinLicenseGate`, `EchoLicenseGate`, and
`FiberLicenseGate`. Each takes a small interface their context already satisfies, so this module
depends on none of them and a product wires the one it uses in three lines.

Which paths stay open while unlicensed is the product's decision: a deployment has to reach its
own license screen in order to stop being unlicensed.

## Turning a refusal into a response

A refusal is a `*LicenseDenied` carrying `Status`, `ErrorCode`, `Denial`, and `HTTPStatus`.
`WriteDenial` writes the document every adapter writes; a product with its own response shape
reads the fields and renders its own:

```go
var denied *aclic.LicenseDenied
if errors.As(err, &denied) {
    message := translate(request.Context(), denied.Denial.MessageKey, denied.Denial.Arguments)
    respond(writer, denied.HTTPStatus, denied.ErrorCode, message)
    return
}
```

Message keys and their argument order: [messages.md](messages.md).

## The core directly

Anything the manager does not wrap is one call away:

```go
var answer map[string]any
if err := aclic.Shared.Release(map[string]any{
    "operation": "WITHIN_CEILING", "version": "3.2.0", "other": "3.2",
}, &answer); err != nil {
    return err
}
```
