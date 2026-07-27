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

아래는 전부 공개다 — 계정도 토큰도 필요 없다. 코어가 각 패키지 안에 함께 들어가므로 애플리케이션 옆에 따로 설치할 것이 없다.

**Java**

```groovy
repositories {
    maven { url 'https://accesscore-dev.github.io/accesscore-license-sdk-dist/packages/maven' }
}
dependencies {
    implementation 'dev.accesscore:license-sdk:1.0.0'
}
```

**Python**

```
pip install --extra-index-url https://accesscore-dev.github.io/accesscore-license-sdk-dist/packages/pypi/simple \
    accesscore-license==1.0.0
```

**Go**

```
go get github.com/accesscore-dev/accesscore-license-sdk-dist/packages/go@v1.0.0
```

| 언어 | 좌표 | 코어 전달 방식 |
| --- | --- | --- |
| Java | `dev.accesscore:license-sdk` | jar 안에 담고 실행 시 추출 |
| Python | `accesscore-license` | wheel 안에 담음 |
| Go | `github.com/accesscore-dev/accesscore-license-sdk-dist/packages/go` | 정적 링크 |
