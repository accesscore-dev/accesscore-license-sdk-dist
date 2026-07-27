# Rust

## Taking it

```toml
[dependencies]
aclic = { version = "1.0.0", features = ["axum"] }
```

Rust reaches the core directly rather than through the C ABI — it is the language the core is
written in, and a native boundary between a crate and itself would buy nothing. What it shares
with every other binding is the layer above: the same neutral gate, the same host interfaces,
the same adapters.

Adapter features: `axum`, `actix-web`, `tower`. A product pays only for the framework it uses.

## What a product has to supply

Traits in `aclic::host`. Only the first two are required, along with the verification keys.

| Trait | What it answers |
| --- | --- |
| `Store` | reading and writing the registration |
| `ProductIdentity` | which product this is, and which release |
| `FeatureCatalogue` | the feature keys this deployment gates on |
| `QuotaCounter` | the current count per quota code |
| `FeatureActivationChecker` | whether an administrator switched a feature on |
| `AuditRecorder` | where license events are recorded |
| `ActivationServerDirectory` | where the activation server is |

## Assembling it

```rust
use aclic::{Gate, KeyMaterial, LicenseManager};
use std::sync::Arc;

let gate = Arc::new(Gate::new());
let manager = LicenseManager::new(
        Arc::clone(&gate),
        store,
        identity,
        vec![KeyMaterial {
            kid: "acps-2026-01".into(),
            alg: None,
            public_key: PUBLIC_KEY_PEM.into(),
        }],
    )
    .with_catalogue(catalogue)
    .with_counter(user_counter)
    .with_audit(audit_recorder)
    .with_configured_server_url(&config.license_server_url)
    .with_instance_id(&config.instance_id);

manager.verify(OffsetDateTime::now_utc(), integrity_ok);
```

`verify` takes the instant and the integrity result from the caller. The core reads no clock,
and what to hash for an integrity check differs per language.

## Enforcing it

```rust
// axum: every request, while the deployment is unlicensed.
let app = Router::new()
    .merge(protected_routes)
    .layer(middleware::from_fn_with_state(
        Arc::clone(&gate),
        |State(gate): State<Arc<Gate>>, request, next| async move {
            aclic::adapter::axum::license_gate(gate, request, next).await
        },
    ))
    .merge(license_screen_routes);   // reachable while unlicensed

// One feature, inside a handler.
gate.require_feature("ATTENDANCE")?;

// One ceiling, before adding.
gate.check_quota("USERS", current)?;
```

`GateError` implements `IntoResponse` under the `axum` feature and `ResponseError` under
`actix-web`, so `?` in a handler already answers with the shared document. tower offers
`LicenseGateLayer` for any tower-based stack.

Which paths stay open while unlicensed is the product's decision: a deployment has to reach its
own license screen in order to stop being unlicensed, so those routes are merged outside the
layer rather than listed inside the SDK.

## Turning a refusal into a response

`GateError` carries `status`, `error_code`, `denial`, and `http_status`. A product with its own
response shape reads them:

```rust
let message = translate(locale, &error.denial.message_key, &error.denial.arguments);
(StatusCode::from_u16(error.http_status)?, Json(MyEnvelope::error(&error.error_code, message)))
```

Message keys and their argument order: [messages.md](messages.md).

## The core crates directly

A product needing something the binding does not wrap takes the crate it belongs to:
`aclic_model` for the format, `aclic_sign` for keys and signatures, `aclic_verify` for the
judgement, `aclic_fingerprint` for machine identifiers, `aclic_protocol` for activation bodies.
