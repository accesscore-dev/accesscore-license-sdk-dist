# Rust

## 가져다 쓰기

```toml
[dependencies]
aclic = { version = "1.0.0", features = ["axum"] }
```

Rust는 C ABI를 거치지 않고 코어를 직접 쓴다 — 코어가 쓰인 언어이고, 크레이트와 자기 자신 사이에 네이티브 경계를 두어 얻을 것이 없다. 다른 바인딩과 공유하는 것은 그 위 계층이다: 같은 중립 게이트, 같은 호스트 인터페이스, 같은 어댑터.

어댑터 기능 플래그: `axum`, `actix-web`, `tower`. 제품은 쓰는 프레임워크 값만 치른다.

## 제품이 제공해야 하는 것

`aclic::host`의 트레이트다. 검증 키와 함께 앞의 둘만 필수다.

| 트레이트 | 답하는 것 |
| --- | --- |
| `Store` | 등록 기록 읽기와 쓰기 |
| `ProductIdentity` | 이 배포가 어떤 제품이고 어떤 릴리스인지 |
| `FeatureCatalogue` | 이 배포가 게이트하는 기능 키 |
| `QuotaCounter` | 한도 코드별 현재 수량 |
| `FeatureActivationChecker` | 운영자가 그 기능을 켰는지 |
| `AuditRecorder` | 라이선스 이벤트 기록 위치 |
| `ActivationServerDirectory` | 활성화 서버 주소 |

## 조립

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

`verify`는 현재 시각과 무결성 검사 결과를 호출자에게서 받는다. 코어는 시계를 읽지 않고, 무결성 검사 대상은 언어마다 다르다.

## 강제

```rust
// axum: 라이선스가 없는 동안 모든 요청을 막는다.
let app = Router::new()
    .merge(protected_routes)
    .layer(middleware::from_fn_with_state(
        Arc::clone(&gate),
        |State(gate): State<Arc<Gate>>, request, next| async move {
            aclic::adapter::axum::license_gate(gate, request, next).await
        },
    ))
    .merge(license_screen_routes);   // 라이선스가 없어도 닿아야 하는 경로

// 기능 하나, 핸들러 안에서.
gate.require_feature("ATTENDANCE")?;

// 한도 하나, 추가 직전에.
gate.check_quota("USERS", current)?;
```

`GateError`는 `axum` 기능에서 `IntoResponse`를, `actix-web`에서 `ResponseError`를 구현하므로 핸들러의 `?` 하나로 공통 문서가 나간다. tower 기반 스택에는 `LicenseGateLayer`가 있다.

라이선스가 없는 동안 어느 경로를 열어 둘지는 제품이 정한다. 배포물은 자기 라이선스 화면에 닿아야 비로소 라이선스를 등록할 수 있으므로, 그 라우트는 SDK 안에 목록으로 두지 않고 레이어 바깥에서 합친다.

## 거부를 응답으로 옮기기

`GateError`는 `status`, `error_code`, `denial`, `http_status`를 싣는다. 자체 응답 형태를 쓰는 제품은 이 필드를 읽는다.

```rust
let message = translate(locale, &error.denial.message_key, &error.denial.arguments);
(StatusCode::from_u16(error.http_status)?, Json(MyEnvelope::error(&error.error_code, message)))
```

메시지 키와 인자 순서는 [messages.ko.md](messages.md)에 있다.

## 코어 크레이트 직접 사용

바인딩이 감싸지 않은 것이 필요하면 해당 크레이트를 직접 쓴다: 형식은 `aclic_model`, 키와 서명은 `aclic_sign`, 판정은 `aclic_verify`, 머신 식별자는 `aclic_fingerprint`, 활성화 본문은 `aclic_protocol`.
