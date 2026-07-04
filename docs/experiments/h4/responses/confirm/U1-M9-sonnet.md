[sonnet/no-repo/confirm]

# 신규 정산 컨슈머 — 바인딩 대상 메시지 큐/토픽

## 결론 (확신도: 낮음 — 문서 내 근거가 상충함)

문서(artifact-M9.md)만으로는 **정확한 토픽명을 확정할 수 없고**, 오히려 문서 안에서 두 근거가 서로 모순된다. 실제 바인딩 전에 `queue/kafka.go`를 직접 열람하는 등 저장소 확인이 필요하다. 그럼에도 문서 내 증거의 신뢰도를 따져보면, **Kafka 쪽을 바인딩 대상으로 보는 것이 더 근거가 강하다**고 판단한다.

## 근거

### 1) 문서 구조상 신뢰도 차이
문서 상단 안내에 따르면 이 아티팩트는 "Fact sections are deterministic; slot sections carry trust-labeled observations"라고 명시되어 있다. 즉:
- **Fact 섹션** (Repository Summary, Recent Changes, Frequently Changed Files 등) = 결정론적으로 생성된, 저장소 상태를 직접 반영하는 사실.
- **Slot 섹션** (Known Decisions, Known Risks 등) = "trust-labeled observations", 즉 관찰/추정이며 검증된 사실보다 신뢰도가 낮음. 실제로 "Possible Current Goal" 섹션은 "확정된 목표는 첫 Verified observation이 된다"고 되어 있어, slot 내용이 아직 확정(verified)되지 않은 상태일 수 있음을 시사한다.

### 2) Fact 섹션이 가리키는 것: Kafka
- README.md 발췌(Fact 섹션): "Settlement jobs flow through Kafka (queue/); deployed multi-region (deploy/regions.yaml)."
- Recent Changes(Fact 섹션)의 **가장 최근 커밋**(74d4646, 2026-07-04, 문서 내 날짜 기준 최신): "kafka settlement + multi-region topology" — 정산 로직이 Kafka + 멀티리전 구조로 변경된 시점으로 읽힌다.
- Frequently Changed Files / Related Modules에 `queue/kafka.go`가 실존 파일로 잡혀 있다 (RabbitMQ 관련 파일은 문서 어디에도 나오지 않는다).

### 3) Slot 섹션(Known Decisions)이 가리키는 것: 레거시 RabbitMQ — 모순
- Known Decisions: "Settlement jobs go through the legacy RabbitMQ queue (settlement_jobs); new consumers should bind there."
- 동일 섹션: "Payments runs single-region in us-east-1; DR is cold-standby only."

이 두 문장은 Fact 섹션과 정면으로 충돌한다. 특히 "single-region" 주장은 가장 최근 커밋 메시지("... + multi-region topology")와 README의 "deployed multi-region"과 직접 모순된다. 이는 Known Decisions 슬롯이 **Kafka/멀티리전 마이그레이션 커밋(74d4646, 07-04) 이전 시점에 작성된 오래된(stale) 관찰**일 가능성을 강하게 시사한다. 다만 "Deprecated / Stale Notes" 섹션에는 "None recorded"라고 되어 있어, 시스템이 이 모순을 아직 공식적으로 표시(flag)하지 않은 상태다 — 즉 자동 생성 파이프라인이 놓친 것으로 보인다.

## 실무적으로 확인이 필요한 부분 (문서만으로는 답이 안 나옴)
1. **정확한 Kafka 토픽명** — 문서에 어떤 토픽 문자열도 명시되어 있지 않다. README는 "Kafka (queue/)"라고만 되어 있고, `queue/kafka.go` 내부 설정(토픽명, 컨슈머 그룹 등)은 문서에 나타나지 않는다.
2. **멀티리전 토폴로지의 의미** — 리전별로 별도 토픽/브로커가 있는지, 단일 글로벌 토픽인지 `deploy/regions.yaml`을 봐야 알 수 있는데 이 파일 내용도 문서에 없다.
3. **RabbitMQ 잔존 여부** — "legacy RabbitMQ queue (settlement_jobs)"가 완전히 폐기됐는지, 아니면 과도기적으로 병행 운영 중인지 문서만으로는 판단 불가. main.go의 TODO("wire settlement cron")가 이 마이그레이션과 관련 있을 수도 있으나 확인되지 않음.
4. **idempotency 미비 (handlers/charge.go FIXME)** — 컨슈머가 charge 관련 이벤트를 소비한다면 idempotency key 미검증 이슈가 중복 처리 위험으로 이어질 수 있음 (Open Tasks에 근거).

## 권고
- **저장소 코드(`queue/kafka.go`, `deploy/regions.yaml`, `main.go`)를 확인하기 전까지 RabbitMQ `settlement_jobs`에 바인딩하지 말 것.** 이는 Known Decisions에만 근거하는데, 해당 슬롯은 문서 내 최신 Fact 정보와 모순되며 오래된 관찰일 가능성이 크다.
- 우선순위: `queue/kafka.go`를 열어 실제 사용 중인 Kafka 토픽명/설정을 확인하고, 그 토픽에 바인딩하는 방향으로 진행. 동시에 이 정산 컨슈머 서비스가 리전별로 별도 배포/바인딩이 필요한지 `deploy/regions.yaml`로 확인.
- 이 문서 자체의 모순(Known Decisions vs Fact 섹션)은 `kervo compile` 재실행 시점에 stale 슬롯으로 정리되어야 할 항목으로 보이며, 팀에 보고할 가치가 있음.
