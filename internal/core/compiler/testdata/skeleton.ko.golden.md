<!-- kervo:artifact v1 skeleton=fact-only lang=ko -->
# Context Artifact

> AI 에이전트용으로 기계 생성된 컨텍스트. Fact 섹션은 결정적으로 생성되며,
> 슬롯 섹션에는 신뢰 라벨이 달린 관찰이 담긴다. `kervo compile`로 재생성 —
> 손으로 수정하지 말 것.

## 저장소 요약

- 이름: demo-api
- 브랜치: main
- 언어: Go, Markdown
- 프레임워크: Go
- 문서: README.md, CLAUDE.md

### README.md (발췌)

> A demo REST API used to exercise the compiler. It has two paragraphs.

## 명령어

- `make build` — go build -o demo ./cmd
- `npm run test` — jest --coverage

## 최근 변경

- `aaaaaaa` 2026-07-03 add auth middleware
- `bbbbbbb` 2026-07-02 fix token refresh
- `ccccccc` 2026-07-01 initial commit

_분석된 커밋 3개 중 3개 표시. 스캔 캡 도달 — 더 오래된 이력은 분석되지 않음 (Partial)._

### 자주 변경된 파일

- mw/auth.go (2)
- main.go (1)
- token.go (1)

## 미해결 작업

- mw/auth.go:42 — TODO: rotate signing keys
- main.go:7 — FIXME: graceful shutdown

## 관련 모듈

- mw/ (파일 3개)
- store/ (파일 5개)

## 워크스페이스 사실

- 분석된 커밋: 3 (부분 — 캡 도달)
- 미해결 작업 (TODO/FIXME): 2
- 최상위 모듈: 2
- 수집된 문서: 2

## 추정 현재 목표

<!-- kervo:slot:goal:begin -->
_아직 제안 없음. 확인된 목표가 첫 Verified 관찰이 된다._
<!-- kervo:slot:goal:end -->

## 알려진 결정

<!-- kervo:slot:decisions:begin -->
_아직 제안 없음. Semantic 제공자(Mode 2/3)가 라벨 달린 관찰을 여기에 부착한다._
<!-- kervo:slot:decisions:end -->

## 알려진 리스크

<!-- kervo:slot:risks:begin -->
_아직 제안 없음. Semantic 제공자(Mode 2/3)가 라벨 달린 관찰을 여기에 부착한다._
<!-- kervo:slot:risks:end -->

## 문서 요약

<!-- kervo:slot:summaries:begin -->
_아직 제안 없음. Semantic 제공자(Mode 2/3)가 라벨 달린 관찰을 여기에 부착한다._
<!-- kervo:slot:summaries:end -->

## 폐기·낡음 기록

<!-- kervo:slot:stale:begin -->
_기록 없음. 낡거나 폐기된 관찰은 조용히 삭제되는 대신
제외 사유와 함께 여기에 나열된다._
<!-- kervo:slot:stale:end -->

## Write-back 프로토콜

> AI 소비자용 — 루프를 닫을 것. 이 세션에서 이 artifact에 없는 지속적
> 사실을 알게 됐다면(무언가의 실행 방법, 컴포넌트의 역할, 대화 중 내려진
> 결정), 사람의 판정을 받도록 올려둘 것:
>
> `kervo capture -type decision|risk|summary|goal -actor "agent:<너>" -body "<사실>"`
>
> 규칙: 관찰한 사실만, 추측 금지 · 사실 하나당 캡처 하나 · 이 artifact에
> 이미 있는 내용은 생략 · 비밀값·파일 내용은 절대 넣지 말 것. 중복은
> 자동으로 버려진다. 사람이 `kervo review`로 큐를 판정한다.
