# AccessCore License SDK

라이선스의 형식과 그에 대한 판정을 하나의 코어가 소유한다.

Rust 코어가 토큰을 서명하고 검증하며 배포물이 무엇을 해도 되는지 판정한다. Java·Python·Go·Rust로 만든 제품은 얇은 바인딩으로 그 판정을 받아 쓴다. 코어는 파일도 시계도 네트워크도 만지지 않는다 — 호스트가 값을 넣고 코어가 결론을 돌려준다. 발급하는 쪽과 검증하는 쪽이 같은 코어를 쓰므로, 한쪽이 쓴 형식을 다른 쪽이 읽지 못하는 배포가 나가지 않는다.

## 구성

```
core/                       Rust workspace
├── aclic-model/            페이로드 형식 · 스키마 번호 · 채널 · 제품 키 · 릴리스 비교기
├── aclic-sign/             서명 생성·검증, 키 입출력
├── aclic-verify/           판정 — 검증 순서와 상태
├── aclic-fingerprint/      머신 식별자 수집과 대조
├── aclic-protocol/         활성화·하트비트 본문 생성과 응답 해석
└── aclic-ffi/              C ABI cdylib — 유일한 네이티브 경계
bindings/
├── java/                   JNA 래퍼 · 중립 게이트 · 어댑터
├── python/                 cffi 래퍼 · 중립 게이트 · 어댑터
├── go/                     cgo 래퍼 · 중립 게이트 · 어댑터
└── rust/                   코어 직접 사용 · 중립 게이트 · 어댑터
conformance/                모든 바인딩이 같은 답을 내야 하는 벡터
include/aclic.h             cbindgen이 생성하는 C 헤더
docs/en, docs/ko            모든 문서, 언어별 한 폴더
```

접두사 `aclic`은 AccessCore License를 줄인 것이다. Linux에 널리 깔린 POSIX ACL 라이브러리가 `acl_` 심볼을 쓰므로 그 접두사는 쓰지 않는다.

## 쓰는 쪽이 가져가는 곳

| 언어 | 좌표 | 코어 전달 방식 |
| --- | --- | --- |
| Java | `dev.accesscore:license-sdk` | jar 안에 담고 실행 시 추출 |
| Python | `accesscore-license` | wheel 안에 담음 |
| Go | `github.com/accesscore-dev/accesscore-license-sdk-dist/go` | 정적 링크 |
| Rust | `aclic` | 정적 링크 |

이 저장소는 비공개이며 코어를 담는다. 릴리스는 위 레지스트리에 발행하고, 쓰는 쪽이 필요한 트리를 공개 저장소에 밀어 넣는다 — `go get`이 모듈을 소스로 받아오기 때문에 공개 저장소가 필요하다. [repositories.md](repositories.md)를 참고한다.

## 빌드

```
cargo test --workspace                        코어
./scripts/build-native.sh --host-only         이 머신에 필요한 라이브러리
cd bindings/java   && ./gradlew test          Java 바인딩과 적합성 실행
cd bindings/python && python -m pytest tests  Python 바인딩
cd bindings/go     && go test ./...           Go 바인딩
cargo build -p aclic-ffi --features header    include/aclic.h 재생성
```

릴리스는 `./scripts/build-native.sh`로 모든 플랫폼을 빌드하며 교차 컴파일 도구가 필요하다. 도구가 없는 플랫폼은 로컬에서 건너뛰고 보고하며, CI의 릴리스 빌드에서는 실패로 처리한다.

## 읽는 순서

[concepts.md](concepts.md)부터 본다 — 코어가 무엇을 소유하고 호스트가 무엇을 소유하는지, 검사가 어떤 순서로 도는지, 각 검사가 왜 그 자리에 있는지. 그다음 쓰는 언어의 가이드를 본다. 메시지 키와 인자 순서는 [messages.md](messages.md)에 있다.
