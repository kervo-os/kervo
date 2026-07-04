[sonnet/no-repo/confirm]

# 정산 컨슈머 바인딩 대상 조사 (h4-repo / payments-api)

전제: 저장소 코드에는 접근하지 못했고, `/Users/[user]/Developer/30_lab/EXPER/h4-kit/artifact-P3.md` (kervo 자동 생성 컨텍스트 아티팩트) 한 건만 근거로 사용함.

## 결론

신규 정산(settlement) 컨슈머 서비스는 **Kafka**에 바인딩해야 한다. RabbitMQ의 `settlement_jobs` 큐는 레거시이며 문서 스스로 "폐기(stale)"로 표시하고 있다. 다만 **바인딩할 정확한 Kafka 토픽명은 문서에 명시되어 있지 않다** — 이는 `queue/kafka.go` 소스를 직접 확인해야 확정할 수 있다.

## 근거

1. **Fact 섹션 (deterministic, 신뢰도 높음)**
   - README.md 발췌: "Settlement jobs flow through Kafka (queue/)" — 정산 작업이 Kafka를 통해 흐른다고 명시.
   - Recent Changes 중 가장 최신 커밋 `74d4646` (2026-07-04): "kafka settlement + multi-region topology" — Kafka 전환이 문서 기준 가장 최근 시점에 반영됨.
   - Frequently Changed Files에 `queue/kafka.go`가 등재되어 있고, Related Modules에도 `queue/` 모듈(1개 파일)이 명시됨 — queue 모듈의 실제 구현체는 kafka.go 하나뿐으로, RabbitMQ 관련 파일은 현재 파일 목록에 없음.

2. **Deprecated/Stale Notes 섹션의 경고 (반드시 무시해야 할 함정)**
   - 이 섹션에 "Settlement jobs go through the legacy RabbitMQ queue (settlement_jobs); new consumers should bind there."라는 문장이 실제로 존재한다. 그러나 이 항목은 **`[stale — superseded by code change: settlement moved to Kafka (queue/kafka.go, June 2026)]`**로 문서 자체가 명시적으로 태깅해 두었다. 즉 "예전엔 맞았지만 이제는 틀린" 정보라고 문서가 스스로 경고하는 항목이다. 이 문장을 그대로 따라 RabbitMQ `settlement_jobs`에 바인딩하면 안 된다.
   - 이 항목은 오히려 "언제, 무엇에서 무엇으로 전환되었는지"를 알려주는 정황 증거로만 활용해야 한다 (2026년 6월 코드 전환, 7월 4일 커밋으로 최종 반영된 것으로 추정).

3. **정합성 체크**: Fact 섹션(결정론적, README/커밋/파일 목록)과 Stale 섹션(시스템이 스스로 "낡음"으로 표시)이 서로 모순 없이 "현재는 Kafka"라는 결론으로 수렴한다. 두 섹션이 상충하지 않고 오히려 서로를 보강한다.

## 확실하지 않은 부분 (문서에 근거 없음 — 추측하지 않음)

- **정확한 Kafka 토픽명**: 문서에는 "Kafka"라고만 나오고 실제 topic 이름(예: `settlement_jobs`, `settlement.events` 등)은 어디에도 기재되어 있지 않다. `queue/kafka.go` 원본을 읽어야 확정 가능하다. RabbitMQ 시절 큐 이름이 `settlement_jobs`였다는 것은 문서로 알 수 있으나, Kafka 토픽명이 이를 그대로 계승했는지는 불명이다.
- **브로커/클러스터 주소 및 파티션 전략**: multi-region 배포(`deploy/regions.yaml`)와 맞물려 리전별로 별도 Kafka 클러스터/토픽을 쓰는지, 단일 글로벌 클러스터인지 문서에 정보가 없다.
- **컨슈머 그룹 이름, 메시지 스키마, 인증/ACL 설정**: 문서에 없음.
- **운영 성숙도 리스크**: main.go에 "TODO: wire settlement cron"이 남아 있고, (신뢰도가 낮은 generated 관찰이지만) "정산 cron에 재시도 로직이 없어 보인다"는 관찰이 있다 — 신규 컨슈머 설계 시 idempotency/재시도를 자체적으로 고려할 필요가 있어 보이나, 이 리스크 항목은 human-verified가 아니라 backend 모델이 생성한 관찰(생성 신뢰도 낮음)이므로 그대로 확정 사실처럼 받아들이면 안 된다.
- **handlers/charge.go의 FIXME(idempotency 키 미적용)**가 정산 컨슈머와 직접 연결되는지는 불명하나, 결제·정산 파이프라인 전반의 idempotency 취약점 정황으로 참고할 만하다.
- Known Decisions 섹션에 "환불은 webhook을 통해 비동기 처리된다"는 항목이 ⚠ conflict 표시와 함께 있는데, 이 충돌이 무엇과 상충하는지 문서에 구체적으로 드러나 있지 않다. 정산 컨슈머와 직접 관련이 있는지도 불확실하다.

## 요약 답변

> 새 정산 컨슈머는 **Kafka**에 바인딩해야 한다 (RabbitMQ `settlement_jobs`는 레거시로 폐기됨, 문서가 이를 stale로 명시). 단, 정확한 토픽명·클러스터 구성·컨슈머 그룹 등 바인딩에 필요한 구체적 파라미터는 이 컨텍스트 문서만으로는 확인 불가능하며 `queue/kafka.go` 코드를 직접 확인해야 한다.
