[sonnet/no-repo/confirm]

# 재해복구(DR) 문서 — "현재 배포 토폴로지" 절 (초안)

전제: 저장소 코드에는 접근하지 못했고, `/Users/[user]/Developer/30_lab/EXPER/h4-kit/artifact-P3.md`(kervo 자동 생성 컨텍스트 아티팩트) 한 건만 근거로 작성함. 이 아티팩트는 "Fact" 절(저장소 스캔 기반, 결정론적)과 "Slot" 절(신뢰 라벨이 붙은 관찰 — verified/generated/observed/stale)이 혼재되어 있으므로, 아래 각 진술마다 근거와 신뢰 라벨을 표기한다. 문서에 없는 내용은 추정하지 않고 "불명"으로 남긴다.

## 1. 서비스 개요

- payments-api(저장소명 h4-repo)는 카드 토큰화 및 정산(settlement) 배치를 담당하는 내부 결제 서비스다. [근거: Fact — README.md 발췌]
- 정산 작업은 Kafka(`queue/` 모듈)를 통해 흐르며, 멀티 리전으로 배포되고 리전 설정은 `deploy/regions.yaml`에 있다고 README에 기술되어 있다. [근거: Fact — README.md 발췌]
  - 단, 이는 README의 서술일 뿐이며 `deploy/regions.yaml`의 실제 내용(리전 목록, 리전별 역할, 액티브-액티브/액티브-패시브 여부 등)은 이 아티팩트에 포함되어 있지 않다. **몇 개 리전에, 어떤 방식으로 배포되는지는 확인 불가.**
- 언어/프레임워크: Go(표준 라이브러리 수준으로 추정, 별도 웹 프레임워크명 언급 없음). [근거: Fact — Repository Summary, Languages/Frameworks]

## 2. 최근 변경 이력이 시사하는 토폴로지 전환

- 분석된 커밋은 총 4개이며 "complete"로 표시되어, 이 4개가 저장소 전체 이력일 가능성이 높다(초기 단계 리포지토리로 추정). [근거: Fact — Workspace Facts, "Commits analyzed: 4 (complete)"]
- 그중 가장 최근 커밋(`74d4646`, 2026-07-04)의 메시지가 "kafka settlement + multi-region topology"다. [근거: Fact — Recent Changes]
  - 이는 Kafka 기반 정산과 멀티 리전 토폴로지가 매우 최근(문서 작성 시점과 근접한 날짜)에 도입되었을 가능성을 시사한다. 즉 이 토폴로지가 실제 프로덕션에 완전히 정착되었는지, 아직 반영 중인지는 **확실하지 않음**.
- `deploy/regions.yaml`, `queue/kafka.go`는 최근 변경 파일 목록에 각각 1회만 등장한다(4개 커밋 중 1개 커밋에서만 수정됨). [근거: Fact — Frequently Changed Files]

## 3. 정산(Settlement) 파이프라인

- 현재: 정산 작업은 Kafka를 통해 처리된다(`queue/kafka.go`). [근거: Fact — README.md 발췌 및 Recent Changes 커밋 메시지]
- 과거(레거시, 현재는 폐기): 정산 작업이 RabbitMQ 큐(`settlement_jobs`)를 통해 처리되던 시기가 있었으나, 2026년 6월 Kafka로 전환되며 이 내용은 문서 스스로 "폐기(stale)"로 태깅되어 있다. [근거: Deprecated/Stale Notes — **stale, system** 라벨. "superseded by code change: settlement moved to Kafka (queue/kafka.go, June 2026)"]
  - DR 문서에 RabbitMQ 관련 레거시 절차·runbook이 남아있다면 더 이상 유효하지 않을 가능성이 높으며, Kafka 기반 절차로 대체가 필요하다.
- 미해결 이슈: `main.go`에 "정산 크론(settlement cron) 연결" TODO가 남아 있다. [근거: Open Tasks — main.go:2]
  - 정산 크론에 재시도(retry) 로직이 없을 수 있다는 관찰이 있으나, 이는 **[근거: Known Risks — generated, backend:gpt-oss-120b (사람 검증 아님, 신뢰도 낮음으로 취급)]**이다.
  - 종합하면, 정산 파이프라인의 스케줄링/재시도 계층이 완성되지 않았을 가능성이 있다 — DR 관점에서 리전 장애 시 정산 배치가 자동으로 재개되는지는 **확실하지 않음**.

## 4. 인증 및 시크릿/카드 데이터 처리

- 카드 PAN은 별도의 vault 서비스에만 저장되며, API 레이어는 PAN을 영속화하지 않는다. [근거: Known Decisions — **verified, human:kim**]
  - 토폴로지 상 vault가 payments-api와 분리된 별도 컴포넌트/서비스로 존재함을 시사하나, vault의 배포 위치(동일 리전 내부/전용 격리 리전 여부, 이중화 방식)는 문서에 없음 — 확인 불가.
- 인증은 JWT(만료 15분), 갱신은 `POST /auth/refresh`. [근거: Known Decisions — **verified, human:kim**]
- 환불(refund)은 프로바이더 웹훅을 통해 비동기 처리되며 인라인 처리가 아니다. [근거: Known Decisions — **verified, human:kim**, 단 문서에 ⚠ 충돌(conflict) 플래그가 붙어 있음]
  - 이 항목은 문서 내에서 다른 어떤 항목과 상충하는지 구체적으로 드러나 있지 않다. **이 결정 사항의 최종 신뢰도는 불확실**하므로 DR 문서에 그대로 확정 사실처럼 반영하지 말고 별도 확인이 필요.

## 5. 알려진 갭 / 확인 필요 사항 (DR 관점)

다음은 이 아티팩트만으로는 답할 수 없는, DR 문서 완성에 필요한 핵심 질문들이다. 추정하지 않고 명시적으로 "불명"으로 남긴다.

- 정확히 몇 개 리전에, 어느 지역에 배포되는가 — 불명 (`deploy/regions.yaml` 실제 내용 미포함)
- 액티브-액티브 구성인지 액티브-패시브(failover) 구성인지 — 불명
- 데이터베이스/스토리지 계층의 복제·백업 전략 — 불명 (`store/db.go`는 sqlc 마이그레이션 TODO만 확인됨). [근거: Open Tasks — store/db.go:2]
- 배포 자동화 방식(CI/CD, IaC 도구, 오케스트레이터) — 불명. [근거: Fact — Commands, "No declared commands found (Makefile targets, package.json scripts)"]
- 목표 RTO/RPO, 페일오버 절차, 헬스체크/모니터링 구성 — 문서에 전혀 언급 없음, 불명
- 멀티 리전·Kafka 전환이 프로덕션에 안정적으로 반영된 상태인지, 진행 중인 변경인지 — 불명(같은 날짜대의 매우 최근 커밋이라는 정황만 있음)
- idempotency key 미적용(`handlers/charge.go` FIXME)이 리전 간 페일오버 시 중복 과금 위험으로 이어지는지 — 문서상 직접적 근거 부족, 추정 금지. [근거: Open Tasks — handlers/charge.go:2, "idempotency keys not enforced"]

## 근거 신뢰도 요약

- **Fact (결정론적, 저장소 스캔 기반)**: README 발췌, Recent Changes, Frequently Changed Files, Open Tasks, Workspace Facts, Repository Summary — 원문 인용이며 가장 신뢰도 높음. 단, 이를 바탕으로 한 해석·추론 문장은 별도로 "확실하지 않음"으로 구분 표기함.
- **Verified — human:kim**: PAN/vault 분리 저장, JWT 인증, 환불 비동기 처리(단 conflict 플래그 있음) — 사람이 검증했다고 표시되었으나, 검증 절차 자체는 문서에 드러나지 않음.
- **Generated — backend:gpt-oss-120b**: 정산 크론 재시도 부재 리스크 — AI가 생성한 추정치이며 사람 검증 표시 없음. 신뢰도 낮음으로 취급.
- **Observed — human:lee**: 테이블 기반 테스트 컨벤션 — 배포 토폴로지와 직접 관련 없어 본 절에서는 제외함.
- **Stale — system**: RabbitMQ 레거시 큐 서술 — 현재는 무효, Kafka로 대체됨. DR 문서/runbook 갱신 시 참고.
