# Conformance keys

The fixed key pairs the conformance vectors are signed with. Four language bindings running the
same vectors have to read the same token signed by the same key, so these are values held once
rather than generated per run.

| `kid` | Algorithm |
| --- | --- |
| `acps-2026-01` | Ed25519 |
| `acps-2026-02` | RSA-4096 PSS SHA-256 |
| `acps-2027-01` | Ed25519 — the incoming key in the rotation scenario |

Every case carries the public keys it is verified against inline, so a runner needs nothing
beside the vector files.

**No license is ever issued with these keys.** The private halves stay in the SDK's own
repository, where they sign the vectors and nothing else. An issuing key pair is created with
`aclic_generate_keypair` and never leaves the issuing environment.
