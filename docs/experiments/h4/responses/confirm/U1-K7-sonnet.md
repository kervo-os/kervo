[sonnet/no-repo/confirm]

# 신규 정산 컨슈머 — 바인딩 대상 큐/토픽

## 결론

**Kafka**(정확히는 `queue/kafka.go`가 다루는 정산용 Kafka 토픽)에 바인딩해야 한다.
문서 내 "Deprecated / Stale Notes" 섹션에 있는 **RabbitMQ `settlement_jobs` 큐는 사용하지 말 것** — 해당 섹션 자체가 문서 스키마상 "폐기/오래됨"으로 명시적으로 라벨링된 정보다.

다만 **정확한 Kafka 토픽 이름은 이 문서만으로는 확인 불가**하다. 아래 "확실하지 않은 부분" 참조.

## 근거

1. **README.md 발췌 (결정론적 팩트 섹션, Repository Summary)**: "Settlement jobs flow through Kafka (queue/); deployed multi-region (deploy/regions.yaml)." — 정산 잡은 Kafka를 통해 흐른다고 명시. 이 섹션은 문서 서두에 "Fact sections are deterministic"라고 규정된 영역으로, 신뢰도가 가장 높다.

2. **Recent Changes 중 최신 커밋** `74d4646` (2026-07-04, 문서 내 가장 최근 날짜): "kafka settlement + multi-region topology" — 정산을 Kafka 기반으로 옮기는(혹은 확정하는) 작업이 가장 최근에 이뤄졌음을 시사한다.

3. **Frequently Changed Files / Related Modules**: `queue/kafka.go`가 목록에 있고 `queue/` 모듈이 Related Modules에 잡혀 있다. 반면 RabbitMQ 관련 파일이나 모듈은 Frequently Changed Files, Related Modules 어디에도 등장하지 않는다 — 즉 RabbitMQ는 현재 활성 코드 경로에 흔적이 없다.

4. **"Deprecated / Stale Notes" 섹션의 구조적 의미**: 문서 서두 설명에 따르면 slot 섹션은 "trust-labeled observations"이며, 이 노트는 별도로 "Deprecated / Stale"이라는 카테고리에 격리되어 있다. 즉 문서 생성기 자체가 이 정보를 "낡았을 가능성이 있음"으로 이미 표시해 둔 것이다. 이 내용(RabbitMQ `settlement_jobs`)은 Kafka 마이그레이션 이전 시점의 지식으로 보이며, 최신 Fact 섹션·최신 커밋과 직접 충돌한다.

## 확실하지 않은 부분 (전제로 삼지 말 것)

- **정확한 Kafka 토픽명은 문서에 없다.** README 발췌는 "Kafka를 통해 흐른다"는 사실만 알려줄 뿐, 토픽 이름·파티션 전략·컨슈머 그룹 규약은 언급하지 않는다. `queue/kafka.go` 실제 코드를 확인해야 한다.
- **"multi-region topology"의 의미가 불명확하다.** `deploy/regions.yaml`이 리전별로 별도 토픽/브로커를 두는 구조인지, 단일 토픽에 리전 태그만 붙이는 구조인지 이 문서만으로는 알 수 없다. 리전별 바인딩이 다를 수 있으므로 확인 필요.
- **`main.go`의 TODO "wire settlement cron"**: 정산 크론 자체가 아직 완전히 연결되지 않았을 가능성이 있다(Open Tasks, Known Risks에서 재확인). 즉 Kafka 프로듀서/컨슈머 배선이 코드상 아직 미완일 수 있어, 신규 컨슈머를 붙이는 시점/방식에 대해 팀과 조율이 필요하다.
- 이 RabbitMQ 노트가 "이미 완전히 폐기됐다"는 것도 문서만으로 100% 확정할 수는 없다 — 다만 Fact 섹션·최신 커밋과 명백히 상충하므로, 신뢰도상 Kafka 쪽을 따르는 것이 합리적이라는 판단이다. 실제 마이그레이션 완료 여부는 코드/인프라 확인으로 검증해야 한다.

## 권장 후속 조치

- `queue/kafka.go`를 열어 실제 토픽명, 파티셔닝, 컨슈머 그룹 설정을 확인한다.
- `deploy/regions.yaml`을 확인해 리전별로 바인딩 대상이 달라지는지 점검한다.
- `main.go`의 "wire settlement cron" TODO 진행 상태를 팀에 확인해, 신규 컨슈머 투입 타이밍이 안전한지 확인한다.
- RabbitMQ `settlement_jobs` 큐가 완전히 폐기(decommission)됐는지, 혹은 과도기적으로 병행 운영 중인지 팀에 재확인한다(문서상 "Deprecated"로만 표시되어 있어 병행 가능성을 배제할 수 없음).
- `handlers/charge.go`의 FIXME(idempotency 미적용)를 감안해, 신규 컨슈머 설계 시 멱등 처리를 자체적으로 구현할 것을 권장한다(정산 이벤트 중복 소비 대비).
