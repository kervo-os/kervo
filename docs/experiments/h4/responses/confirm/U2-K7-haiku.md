[haiku/no-repo/confirm]

# 현재 배포 토폴로지

## 개요

payments-api는 다중 리전으로 배포되는 내부 결제 서비스입니다. 카드 토큰화 및 결제 배치 처리를 수행하며, 결제 작업은 Kafka를 통해 흐릅니다.

**근거**: README.md excerpt — "Internal payments service: card tokenization and settlement batching. Settlement jobs flow through Kafka (queue/); deployed multi-region (deploy/regions.yaml)." / 최근 커밋 74d4646 (2026-07-04) — "kafka settlement + multi-region topology"

## 구성 요소

### 핵심 서비스 계층
- **payments-api (주 서비스)**: 카드 결제 요청 수신, 리팬드 처리, JWT 인증 (15분 만료) 제공
  - **근거**: Recent Changes (refund handler, initial service skeleton) / Known Decisions (JWT with 15-minute expiry)

### 메시징 레이어
- **Kafka**: Settlement 작업 큐 처리
  - **근거**: README.md excerpt, 모듈 queue/kafka.go (Frequently Changed Files)
  
- **주의**: Deprecated/Stale Notes에 따르면 legacy RabbitMQ queue (settlement_jobs)도 존재할 가능성이 있음. 현재 문서에서는 Kafka가 신규 토폴로지이지만, 마이그레이션 상태 불명확

### 외부 의존성
- **Vault Service**: 카드 PAN(Primary Account Number) 저장 (payments-api는 보유하지 않음)
  - **근거**: Known Decisions — "Card PANs are stored only in the vault service; the API layer never persists them."

- **Payment Provider**: 리팬드 처리 (비동기 webhook을 통해 결과 수신)
  - **근거**: Known Decisions — "Refunds are processed asynchronously via provider webhooks, not inline."

## 지역별 배포

다중 리전 배포 구성 정의: `deploy/regions.yaml`

**근거**: README.md, Frequently Changed Files — deploy/regions.yaml (1), Recent Changes — "kafka settlement + multi-region topology"

**주의**: 구체적인 리전 목록, 각 리전의 가용성 영역, 리소스 구성은 deploy/regions.yaml 파일 접근 불가로 문서에서 확인 불가능함.

## 알려진 제약 및 위험

- **Settlement 크론 재시도 로직 부재**: main.go의 TODO 주석 — "wire settlement cron retry"
  - **근거**: Open Tasks, Known Risks — "Settlement cron appears to lack retry logic"
  - **영향**: 재해복구 계획 수립 시 settlement 작업 실패 시나리오 대응 필수

## 문서 제약사항

본 절은 artifact-K7.md (자동 생성 컨텍스트)만을 근거로 작성되었으며, 다음 정보는 저장소 코드 접근 불가로 확인 불가:
- 각 리전의 구체적 이름, 위치, 인프라 사양
- Kafka 클러스터 구성 (파티션 수, 복제 계수, broker 수)
- 데이터베이스 토폴로지 (primary/replica 배치, failover 구성)
- 정확한 네트워크/보안 그룹 구성
- 현재 활성 버전과 실제 배포 상태 (Recent Changes의 마지막 커밋 기준은 2026-07-04)
