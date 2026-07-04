[sonnet/no-repo/confirm]

# 재해복구(DR) 문서 — 현재 배포 토폴로지

> 작성 조건 고지: 이 절은 저장소 코드에 직접 접근하지 않고, 자동 생성된 컨텍스트 문서
> `artifact-K7.md` 단 하나만을 근거로 작성되었다. 해당 문서는 스스로를 "Fact 섹션(결정론적,
> 저장소 스캔 기반)"과 "슬롯 섹션(trust-labeled observations, 검증되지 않은 관찰)"으로
> 구분하고 있으며, 아래에서는 각 진술 옆에 어느 쪽 근거인지와 원문 섹션명을 표기했다.
> 슬롯 섹션 근거는 Fact 섹션 근거보다 신뢰도가 낮다는 점을 감안해야 한다.

## 1. 서비스 개요
- payments-api는 카드 토큰화(card tokenization)와 정산 배치(settlement batching)를
  수행하는 내부 결제 서비스다. — [Fact/README.md 발췌]
- 구현 언어는 Go. — [Fact/Repository Summary]

## 2. 정산 메시징 계층 (핵심 불확실 구간)
- README 기준 정산 작업은 Kafka(`queue/` 디렉터리)를 통해 흐른다. — [Fact/README.md 발췌]
- 가장 최근 커밋 `74d4646` (2026-07-04)의 메시지가 "kafka settlement + multi-region
  topology"로, Kafka 기반 정산 및 멀티리전 구성이 이 커밋에서 도입/변경되었을 개연성이
  높다. — [Fact/Recent Changes]
- `queue/kafka.go`가 존재하며 변경 이력에 1회 등장(분석된 4개 커밋 중 1개에서만 수정).
  — [Fact/Frequently Changed Files, Related Modules]
- **상충 발견**: "Deprecated/Stale Notes" 슬롯에는 "정산 작업은 레거시 RabbitMQ
  큐(`settlement_jobs`)를 통해 흐르며, 신규 컨슈머는 그쪽에 바인딩해야 한다"는
  README와 정반대되는 진술이 있다. — [Slot/Deprecated·Stale Notes]
  - 이 슬롯 자체가 "Deprecated/Stale"로 분류되어 있어 Kafka 전환 이전의 오래된 서술로
    추정되지만, 그 안의 "신규 컨슈머는 그쪽(RabbitMQ)에 바인딩해야 한다"는 지시문은
    현재형 운영 지침처럼 읽혀 해석이 모순적이다.
  - **결론(불확실)**: 정산 메시징이 Kafka로 완전히 전환되었는지, 아니면 RabbitMQ와
    이중 운영 중인지 이 문서만으로는 확정할 수 없다. DR 절차 작성 전 저장소의
    `queue/` 디렉터리 전체(kafka.go 외 RabbitMQ 관련 파일 존재 여부)를 반드시
    직접 확인해야 한다.

## 3. 멀티 리전 배포
- README은 `deploy/regions.yaml`을 근거로 멀티리전 배포를 명시한다. — [Fact/README.md 발췌]
- `deploy/` 모듈에 파일 1개 존재(= regions.yaml로 추정), 변경 이력 1회 — 최신 커밋
  `74d4646`과 결부된 것으로 추정된다. — [Fact/Related Modules, Frequently Changed Files]
- **불확실**: `regions.yaml`의 실제 내용(리전 목록/개수, active-active vs
  active-passive, 트래픽 라우팅·페일오버 정책)은 이 문서에 전혀 노출되어 있지
  않다. "멀티리전"이라는 선언 이상의 구체적 토폴로지는 알 수 없다.
- 이 구성이 가장 최신 커밋에서 막 도입된 것으로 보이는 만큼, 실제 운영에서 충분히
  검증된 상태인지는 근거가 없어 판단 불가(추정일 뿐, 문서상 명시적 근거 없음).

## 4. 데이터 계층
- `store/db.go` 존재, 변경 이력 2회(4개 커밋 중 2개에서 수정). — [Fact/Frequently Changed Files]
- Open Tasks의 `store/db.go:2 — TODO: migrate to sqlc` 및 커밋 `54be496`
  "stub sqlc migration" — 데이터 계층이 sqlc로의 마이그레이션 도중이며 미완료 상태로
  추정된다. — [Fact/Open Tasks, Recent Changes]
- **불확실**: 실제 DB 엔진(PostgreSQL/MySQL 등), 복제·백업 전략, 위 마이그레이션이
  DR 관점에서 어떤 상태(스키마 불일치 가능성 등)를 남기는지는 문서에 없음.

## 5. 카드 데이터 보관 (PCI 관련)
- "카드 PAN은 vault 서비스에만 저장되며, API 레이어는 이를 영속화하지 않는다."
  — [Slot/Known Decisions — 검증되지 않은 관찰]
- DR 관점 함의: vault 서비스가 payments-api와 분리된 별도 서비스로 존재한다는 뜻이나,
  그 vault 서비스의 배포 위치·가용성·이중화 여부는 이 아티팩트에 전혀 언급이 없다
  — 저장소 외부(별도 리포/별도 팀 관리) 가능성이 높음(불확실, 문서 근거 없음).

## 6. 인증 계층
- JWT 기반 인증, 만료 15분, `POST /auth/refresh`로 갱신. — [Slot/Known Decisions]
- 이 인증 기능이 payments-api 내부 구현인지 외부 IdP 연동인지는 문서에 근거 없음(불확실).

## 7. 정산 스케줄링 / 크론
- `main.go:2 — TODO: wire settlement cron` — 정산을 트리거하는 크론 작업이 아직
  배선(연결)되지 않은 것으로 보인다. — [Fact/Open Tasks]
- "정산 크론에 재시도 로직이 없어 보인다"는 서술은 위 TODO를 재인용한 것에 가깝다.
  — [Slot/Known Risks, 근거는 Open Tasks 재인용]
- **DR 관점 중요 미확정 사항**: 정산 크론이 아직 "wire"되지 않았다면, 현재 배포
  토폴로지에 정산 자동화 컴포넌트가 실제로 가동 중인지 자체가 불확실하다.

## 8. 환불 처리
- 환불은 인라인 처리가 아니라 결제 프로바이더 웹훅을 통해 비동기로 처리된다.
  — [Slot/Known Decisions]
- `handlers/refund.go` 존재(커밋 `bb2a9d7` "add refund handler"). — [Fact/Recent Changes]
- **불확실**: 웹훅 재시도/멱등성 처리 여부는 문서에 명시가 없다. `handlers/charge.go:2`의
  FIXME("idempotency keys not enforced")는 charge 핸들러에 대한 진술이며, refund
  핸들러에 동일 문제가 있는지는 추정만 가능하고 확정할 수 없다.

## 9. 배포/인프라 자동화
- Makefile, package.json 등 선언된 커맨드가 발견되지 않았다
  ("No declared commands found"). — [Fact/Commands]
- 따라서 CI/CD 파이프라인, 배포 스크립트, IaC(Terraform 등) 존재 여부는 이 문서만으로는
  전혀 알 수 없다. `deploy/regions.yaml`이 Kubernetes manifest인지, Terraform
  변수 파일인지, 자체 포맷인지도 불명.

## 10. 현재 진행 중인 작업 목표
- "Possible Current Goal" 슬롯은 비어 있다("No proposal yet. A confirmed goal
  becomes the first Verified observation.") — 즉 문서 생성 시점 기준 검증된
  현재 작업 목표 진술이 없다. — [Slot/Possible Current Goal]

## 종합: DR 문서에 반영할 확정/미확정 구분
**확정(Fact 섹션 근거)**
- Go 기반 내부 결제 서비스(카드 토큰화 + 정산 배치).
- 가장 최근 커밋에서 Kafka 기반 정산 + 멀티리전 토폴로지가 도입/변경됨.
- sqlc 마이그레이션 진행 중(미완료), 정산 크론 미배선, charge 핸들러 멱등성 키 미시행.

**상충 — 반드시 저장소 직접 확인 필요**
- 정산 메시징이 Kafka(README, Fact)인지 레거시 RabbitMQ(Stale Notes 슬롯)인지 불명.
  DR 절차(장애 시 정산 재처리 경로)를 이 상충이 해소되기 전까지 확정하면 안 됨.

**미확정(문서에 근거 없음, 확인 필요)**
- 리전 개수·위치·페일오버 전략(regions.yaml 내용).
- DB 엔진 종류, 복제/백업 방식.
- vault 서비스(카드 PAN 보관)의 배포 위치·가용성·이중화.
- 인증(JWT 발급) 서비스가 내부/외부 어디에 배포되어 있는지.
- CI/CD, IaC 존재 여부 및 배포 자동화 방식.
- 정산 크론의 실제 가동 여부(코드상 미배선 TODO 존재).

**신뢰도 주의**
- 위 "Slot" 표기 항목(Known Decisions, Known Risks, Doc Summaries, Deprecated/Stale
  Notes, Possible Current Goal)은 아티팩트 정의상 "trust-labeled observations"이며
  결정론적 저장소 스캔 결과가 아니다. DR 문서에 최종 반영하기 전 해당 저장소 코드로
  교차 검증이 필요하다.
