# [sonnet/no-repo/confirm]

## 전제 및 자료 한계

- 이 정리는 오직 `/Users/[user]/Developer/30_lab/EXPER/h4-kit/artifact-M9.md` 한 파일(자동 생성 컨텍스트 아티팩트, `h4-repo`)만을 근거로 한다. 실제 저장소 코드(`handlers/refund.go` 등)는 열람하지 않았다.
- 이 아티팩트는 자체적으로 "Fact sections are deterministic; slot sections carry trust-labeled observations"라고 명시한다. 즉 Repository Summary / Recent Changes / Frequently Changed Files / Open Tasks 등은 저장소 상태에서 기계적으로 추출된 사실(fact)이고, Known Decisions / Known Risks / Doc Summaries 등은 별도의 신뢰도 검증을 거치지 않은 "슬롯" 관찰이다. "Possible Current Goal" 섹션의 문구("A confirmed goal becomes the first Verified observation")로 미루어 볼 때, 이 문서 안에서 "Verified"로 명시 태깅된 항목은 하나도 없다 — Known Decisions의 항목들은 전부 미검증 상태로 취급해야 한다.
- 아래 각 진술 뒤에 확신도(높음/중간/낮음/불명)를 표시했다. 확신도는 "그 사실이 참인지"가 아니라 "이 문서만으로 그렇게 판단할 근거가 얼마나 되는지"를 뜻한다.

## 환불 처리 방식에 대해 알려진 것

1. **환불은 인라인이 아니라 결제대행사(provider) 웹훅을 통해 비동기로 처리된다.**
   확신도: **낮음~중간**. 출처는 `Known Decisions` 슬롯의 문장 "Refunds are processed asynchronously via provider webhooks, not inline." 뿐이다. 이 슬롯 섹션은 신뢰도 라벨이 없는 미검증 관찰이며, 아래 "모순 경고"에서 보듯 같은 섹션의 다른 항목 2개가 최신 fact 섹션과 직접 충돌하는 것으로 보아 이 섹션 전체의 최신성/정확성을 의심할 근거가 있다. 따라서 이 진술은 "팀이 한때 그렇게 문서화했다"는 정도로만 신뢰할 수 있고, 현재도 유효한지는 불확실하다.

2. **환불 처리 코드(핸들러)가 존재하며 최근에 추가/변경되었다.**
   확신도: **높음**. `Recent Changes`(fact)에 `bb2a9d7 2026-07-03 add refund handler` 커밋이 있고, `Frequently Changed Files`(fact)에 `handlers/refund.go (1)`이 나열되며, `Related Modules`(fact)에 `handlers/ (2 files)`가 있다. 파일의 존재와 최근 변경 이력 자체는 결정론적 사실 섹션에서 나온 것이므로 신뢰도가 높다. 다만 그 파일의 **내부 구현 방식**(비동기/웹훅 여부 등)은 이 fact 섹션만으로는 확인되지 않는다 — 그 세부는 위 1번 항목(슬롯, 저신뢰)에만 나온다.

3. **환불 핸들러에 멱등성(idempotency) 키 검증이 되어 있는지는 불명확하다.**
   확신도: **불명**. `Open Tasks`(fact)에는 `handlers/charge.go:2 — FIXME: idempotency keys not enforced`만 명시되어 있고, `handlers/refund.go`에 대한 TODO/FIXME는 문서에 전혀 나열되지 않았다. 이는 "refund.go는 문제없다"는 뜻이 아니라, 단지 이 아티팩트가 포착한 4개 커밋 범위 내에서 refund.go 관련 TODO/FIXME 주석이 없었다는 뜻일 뿐이다(문서는 "Commits analyzed: 4 (complete)"라고 밝힘). 환불에도 동일한 멱등성 이슈가 있는지 여부는 판단 근거가 없다.

4. **환불 관련 재시도(retry) 로직 여부.**
   확신도: **불명**. `Known Risks` 슬롯에는 정산(settlement) 크론의 재시도 로직 부재만 언급되어 있고("main.go carries a TODO"), 환불 경로의 재시도/실패 처리에 대한 언급은 문서 어디에도 없다.

5. **환불 핸들러의 테스트 커버리지가 팀 컨벤션(table-driven tests)을 따르는지.**
   확신도: **낮음**. `Doc Summaries` 슬롯에 "Team convention: table-driven tests for all handler logic"라는 일반 컨벤션만 있을 뿐, 이것이 refund.go에 실제 적용됐는지 확인하는 fact는 없다. 컨벤션 존재 자체도 슬롯(미검증) 출처다.

6. **환불과 정산(Kafka/Queue) 파이프라인의 관계.**
   확신도: **불명**. README 발췌(fact)는 "Settlement jobs flow through Kafka (queue/)"라고 하고 가장 최근 커밋(fact, 2026-07-04)은 "kafka settlement + multi-region topology"다. 반면 환불이 이 큐/정산 파이프라인과 연동되는지, 아니면 완전히 별도 경로(웹훅)로만 처리되는지는 문서에 명시된 바 없다. 환불이 정산 배치와 상호작용하는지는 추정만 가능하고 확인할 수 없다.

## 모순 경고 (신뢰도 판단에 중요)

`Known Decisions` 슬롯 안에서 아래 두 항목이 더 신뢰도 높은 fact 섹션과 직접 충돌한다.

- 슬롯: "Settlement jobs go through the legacy RabbitMQ queue (settlement_jobs)" ↔ fact: 최신 커밋(`74d4646`, 2026-07-04)이 "kafka settlement"이고, `Frequently Changed Files`에 `queue/kafka.go`가 있음(RabbitMQ 파일은 전혀 언급 안 됨).
- 슬롯: "Payments runs single-region in us-east-1; DR is cold-standby only" ↔ fact: 같은 최신 커밋이 "multi-region topology"이며 `deploy/regions.yaml`이 최근 변경 파일로 잡혀 있음.

이 두 건은 `Known Decisions` 섹션이 **적어도 부분적으로 stale(오래된 정보)**임을 강하게 시사한다. 이 섹션에 함께 있는 환불 관련 진술("비동기·웹훅 처리")도 같은 섹션 출처이므로, 다른 두 항목과 마찬가지로 이미 바뀌었을 가능성을 배제할 수 없다 — 다만 이를 직접 반박하는 fact는 문서에 없으므로 "확실히 틀렸다"고는 말할 수 없고, "신뢰도를 낮춰서 봐야 한다" 정도로만 판단 가능하다.

## 확인되지 않아 답할 수 없는 것

- 환불 승인/거절 조건, 부분 환불 지원 여부, 환불 가능 기간(윈도우), 환불 금액 검증 로직 — 문서에 근거 없음.
- 환불 실패 시 재처리·알림 절차 — 문서에 근거 없음.
- 환불 요청과 카드 벤더(vault) 서비스 간의 상호작용 — 문서는 "Card PANs are stored only in the vault service"만 언급하며 환불과의 연결은 다루지 않음(이 진술 자체도 Known Decisions 슬롯 출처, 저신뢰).
- 환불 핸들러 자체의 인증/인가 방식 — Auth 관련 항목(JWT 15분 만료)은 있으나 환불 엔드포인트에 특정되지 않음.

## 결론

문서 근거로 확언할 수 있는 것은 "환불 핸들러 코드가 존재하고 최근에 추가/수정되었다"(높음) 정도다. "환불이 비동기·웹훅 기반으로 처리된다"는 팀 진술은 있지만 미검증 슬롯 출처이며, 같은 섹션의 다른 두 항목이 최신 fact와 모순되는 것으로 보아 낮음~중간 신뢰도로만 취급해야 한다. 멱등성, 재시도, 테스트 커버리지, 정산 파이프라인과의 연동 여부는 이 문서만으로는 판단 불가(불명)이며, 실제 `handlers/refund.go` 코드를 직접 확인하기 전까지는 확정할 수 없다.
