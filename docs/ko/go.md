# Go

## 가져다 쓰기

```
go get github.com/accesscore-dev/accesscore-license-sdk-dist/go@v1.0.0
```

코어는 정적으로 링크되므로 Go 바이너리가 코어를 안고 다닌다 — 옆에 공유 라이브러리를 깔 필요가 없다. cgo가 필요하며 `CGO_ENABLED=0`으로는 빌드되지 않는다.

## 제품이 제공해야 하는 것

패키지의 인터페이스다. 검증 키와 함께 앞의 둘만 필수다.

| 인터페이스 | 답하는 것 |
| --- | --- |
| `Store` | 등록 기록 읽기와 쓰기 |
| `ProductIdentity` | 이 배포가 어떤 제품이고 어떤 릴리스인지 |
| `FeatureCatalogue` | 이 배포가 게이트하는 기능 키 |
| `QuotaCounter` | 한도 코드별 현재 수량 |
| `FeatureActivationChecker` | 운영자가 그 기능을 켰는지 |
| `AuditRecorder` | 라이선스 이벤트 기록 위치 |
| `ActivationServerDirectory` | 활성화 서버 주소 |

## 조립

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

`Verify`는 현재 시각과 무결성 검사 결과를 호출자에게서 받는다. 코어는 시계를 읽지 않고, 무결성 검사 대상은 언어마다 다르다.

## 강제

```go
// 라이선스가 없는 동안 모든 요청을 막는다.
handler := aclic.LicenseGateMiddleware(manager.Gate(), []string{"/license", "/activations"})(mux)

// 기능 하나.
protected := aclic.RequireFeatureMiddleware(manager.Gate(), "ATTENDANCE")(attendanceHandler)

// 한도 하나, 추가 직전에.
if err := manager.CheckQuota("USERS"); err != nil {
    aclic.WriteDenial(writer, err)
    return
}
```

gin·echo·fiber는 `GinLicenseGate`, `EchoLicenseGate`, `FiberLicenseGate`로 얹는다. 각 프레임워크의 컨텍스트가 이미 만족하는 작은 인터페이스만 받으므로 이 모듈은 셋 중 무엇에도 의존하지 않고, 제품은 쓰는 것 하나를 세 줄로 연결한다.

라이선스가 없는 동안 어느 경로를 열어 둘지는 제품이 정한다. 배포물은 자기 라이선스 화면에 닿아야 비로소 라이선스를 등록할 수 있다.

## 거부를 응답으로 옮기기

거부는 `Status`, `ErrorCode`, `Denial`, `HTTPStatus`를 실은 `*LicenseDenied`다. `WriteDenial`이 모든 어댑터가 쓰는 문서를 쓰고, 자체 응답 형태를 쓰는 제품은 필드를 읽어 자기 형태로 만든다.

```go
var denied *aclic.LicenseDenied
if errors.As(err, &denied) {
    message := translate(request.Context(), denied.Denial.MessageKey, denied.Denial.Arguments)
    respond(writer, denied.HTTPStatus, denied.ErrorCode, message)
    return
}
```

메시지 키와 인자 순서는 [messages.ko.md](messages.md)에 있다.

## 코어 직접 호출

매니저가 감싸지 않은 연산은 호출 하나로 닿는다.

```go
var answer map[string]any
if err := aclic.Shared.Release(map[string]any{
    "operation": "WITHIN_CEILING", "version": "3.2.0", "other": "3.2",
}, &answer); err != nil {
    return err
}
```
