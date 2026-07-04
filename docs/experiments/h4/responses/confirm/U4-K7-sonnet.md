[sonnet/no-repo/confirm]

# 환불(Refund) 처리 방식 — 팀이 아는 바 정리

## 0. 전제 및 제약

- 저장소(payments-api) 코드에는 접근하지 못했고, 자동 생성된 컨텍스트 문서 `artifact-K7.md` 한 건만 근거로 삼았다.
- 이 문서는 스스로를 "Fact 섹션은 결정적(deterministic), Slot 섹션은 trust-labeled observation"이라고 규정하지만, 실제 Slot 섹션(Known Decisions, Known Risks, Doc Summaries, Deprecated/Stale) 안에는 개별 항목별 신뢰도 라벨(예: Verified/Inferred)이 실제로 붙어 있지 않다. 따라서 아래 확신도는 문서가 부여한 것이 아니라, "Fact 섹션 여부 + 근거(Evidence) 명시 여부 + 다른 섹션과의 정합성"을 기준으로 **내가 추정**한 것이다. 이 추정 자체가 불확실성의 한 층위임을 밝힌다.
- 커밋은 4개만 분석되었고 "(complete)"라고 표시되어 있으나, 이것이 "저장소 전체 히스토리가 4커밋뿐"이라는 뜻인지 "표본 분석이 완료됨"이라는 뜻인지 문서만으로는 구분할 수 없다.

## 1. 환불 관련 진술과 확신도

| # | 진술 | 확신도 | 근거 / 비고 |
|---|---|---|---|
| 1 | `handlers/refund.go` 파일이 존재하며, 별도의 환불 핸들러 코드가 있다 | **높음** | Recent Changes의 `bb2a9d7` "add refund handler"(2026-07-03) 커밋, Frequently Changed Files 목록, Related Modules(`handlers/` 2 files)가 모두 일치. Fact 섹션(결정적) 기반. |
| 2 | 환불은 인라인 처리가 아니라 **프로바이더(PG사) 웹훅을 통해 비동기로 처리**된다 | **중간** | Known Decisions(Slot) 섹션에 명시. 다만 이 항목은 Known Risks 항목과 달리 "Evidence: ..." 같은 근거 인용이 없어, 코드로 실제 확인된 것인지 팀 구두 합의/설계 의도인지 구분 불가. |
| 3 | 환불 핸들러가 idempotency key를 강제하는지 여부 | **불명(정보 없음)** | Open Tasks의 idempotency 관련 FIXME는 `handlers/charge.go:2`에만 명시되어 있고 `refund.go`에는 해당 언급이 없음. "charge에 없으니 refund에도 없을 것"이라는 추정은 문서에 없는 내용을 확장하는 것이라 하지 않는다. |
| 4 | 카드 PAN은 vault 서비스에만 저장되고 API 레이어(따라서 환불 핸들러 포함)는 PAN을 직접 다루지 않는다 | **중간(일반 원칙), 환불 특화는 낮음** | Known Decisions에 아키텍처 전반 원칙으로 기술됨. 환불 흐름에도 적용된다는 것은 합리적 추정이지만, 문서가 환불 케이스를 콕 집어 언급한 것은 아님. |
| 5 | 환불 핸들러도 "테이블 기반 테스트(table-driven tests)" 컨벤션을 따른다 | **낮음** | Doc Summaries에 "모든 핸들러 로직에 대한 팀 컨벤션"으로 일반 서술되어 있을 뿐, refund.go에 실제 테스트가 존재/작성되었다는 직접 증거는 없음. |
| 6 | 환불 처리와 정산(settlement) 배치·크론의 관계 | **불명 + 문서 내 모순** | README는 "Settlement jobs flow through Kafka(queue/)"라 하고 최근 커밋 `74d4646`도 "kafka settlement + multi-region topology"인 반면, Deprecated/Stale Notes는 "Settlement jobs go through the legacy RabbitMQ queue(settlement_jobs); new consumers should bind there"라고 정반대로 서술한다. 이 모순은 정산 자체에 대한 것이며 환불이 정산에 어떻게 연결되는지는 문서에 아예 언급이 없다. 환불→정산 연계를 전제하면 안 된다. |
| 7 | 정산 크론에 재시도(retry) 로직이 없다 (`main.go` TODO) | 정산에 한해 **중간**(Known Risks에 "Evidence: Open Tasks" 명시), **환불과의 관련성은 불명** | 이 리스크는 정산 크론(`main.go`)에 관한 것으로 환불 핸들러와 직접 연결된다는 근거는 없음. |
| 8 | 환불을 로컬/CI에서 어떻게 실행·테스트하는지(커맨드) | **불명** | Commands 섹션: "No declared commands found (Makefile targets, package.json scripts)." |
| 9 | 부분 환불, 환불 상태값, 환불 실패 처리, 사용하는 PG(프로바이더) 종류 등 세부 로직 | **불명** | 문서 어디에도 언급 없음. |

## 2. 확실하지 않은 전제 (별도 강조)

- **"비동기·웹훅 기반 환불"이라는 핵심 진술(위 표 2번)의 근거 수준이 낮다.** Known Risks 섹션과 달리 이 결정 항목에는 근거 인용이 없어, 실제 코드(`handlers/refund.go`)를 열어보기 전에는 "설계 의도"인지 "현재 구현 상태"인지 구분할 수 없다.
- **정산 큐 기술 스택(Kafka vs. 레거시 RabbitMQ)이 문서 내에서 서로 모순된다.** 이는 환불 자체에 대한 것은 아니지만, "환불 확정 후 정산에 반영되는 방식"을 팀에 물어볼 때 이 모순을 먼저 해소해야 한다.
- **idempotency 이슈가 charge에만 명시되어 있다는 사실**을 "refund는 안전하다"는 뜻으로 오독하지 않아야 한다. 단순히 문서에 기록이 없는 것뿐이다.
- **Slot 섹션에 개별 신뢰도 라벨이 실제로는 부착되어 있지 않다.** 문서 헤더의 설명(trust-labeled observations)과 본문 실제 내용이 불일치하므로, 이 문서 형식 자체가 아직 신뢰도 메타데이터를 완전히 채우지 못한 초기/스텁 상태일 가능성이 있다.
- 커밋 4개 분석이 "저장소 전체"인지 "최근 일부"인지 불명확하여, 환불 관련 히스토리(예: refund.go 이후 수정 이력)가 더 있었는지 알 수 없다.

## 3. 결론

문서 기준으로 확실하게 말할 수 있는 것은 **"환불 핸들러 코드가 별도로 존재한다"** 는 사실(높음) 정도이며, **"비동기 웹훅 기반 처리"** 라는 설계 진술은 팀 문서상 존재하나 근거 인용이 없어 중간 확신도로만 다룰 수 있다. idempotency, 부분 환불, 테스트 존재 여부, 정산과의 연계 등 실무적으로 중요한 세부사항은 문서에 없거나 모순되어 있어, 실제 코드(`handlers/refund.go`, `queue/kafka.go`, `main.go`)를 직접 확인하기 전까지는 "불명"으로 남겨두는 것이 안전하다.
