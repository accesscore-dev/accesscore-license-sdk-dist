# AccessCore License SDK — 배포본

릴리스 **1.0.0**. SDK를 쓰는 쪽이 필요한 것만 담는다. 코어는 비공개 저장소에 있고, 여기 있는 것으로는 라이선스를 발급할 수 없다.

## 담긴 것

| 경로 | 내용 |
| --- | --- |
| `go/` | Go 바인딩과 정적 링크할 코어 |
| `include/aclic.h` | C ABI — SDK가 지원하지 않는 언어를 직접 얹을 때 쓴다 |
| `native/` | 미리 빌드한 코어, 플랫폼당 한 디렉터리 |
| `docs/en`, `docs/ko` | 모든 문서, 언어별 한 폴더 |
| `conformance/` | 모든 바인딩이 같은 답을 내야 하는 벡터 |

## 가져다 쓰기

**Java** — GitHub Packages의 `dev.accesscore:license-sdk:1.0.0`.

**Python** — `accesscore-license==1.0.0`.

**Go** — `go get github.com/accesscore-dev/accesscore-license-sdk-dist/go@v1.0.0`.

**Rust** — `aclic = "1.0.0"`.

## 담기지 않은 것

Rust 코어, 서명 경로, 발급 측. 배포물은 토큰을 검증하고, 서명은 발급하는 쪽만 한다 — 서명에 필요한 개인키는 어느 저장소에도 들어간 적이 없다.
