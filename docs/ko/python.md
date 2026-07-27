# Python

## 가져다 쓰기

```
pip install accesscore-license==1.0.0
```

네이티브 코어는 wheel 안에 플랫폼당 한 디렉터리로 담긴다. 공유 라이브러리의 출처를 자체 정책으로 고정한 배포는 `ACLIC_NATIVE_LIBRARY`로 파일을 지정한다 — 적재할 파일을 지정할 뿐 신뢰할 키를 지정하지 않는다.

프레임워크 추가 설치는 쓰는 것만 가져온다: `accesscore-license[fastapi]`, `[django]`, `[flask]`.

## 제품이 제공해야 하는 것

`accesscore_license.host`의 추상 클래스를 구현한다. 검증 키와 함께 앞의 둘만 필수다.

| 클래스 | 답하는 것 |
| --- | --- |
| `Store` | 등록 기록 읽기와 쓰기 |
| `ProductIdentity` | 이 배포가 어떤 제품이고 어떤 릴리스인지 |
| `FeatureCatalogue` | 이 배포가 게이트하는 기능 키 |
| `QuotaCounter` | 한도 코드별 현재 수량 |
| `FeatureActivationChecker` | 운영자가 그 기능을 켰는지 |
| `AuditRecorder` | 라이선스 이벤트 기록 위치 |
| `ActivationServerDirectory` | 활성화 서버 주소 |

## 조립

```python
from datetime import datetime, timezone
from accesscore_license import LicenseManager

manager = LicenseManager(
    store=store,
    identity=identity,
    keys=[{"kid": "acps-2026-01", "publicKey": PUBLIC_KEY_PEM}],
    catalogue=catalogue,
    counters=[user_counter, hardware_counter],
    audit=audit_recorder,
    configured_server_url=settings.LICENSE_SERVER_URL,
    instance_id=settings.INSTANCE_ID,
)

manager.verify(datetime.now(timezone.utc), integrity_ok=check_integrity())
```

`verify`는 현재 시각과 무결성 검사 결과를 호출자에게서 받는다. 코어는 시계를 읽지 않고, 무결성 검사 대상은 언어마다 다르다.

## 강제

```python
from accesscore_license.adapters import (
    asgi_license_gate, fastapi_license_dependency, fastapi_register_exception_handler,
)

# 라이선스가 없는 동안 모든 요청을 막는다.
app = asgi_license_gate(app, manager.gate, open_prefixes=["/license", "/activations"])

# 기능 하나.
@router.get("/attendance", dependencies=[Depends(
    fastapi_license_dependency(manager.gate, "ATTENDANCE"))])
async def list_attendance(): ...

# 한도 하나, 추가 직전에.
manager.check_quota("USERS")

# 어디서 난 거부든 공통 문서로 나가게.
fastapi_register_exception_handler(app)
```

Django는 `django_license_middleware(gate, open_prefixes)`, Flask는 `flask_register(app, gate, open_prefixes)`, WSGI 전반은 `wsgi_license_gate`를 쓴다.

라이선스가 없는 동안 어느 경로를 열어 둘지는 제품이 정한다. 배포물은 자기 라이선스 화면에 닿아야 비로소 라이선스를 등록할 수 있다.

## 거부를 응답으로 옮기기

`LicenseDenied`는 `status`, `error_code`, `message_key`, `arguments`, `http_status`를 싣고, `body()`가 모든 어댑터가 쓰는 문서를 만든다. 자체 응답 형태를 쓰는 제품은 키를 직접 번역한다.

```python
from django.utils.translation import gettext

@app.exception_handler(LicenseDenied)
async def handle(request, denied):
    message = gettext(denied.message_key) % tuple(denied.arguments)
    return JSONResponse(
        status_code=denied.http_status,
        content={"code": denied.error_code, "message": message},
    )
```

메시지 키와 인자 순서는 [messages.ko.md](messages.md)에 있다.

## 발급

```python
from accesscore_license import LicenseIssuer

issuer = LicenseIssuer()
print("License signing key fingerprint:", issuer.fingerprint_of(private_key_pem))
signed = issuer.sign(payload, private_key_pem)
```

개인키는 발급하는 쪽만 갖는다. 코어는 서명 직후 개인키를 메모리에서 지운다.

## 제품 키와 릴리스 버전

```python
from accesscore_license import ProductKeys, ReleaseVersions

keys = ProductKeys()
key = keys.generate("ACPS")
shown = keys.masked("ACPS", key)

releases = ReleaseVersions()
entitled = releases.is_within_ceiling("3.2.0", "3.2")
```
