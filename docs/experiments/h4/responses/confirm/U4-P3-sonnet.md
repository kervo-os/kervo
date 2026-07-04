[sonnet/no-repo/confirm]

# 환불 처리 방식 정리 (payments-api / h4-repo)

## 전제 및 한계
- 이 요약은 오직 `artifact-P3.md` 한 문서에만 근거한다. 실제 소스코드(예: `handlers/refund.go`)는 열람하지 못했으므로, 문서에 명시된 내용 이상은 알 수 없다.
- 문서 서두에 따르면 "Fact" 섹션(Repository Summary, Recent Changes, Frequently Changed Files, Open Tasks 등)은 **결정론적**으로 기록되고, "Slot" 섹션(Known Decisions/Risks/Summaries/Stale)은 **신뢰 라벨**(verified / observed / generated / stale)이 붙은 관찰이다. 아래 확신도 판정은 이 구분을 그대로 따른다.

## 1. 환불은 비동기(제공업체 웹훅 기반)로 처리된다 — 확신도: 중간 (⚠ 주의 필요)
- 근거: Known Decisions 슬롯 — `[verified — human:kim] ⚠ conflict — Refunds are processed asynchronously via provider webhooks, not inline.`
- 라벨 자체는 사람(kim)이 "verified"로 표시해 원칙적으로 신뢰도가 높아야 하나, 동시에 `⚠ conflict` 플래그가 붙어 있다. 이 문서만으로는 **무엇과 충돌하는지, 충돌이 해소되었는지 전혀 알 수 없다.** 따라서 "팀 전체가 확정적으로 합의한 사실"이 아니라 "한 사람이 이렇게 파악했고, 다른 소스(코드/다른 결정)와 상충하는 부분이 남아 있어 재확인이 필요한 상태"로 취급해야 한다.
- → 종합 확신도: **중간** (라벨상으로는 verified지만 미해결 conflict로 하향)

## 2. 환불 핸들러 코드가 저장소에 존재한다 — 확신도: 높음
- 근거: Recent Changes(결정론적 Fact) — `bb2a9d7 2026-07-03 add refund handler`; Frequently Changed Files에 `handlers/refund.go` (변경 1회) 포함.
- 이는 신뢰 라벨이 붙지 않은 Fact 섹션이므로 "파일 존재 + 해당 커밋으로 추가/수정됨" 자체는 신뢰도가 높다.
- 다만 변경 횟수가 1회뿐이라 초기 구현 단계일 가능성이 있다 — 이는 추정이며 문서가 명시한 결론은 아니다.

## 3. 환불 관련 idempotency/재시도 로직 상태 — 확신도: 불명 (정보 없음)
- Open Tasks에는 `handlers/charge.go:2 — FIXME: idempotency keys not enforced`만 있고, `refund.go`에 대한 TODO/FIXME는 문서에 없다.
- 이것이 "환불에는 idempotency 문제가 없다"는 뜻인지, 단순히 아직 이슈로 등록되지 않았을 뿐인지 문서로는 판단 불가하다. charge의 FIXME를 refund에 확장 적용해서는 안 된다.
- Known Risks 슬롯도 "settlement cron 재시도 로직 부재"만 언급하며(`[generated — backend:gpt-oss-120b]`, 즉 사람 검증 없는 모델 생성 관찰이라 그 자체로도 신뢰도가 낮음) 환불과는 무관하다.

## 4. 정산(Kafka 전환)과 환불의 관계 — 확신도: 낮음/추정 불가
- Deprecated/Stale Notes: "정산 작업이 레거시 RabbitMQ에서 Kafka로 이전됨(2026년 6월, `queue/kafka.go`)" — 이는 명시적으로 **settlement**에 대한 내용이며, 환불 플로우와 직접 연결된다는 서술은 문서 어디에도 없다.
- 환불 처리가 이 큐 인프라를 공유하는지 여부는 이 문서로는 확인할 수 없다. 임의로 연관 지어 추정하는 것은 위험하다.

## 5. 관련 일반 규칙 — 확신도: 중간
- Doc Summaries: `[observed — human:lee] 팀 컨벤션: 모든 핸들러 로직에 대해 table-driven test 작성.` 환불 핸들러를 특정한 서술은 아니지만, "모든 핸들러"라는 일반 규칙이므로 refund.go에도 적용될 것으로 보인다. 사람이 관찰(observed)한 라벨로, "verified"보다는 한 단계 낮은 신뢰도다.

## 6. 환불 관련 현재 목표/의도 — 확신도: 없음 (공백)
- "Possible Current Goal" 슬롯은 비어 있다 ("No proposal yet. A confirmed goal becomes the first Verified observation."). 환불 기능이 현재 작업의 목표인지, 이미 완료로 간주되는지 등은 전혀 알 수 없다.

## 요약 확신도 등급표

| 진술 | 확신도 | 근거 라벨 |
|---|---|---|
| 환불은 provider webhook을 통한 비동기 처리 | 중간 (충돌 미해결) | verified(human:kim) + ⚠conflict |
| refund.go 핸들러 존재, `bb2a9d7`에서 추가 | 높음 | Fact 섹션(결정론적) |
| 환불 관련 idempotency/재시도 상태 | 불명 | 문서에 해당 항목 부재 |
| 정산 Kafka 전환과 환불의 연관성 | 낮음(추정 불가) | stale note는 settlement에 한정 |
| 핸들러 테스트는 table-driven 컨벤션 적용 | 중간 | observed(human:lee) |
| 환불 관련 현재 작업 목표 | 없음 | goal 슬롯 비어있음 |

## 확실하지 않은 전제 (명시)
- "팀이 안다"는 표현이 실제 팀 전체 합의인지, 개인(kim, lee) 한 명의 진술인지는 문서만으로 구분할 수 없다 — 라벨이 개인 단위(`human:kim`, `human:lee`)로 부여되어 있다.
- `⚠ conflict` 표시의 구체적 내용(무엇과, 왜 충돌하는지)은 문서에 없어 확인 불가하다.
- 문서 자체가 "Regenerate with `kervo compile` — do not edit by hand"라고 명시된 자동 생성 스냅샷이므로, 생성 시점 이후 저장소가 추가로 변경되었을 가능성이 있다(문서 내 최신 커밋 기재일은 2026-07-04).
