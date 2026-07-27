# Message keys

The core returns a key and its arguments; the host translates. Every language the SDK binds has
its own internationalization, and a core holding wording would grow a phrase book each time a
language is added.

A host must define every key below in each locale it serves. A missing key surfaces as an
untranslated placeholder at the worst moment — the operator is already locked out.

## Where a message comes from

```mermaid
sequenceDiagram
    autonumber
    participant App as Your application
    participant Gate as The gate
    participant I18n as Your message bundle
    actor User

    App->>Gate: requireFeature("ATTENDANCE")
    Gate-->>App: refusal { errorCode, messageKey, arguments, httpStatus }
    Note over App: the SDK stops here —<br/>it never renders a sentence
    App->>I18n: lookup(messageKey, arguments, locale)
    I18n-->>App: "이 배포에는 ATTENDANCE 모듈이 포함되어 있지 않습니다."
    App-->>User: 403 + the rendered message
```

The `errorCode` is for your client code to branch on; the `messageKey` and its arguments are for
a human. Branching on the rendered sentence breaks the moment someone changes a translation.

## Argument order is part of the key

The order is fixed here and nowhere else. A host that swaps two arguments renders a number in the
wrong place and tells an operator something untrue.

## Denials a judgement produces

| Key | Arguments | When |
| --- | --- | --- |
| `error.license.notValid` | — | the general line, used when a specific one cannot be filled in |
| `error.license.gracePeriodReadOnly` | — | `GRACE_PERIOD` — expired, still serving |
| `error.license.releaseNotEntitled` | 1. running release<br>2. entitled ceiling | `RELEASE_NOT_ENTITLED` |
| `error.license.expired` | 1. expiry instant, RFC 3339 | `EXPIRED` |
| `error.license.schemaUnsupported` | 1. the license's schema<br>2. this build's schema | `SCHEMA_UNSUPPORTED` |
| `error.license.heartbeatOverdue` | 1. tolerated days | `HEARTBEAT_OVERDUE` |
| `error.license.productMismatch` | 1. the license's product<br>2. this deployment's product | `PRODUCT_MISMATCH` |
| `error.license.invalidSignature` | — | `INVALID_SIGNATURE` |
| `error.license.fingerprintMismatch` | — | `MACHINE_MISMATCH` |
| `error.license.fingerprintUnavailable` | — | `FINGERPRINT_UNAVAILABLE` |
| `error.license.notActivated` | — | `NOT_ACTIVATED` |
| `error.license.unreadable` | — | `NOT_LOADED` |
| `error.license.revoked` | — | `REVOKED` |
| `error.license.clockTampered` | — | `CLOCK_TAMPERED` |
| `error.license.integrityFailed` | — | `INTEGRITY_FAILED` |

A denial that a number makes actionable carries it. When the payload could not be read far
enough to fill one in, the denial falls back to `error.license.notValid` rather than rendering an
empty placeholder.

## Denials a gate produces

| Key | Arguments | When |
| --- | --- | --- |
| `error.license.featureNotLicensed` | 1. feature key | the license does not grant the feature |
| `error.license.featureNotActivated` | 1. feature key | licensed, but an administrator has not switched it on |
| `error.license.quotaExceeded` | 1. quota code<br>2. the ceiling | one more would pass the ceiling |

## Reasons an activation attempt returns

| Key | When |
| --- | --- |
| `error.license.activationServerNotConfigured` | no server address to reach |
| `error.license.fingerprintUnavailable` | no machine identifier could be read |
| `error.license.unknownProductKey` | the server does not know the key |
| `error.license.seatExhausted` | every seat is held by another machine |
| `error.license.contractSuspended` | the contract behind the license is suspended |
| `error.license.productMismatch` | the key sells another product |
| `error.license.releaseNotEntitled` | the running release is above the ceiling |
| `error.license.revoked` | the activation was revoked |
| `error.license.serverUnreachable` | the server could not be reached |
| `error.license.serverError` | the server answered with an error |
| `error.license.invalidResponse` | the answer was not an activation response |
| `error.license.nonceMismatch` | the returned token echoes another nonce |
| `error.license.fingerprintMismatch` | the returned token is bound elsewhere |
| `error.license.invalidSignature` | the returned token is signed by another key |
| `error.license.notActivated` | nothing to refresh or release |
| `error.license.alreadyRegistered` | a working license under another key is registered here |

The four ways a token is refused after it arrives — a wrong signing key, a license bound to
another machine, a replayed answer, an unreadable one — are named apart on purpose. They need
four different actions from an operator, and one shared "could not be verified" line leaves every
one of them undiagnosable.

## Error codes

Every refusal also carries a code a client branches on, and every one begins with `LICENSE_`.
That prefix is how a client tells a license denial from a permission denial without parsing a
message.

- A judgement's code is `LICENSE_` plus the status name: `LICENSE_HEARTBEAT_OVERDUE`.
- A gate's codes are `LICENSE_FEATURE_NOT_LICENSED`, `LICENSE_FEATURE_NOT_ACTIVATED`, and
  `LICENSE_QUOTA_EXCEEDED`.

## The body every adapter writes

```json
{
  "errorCode": "LICENSE_NOT_ACTIVATED",
  "status": "NOT_ACTIVATED",
  "messageKey": "error.license.notActivated",
  "arguments": []
}
```

One shape, whichever framework produced it. The adapter contract tests assert exactly this, which
is what makes a new adapter cheap: it moves the gate's answer into a response and decides nothing.
