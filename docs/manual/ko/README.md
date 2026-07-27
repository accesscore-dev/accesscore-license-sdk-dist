# AccessCore License SDK — 매뉴얼

SDK를 제품에 넣는 쪽을 위한 문서다.

Rust 코어가 라이선스를 서명하고 검증하며 배포물이 무엇을 해도 되는지 판정한다. Java·Python·Go로 만든 제품은 얇은 바인딩으로 그 판정을 받아 쓴다. 코어는 파일도 시계도 네트워크도 만지지 않는다 — 애플리케이션이 값을 넣고 코어가 결론을 돌려준다. 모든 언어가 같은 코어를 부르므로, 같은 라이선스를 읽는 두 배포는 거절 시의 메시지 키까지 같은 결론에 도달한다.

| | |
| --- | --- |
| [개념](concepts.md) | 코어가 판정하는 것, 애플리케이션이 판단할 것, 검사가 도는 순서 |
| [Java](java.md) | 가져다 쓰기, 구현할 인터페이스, 게이트, 프레임워크 어댑터 |
| [Python](python.md) | Python에서 같은 것 |
| [Go](go.md) | Go에서 같은 것 |
| [메시지 키](messages.md) | SDK가 돌려줄 수 있는 모든 거절 사유와 인자 순서 |

개념부터 읽고, 그다음 쓰는 언어의 가이드를 본다.

## 가져다 쓰기

코어가 각 패키지 안에 함께 들어가므로 애플리케이션 옆에 따로 설치할 것이 없다.

| 언어 | 받는 곳 | 토큰 |
| --- | --- | --- |
| Java | GitHub Packages 레지스트리의 `dev.accesscore:license-sdk` | `read:packages` — [Java 가이드](java.md) 참고 |
| Python | [릴리스](https://github.com/accesscore-dev/accesscore-license-sdk-dist/releases/tag/v1.0.0)에 붙은 wheel | 없음 |
| Go | `github.com/accesscore-dev/accesscore-license-sdk-dist/packages/go` | 없음 |

모든 언어의 산출물이 릴리스에 파일로도 붙는다 — jar, wheel, Go 모듈, 그리고 SDK가 지원하지 않는 언어를 C ABI로 직접 얹을 때 쓸 미리 빌드된 코어. 레지스트리에 닿을 수 없거나 닿고 싶지 않은 빌드는 파일을 가져간다.
