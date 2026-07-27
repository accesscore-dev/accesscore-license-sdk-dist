# Concepts

One core owns the license format and the judgement. Four bindings hand it what a host knows and
read back what the host may do. Nothing about a license is decided twice.

## What the core owns, and what the host owns

| The core | The host |
| --- | --- |
| The payload format and the schema number | Storing and loading the record (JPA, a file, a registry) |
| Signing and verifying | The current instant |
| Collecting machine identifiers, and matching them | HTTP transport, retries, timeouts |
| Expiry, the grace period, per-channel relaxations | Running the heartbeat on a schedule |
| The heartbeat deadline | Where the gate is attached |
| Deciding a clock was rolled back | Hashing its own binary |
| Product code and release-ceiling checks | Recording audit rows |
| Refusal statuses, and their message keys and arguments | Translating the messages |
| Writing activation and heartbeat bodies, and reading the answers | |

Binary integrity is the host's alone because what to hash differs per language: a jar hash, an
executable hash, and a wheel hash are not obtained the same way. The host measures and passes
the result in, and the core places it in the same order as every other check.

## The token

```
base64url(payload).base64url(signature)
```

The signature covers the payload bytes as transmitted, and those exact bytes travel in the
token. A verifier never re-serializes anything, so the format needs no canonical JSON form and a
language whose JSON writer orders keys differently cannot break a signature.

## The judgement

Fifteen checks in one order:

```
 1. binary integrity (the host's result)   failed   → INTEGRITY_FAILED
 2. a stored token                         none     → NOT_ACTIVATED
 3. a revocation stamp                     present  → REVOKED
 4. the token parses                       no       → NOT_LOADED
 5. the schema number                      unread   → SCHEMA_UNSUPPORTED
 6. allow list · key lookup · pair check    off      → INVALID_SIGNATURE
 7. the signature                          fails    → INVALID_SIGNATURE
 8. the product code                        differs  → PRODUCT_MISMATCH
 9. channel and wildcard agree             no       → INVALID_SIGNATURE
10. machine fingerprints                   differ   → MACHINE_MISMATCH
                                           unread   → FINGERPRINT_UNAVAILABLE
11. expiry and grace                                → VALID / GRACE_PERIOD / EXPIRED
12. the release ceiling                    above    → RELEASE_NOT_ENTITLED
13. the clock watermark                    behind   → CLOCK_TAMPERED
14. the heartbeat deadline                 passed   → HEARTBEAT_OVERDUE
15. the watermark advances                          → max(now, the one held)
```

The schema number is read before the signature: a format this deployment cannot read is not
worth computing a signature over. The product code sits right after the signature and before the
machine binding — if the token belongs to another product, whether the machine matches is not a
question worth asking. Steps 13 and 14 are only reached when step 12 left a usable status.

## Statuses

| Status | Meaning | Usable |
| --- | --- | --- |
| `VALID` | every check passed | ● |
| `GRACE_PERIOD` | expired, inside the grace period | ● |
| `RELEASE_NOT_ENTITLED` | above the contract ceiling | ● |
| `EXPIRED` | past the grace period | ○ |
| `INVALID_SIGNATURE` | signature, algorithm, or a production wildcard | ○ |
| `MACHINE_MISMATCH` | stamped for other machines | ○ |
| `FINGERPRINT_UNAVAILABLE` | not one identifier could be read | ○ |
| `NOT_ACTIVATED` | nothing registered | ○ |
| `NOT_LOADED` | a token that cannot be read | ○ |
| `SCHEMA_UNSUPPORTED` | a schema this build does not read | ○ |
| `REVOKED` | the server revoked it | ○ |
| `CLOCK_TAMPERED` | the clock moved back past the watermark | ○ |
| `INTEGRITY_FAILED` | the binary is not what it should be | ○ |
| `HEARTBEAT_OVERDUE` | a required refresh did not arrive | ○ |
| `PRODUCT_MISMATCH` | another product's license | ○ |

`RELEASE_NOT_ENTITLED` stays usable on purpose. A ceiling that closed the state gate would stop
access control at the door; what it costs instead is the feature list, which comes back empty.

## Machine binding

Each hardware identifier is hashed on its own into `sha256:<hex>`, and the list travels in the
payload.

| Platform | What is read |
| --- | --- |
| Linux | `/sys/class/dmi/id/product_uuid`, `product_serial`, `/etc/machine-id` |
| macOS | `IOPlatformUUID` |
| Windows | `MachineGuid` under `HKLM\SOFTWARE\Microsoft\Cryptography` |

Matching is by intersection, and the threshold is measured against what was **collected**:

```
matched = |collected ∩ stamped|
passes when  matched ≥ 1  and  matched ≥ ceil(|collected| / 2)
```

| Situation | Collected | Matched | Verdict |
| --- | --- | --- | --- |
| Redeployed without the DMI mount | 1 | 1 | pass |
| Mainboard replaced | 3 | 2 | pass |
| Only the machine-id copied elsewhere | 3 | 1 | refuse |
| Nothing readable | 0 | 0 | `FINGERPRINT_UNAVAILABLE` |

Measuring against the stamped count instead would leave an environment that reads fewer
identifiers no way to pass at all. MAC addresses are not used: `--mac-address` sets one to
whatever a caller likes.

In a container the Linux identifiers have to be mounted read-only from the host:

```
-v /sys/class/dmi/id:/sys/class/dmi/id:ro -v /etc/machine-id:/etc/machine-id:ro
```

## Signatures and keys

Issuance uses Ed25519 — a 64-byte signature and a 32-byte public key keep a hand-carried token
short and make embedding the verification key cost nothing. Verification also reads RSA-4096
PSS, so a customer who must keep the private key in a hardware module is served without touching
a deployed verifier.

`alg` names the algorithm and `kid` names the key. Both sit under the signature, so changing
either breaks it — but only if the checks run in this order:

```
1. read alg and kid from the payload
2. is alg on this deployment's allow list?   no → INVALID_SIGNATURE
3. find the key named by kid                 none → INVALID_SIGNATURE
4. is that key's algorithm the declared one? no → INVALID_SIGNATURE
5. check the signature
```

Steps 2 and 4 are what close algorithm confusion.

A `kid` is `<family>-<year>-<serial>`, for example `acps-2026-01` — an ordered name an operator
reads against the ledger, not a random one. A deployment may carry several keys at once, which
is what makes a rotation possible: during one, the old and the new key ship together and the
token's `kid` decides which applies.

The verification key is fixed in the deployment's binary. A path a deployment could redirect
would let a customer verify a license they minted themselves, and at that moment the signature
check defends nothing. Rotation is a rebuild.

## The gate

```
gate.state()                     the cached judgement
gate.require_feature(key)        refuse when the license does not grant it
gate.check_quota(code, current)  refuse when the ceiling is passed
gate.refresh(evaluation)         replace the cache with a fresh judgement
```

The binding holds the judgement; the host decides when to refresh it — at start, after a
successful activation, after a successful heartbeat, and on whatever schedule it keeps. The gate
that runs on every request reads the cache and never calls the core, which is why the JSON
boundary costs nothing at request time.

A quota with no entry in the payload is unlimited; an entry of zero is zero. A license issued
before a quota code existed must not read as forbidding everything under it.

## Activation

```
host                                   core
product key typed
  ├── fingerprint() ────────────────▶ collect and hash identifiers
  ├── activation_request() ─────────▶ the body, and a nonce
  ◀── {body, nonce}
  ├── HTTP POST (the host sends it)
  ◀── the server's answer
  ├── activation_response() ────────▶ nonce, machine, and payload checked
  ◀── {licenseToken, activationSecret, …}
  └── stored (the host)
```

First activation authenticates with the product key alone. Every later request is signed with
the activation secret, in `X-Activation-Signature` as base64 HMAC-SHA256 over the body bytes as
sent. A closed network runs the same functions with a file where the socket was.

## What the SDK does not decide

- **The ledger, the catalogue, the back office** — the issuing server owns them. The SDK knows
  only the shape of what was issued.
- **Product vocabulary** — feature keys and quota codes are a product's constants. The SDK
  carries them as free strings.
- **Wording** — the SDK returns message keys and arguments; the host translates them.
- **Binary integrity** — the host measures, the SDK places the answer in order.
