# Java

## 가져다 쓰기

```groovy
repositories {
    maven {
        url = uri('https://accesscore-dev.github.io/accesscore-license-sdk-dist/packages/maven')
        content { includeGroup 'dev.accesscore' }
    }
}

dependencies {
    implementation 'dev.accesscore:license-sdk:1.0.0'
}
```

For Maven:

```xml
<repository>
  <id>accesscore</id>
  <url>https://accesscore-dev.github.io/accesscore-license-sdk-dist/packages/maven</url>
</repository>
```

네이티브 코어는 jar 안에 플랫폼당 한 디렉터리로 담기고, 처음 필요할 때 임시 파일로 추출된다. 공유 라이브러리의 출처를 자체 정책으로 고정한 배포는 `-Daclic.native.library=<경로>`로 파일을 지정한다 — 적재할 파일을 지정할 뿐 신뢰할 키를 지정하지 않으므로 어떤 검사도 약화시키지 않는다.

## 제품이 제공해야 하는 것

`LicenseSpi`에 모인 인터페이스 일곱 개다. 앞의 셋만 필수다.

| 인터페이스 | 답하는 것 |
| --- | --- |
| `LicenseStore` | 등록 기록 읽기와 쓰기 |
| `ProductIdentity` | 이 배포가 어떤 제품이고 어떤 릴리스인지 |
| — 검증 키 | 클래스패스 리소스에서 읽어 빌더에 넘긴다 |
| `FeatureCatalogue` | 이 배포가 게이트하는 기능 키 |
| `QuotaCounter` | 한도 코드별 현재 수량 |
| `FeatureActivationChecker` | 운영자가 그 기능을 켰는지 |
| `AuditRecorder` | 라이선스 이벤트 기록 위치 |
| `ActivationServerDirectory` | 활성화 서버 주소 |

## 조립

```java
LicenseManager manager = LicenseManager
        .builder(store, identity, List.of(KeyMaterial.of("acps-2026-01", publicKeyPem)))
        .catalogue(catalogue)
        .counter(userCounter)
        .counter(hardwareCounter)
        .audit(auditRecorder)
        .configuredServerUrl(properties.getServerUrl())
        .instanceId(properties.getInstanceId())
        .build();

manager.verify(Instant.now(), integrityChecker.verify());
```

`verify`는 현재 시각과 무결성 검사 결과를 호출자에게서 받는다. 코어는 시계를 읽지 않고, 무결성 검사 대상은 언어마다 다르다.

## 강제

```java
// 라이선스가 없는 동안 모든 요청을 막는다.
new LicenseEnforcementFilter(manager.gate(), List.of("/license/**", "/activations/**"), mapper)

// 기능 하나.
@RequiresFeature(PacsFeatures.ATTENDANCE)
public class AttendanceRestController { }

// 한도 하나, 추가 직전에.
manager.checkQuota(PacsQuotas.USERS);
```

라이선스가 없는 동안 어느 경로를 열어 둘지는 제품이 정한다. 배포물은 자기 라이선스 화면에 닿아야 비로소 라이선스를 등록할 수 있고, 그 화면이 어디인지는 제품만 안다.

## 거부를 응답으로 옮기기

`LicenseDeniedException`은 상태, 오류 코드, 메시지 키, 인자를 싣는다. 모든 어댑터가 이것으로 같은 문서를 쓴다. 자체 응답 봉투를 쓰는 제품은 핸들러 하나로 옮긴다.

```java
@ExceptionHandler(LicenseDeniedException.class)
public ResponseEntity<ApiResponse<Void>> handle(LicenseDeniedException denied, Locale locale) {
    String message = messageSource.getMessage(
            denied.denial().messageKey(), denied.denial().argumentArray(), locale);
    return ResponseEntity.status(denied.httpStatus())
            .body(ApiResponse.error(denied.errorCode(), message));
}
```

메시지 키와 인자 순서는 [messages.ko.md](messages.md)에 있다.

## 발급

개인키는 발급하는 쪽만 갖는다. 배포물은 `LicenseIssuer`를 클래스패스에 두고도 쓸 일이 없다 — 서명에는 고객이 갖지 못하는 키가 필요하다.

```java
LicenseIssuer issuer = new LicenseIssuer();

// 기동 시: 이 서버가 어느 키페어로 서명하는지 알린다. 배포물이 가진 키와 어긋나면
// 로그 한 줄로 드러나고, 고객의 첫 활성화가 실패한 뒤에 찾을 일이 없다.
log.info("License signing key fingerprint: {}", issuer.fingerprintOf(privateKeyPem));

SignedToken signed = issuer.sign(payload, privateKeyPem);
```

개인키는 PKCS#8로 넘기고 코어가 서명 직후 메모리에서 지운다. 호출자가 든 사본은 호출자가 지운다.

## 제품 키와 릴리스 버전

양쪽이 모두 필요로 하고 서로 답이 같아야 하므로, 모든 연산이 코어로 건너간다.

```java
ProductKeys keys = new ProductKeys();
String key = keys.generate("ACPS");
String shown = keys.masked("ACPS", key);

ReleaseVersions releases = new ReleaseVersions();
boolean entitled = releases.isWithinCeiling("3.2.0", "3.2");
String sortKey = releases.sortKey("3.2");   // SQL 정렬용
```
