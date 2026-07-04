[sonnet/no-repo/confirm]

# 현재 배포 토폴로지 (재해복구 문서용 초안)

## 0. 전제 및 신뢰도 안내

이 절은 저장소 코드나 실제 인프라 설정에 직접 접근하지 않고, 자동 생성된 컨텍스트 문서 `artifact-M9.md` (h4-repo, kervo compile 산출물) 한 건만을 근거로 작성했다. 해당 문서 자체가 "Fact sections are deterministic; slot sections carry trust-labeled observations" 라고 명시하고 있으므로, 아래에서는 출처를 ①결정론적 Fact 섹션(저장소 스캔 결과)과 ②Slot 섹션(신뢰도 표시가 있는 관찰/추정)으로 구분해 표시했다. 또한 "Possible Current Goal" 슬롯에 "A confirmed goal becomes the first Verified observation"라는 문구가 있는데, 이는 현재 어떤 슬롯 항목도 아직 "Verified" 라벨을 받지 못했음을 시사한다. 즉 Slot 섹션(Known Decisions, Known Risks, Doc Summaries)의 내용은 전부 미검증 상태로 취급해야 한다.

**이 절을 DR 문서에 반영하기 전, 아래 "핵심 미해결 충돌"(§2)을 반드시 사람이 확인해야 한다.** 문서 내부에 서로 모순되는 토폴로지 진술이 존재하며, 이 아티팩트만으로는 어느 쪽이 현재 운영 중인 실제 상태인지 판단할 수 없다.

## 1. 확인 가능한 사실 (결정론적 Fact 섹션 근거)

- 서비스는 카드 토큰화 및 정산(settlement) 배치를 담당하는 내부 결제 서비스다. [근거: README.md excerpt]
- 정산 작업은 `queue/` 디렉터리 코드를 통해 큐를 거친다. [근거: README.md excerpt — "Settlement jobs flow through Kafka (queue/)"] 다만 이 문구가 "Kafka"라고 명시한 것과, §2의 Known Decisions 슬롯이 "legacy RabbitMQ"라고 명시한 것이 상충한다(§2 참조).
- 배포는 "multi-region"이며 관련 설정은 `deploy/regions.yaml`에 있다고 README가 명시한다. [근거: README.md excerpt — "deployed multi-region (deploy/regions.yaml)"] 이 역시 Known Decisions 슬롯의 "single-region in us-east-1" 진술과 상충한다(§2 참조).
- 가장 최근 커밋(`74d4646`, 2026-07-04, 문서에서 확인 가능한 커밋 중 최신)의 메시지는 "kafka settlement + multi-region topology"이다. [근거: Recent Changes]
- 최근 변경 파일 목록에 `queue/kafka.go`(1회)와 `deploy/regions.yaml`(1회)이 포함되어 있어, 위 최신 커밋이 큐/리전 토폴로지와 관련된 파일을 직접 건드렸을 가능성이 높다. [근거: Frequently Changed Files] 다만 이 파일들의 실제 내용(브로커 목록, 리전 목록, 활성-활성/활성-대기 여부 등)은 아티팩트에 포함되어 있지 않아 확인 불가.
- 저장소 최상위 관련 모듈은 `deploy/`, `handlers/`, `queue/`, `store/` 4개로 파악된다. [근거: Related Modules, Workspace Facts]
- 빌드/배포/실행 관련 선언된 커맨드(Makefile, package.json 등)는 발견되지 않았다. [근거: Commands 섹션 — "No declared commands found"] 따라서 배포 파이프라인, 컨테이너화 방식, 오케스트레이션 도구(K8s 여부 등)는 이 아티팩트로는 전혀 알 수 없다.
- `main.go`에 "TODO: wire settlement cron"이 남아있다. [근거: Open Tasks] 즉 정산 작업을 주기적으로 트리거하는 크론/스케줄러가 아직 코드상 연결되지 않았을 가능성이 있으며, 이는 DR 시나리오에서 정산 배치의 자동 재개 여부에 직접 영향을 줄 수 있는 사실이다.
- `handlers/charge.go`에 "FIXME: idempotency keys not enforced"가 남아있다. [근거: Open Tasks] 페일오버/재시도 시 중복 청구(double charge) 위험과 직결되므로 DR 절차(특히 리전 전환 시 in-flight 요청 재시도)에서 리스크로 반드시 언급해야 한다.
- `store/db.go`에 "TODO: migrate to sqlc"가 남아있다. [근거: Open Tasks] DB 접근 계층이 과도기 상태임을 시사하나, DB 자체의 복제/백업/리전 토폴로지에 대한 정보는 아티팩트에 없음.

## 2. 핵심 미해결 충돌 — 사람 확인 필요

Known Decisions 슬롯(미검증)은 다음을 명시한다:
- "Settlement jobs go through the legacy RabbitMQ queue (settlement_jobs); new consumers should bind there." [근거: Known Decisions 슬롯]
- "Payments runs single-region in us-east-1; DR is cold-standby only." [근거: Known Decisions 슬롯]

그런데 §1의 결정론적 Fact 근거(README 요약, 최신 커밋 메시지, 최근 변경 파일)는 정반대로 **Kafka + 멀티리전** 전환을 가리킨다. 두 진술은 동시에 참일 수 없다(적어도 "현재" 시점 기준으로는). 가능한 해석은 다음과 같으며, 이 아티팩트만으로는 어느 것이 맞는지 구분할 수 없다.

1. Known Decisions 슬롯이 **과거(마이그레이션 이전) 상태를 반영한 오래된(stale) 관찰**이고, `74d4646` 커밋으로 RabbitMQ/단일 리전 → Kafka/멀티 리전 전환이 이미 (적어도 코드 레벨에서는) 이루어졌다.
2. `74d4646` 커밋이 아직 main 브랜치에만 merge되었을 뿐 **운영 환경에 배포되지 않았고**, 실제 프로덕션은 여전히 RabbitMQ + us-east-1 단일 리전 + 콜드 스탠바이 DR 상태다.
3. 두 설명이 부분적으로만 겹치는 과도기 상태(예: 신규 트래픽은 Kafka로, 레거시 컨슈머는 RabbitMQ로 병행 운영)일 수 있으나, 이를 뒷받침할 근거는 없다.

주목할 점은, 이런 명백한 모순이 있음에도 "Deprecated / Stale Notes" 섹션은 "None recorded"로 비어 있다는 것이다. [근거: Deprecated/Stale Notes] 즉 아티팩트 생성기 자신도 이 충돌을 stale로 플래그하지 못한 상태이며, 이는 아티팩트의 한계로 별도 기록해 둘 필요가 있다.

**결론: 재해복구 문서의 "현재 배포 토폴로지"를 Kafka/멀티리전으로 볼지, RabbitMQ/단일리전·콜드스탠바이로 볼지는 이 문서만으로 확정할 수 없다. 실제 운영 환경(배포 매니페스트, 인프라 콘솔, 온콜 담당자 확인 등)을 통해 반드시 재확인해야 한다.** DR 계획(특히 RTO/RPO, 페일오버 절차)은 이 두 토폴로지에서 완전히 다르게 설계되어야 하므로, 확정 전에는 문서에 두 가능성을 모두 병기하고 "미확정"으로 표기하는 것을 권장한다.

## 3. 아티팩트로는 확인 불가한 항목

- 실제 클라우드/인프라 제공자, 오케스트레이션 방식(K8s/VM/서버리스 등) — Commands 섹션 공란으로 근거 없음.
- `deploy/regions.yaml`의 실제 내용(리전 목록, active-active 여부 등) — 파일이 최근 변경되었다는 사실만 확인되고 내용은 아티팩트에 없음.
- `queue/kafka.go`의 실제 내용(브로커, 토픽, 파티셔닝, 컨슈머 그룹) — 동일하게 내용 미포함.
- RabbitMQ(`settlement_jobs`)가 실제로 폐기(retire)되었는지, 아직 병행 운영 중인지 — Known Decisions 슬롯 자체가 미검증이라 근거로 삼기 부족.
- RTO/RPO 목표치 — 문서 어디에도 언급 없음.
- Vault 서비스(카드 PAN 보관)의 배포 토폴로지 — "Card PANs are stored only in the vault service" [근거: Known Decisions 슬롯, 미검증]라는 언급만 있고 vault 자체의 리전/이중화 구성은 불명.
- DB(`store/`) 복제·백업 리전 구성 — sqlc 마이그레이션 TODO 외 정보 없음.
- 위 `74d4646` 커밋의 실제 배포/릴리스 상태(스테이징만인지 프로덕션까지 반영됐는지) — 커밋 로그만으로는 판단 불가.

## 4. DR 관점에서 추가로 짚어야 할 리스크 (미검증 관찰 포함)

- 정산 크론이 아직 코드상 wiring되지 않은 상태(§1)로 보이며, Known Risks 슬롯은 이를 "retry logic 부재" 리스크로 해석한다. [근거: Known Risks 슬롯 — "Settlement cron appears to lack retry logic — main.go carries a TODO"] 페일오버 후 정산 배치가 자동으로 재개되는지, 수동 개입이 필요한지 확인 필요.
- 청구(charge) 처리에 idempotency key가 강제되지 않는 상태([근거: Open Tasks])이므로, 리전 전환/재시도 시 중복 청구 가능성을 DR 런북에 명시적 경고로 남겨야 한다.
- 환불은 프로바이더 웹훅을 통한 비동기 처리라는 진술이 있다. [근거: Known Decisions 슬롯, 미검증] 사실이라면 DR 전환 중 웹훅 수신 불가 구간에 대한 재시도/재전송 정책이 토폴로지 설계에 포함되어야 하나, 이 아티팩트로는 웹훅 인프라의 이중화 여부를 알 수 없다.

---
(본 절은 코드 저장소에 대한 직접 접근 없이 `artifact-M9.md` 한 건만을 근거로 작성되었으며, 특히 §2의 토폴로지 충돌은 반드시 실제 운영 환경 확인 후 확정해야 한다.)
