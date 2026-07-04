[haiku/no-repo/confirm]

# 현재 배포 토폴로지

## 서비스 개요

payments-api는 카드 토큰화 및 정산 배치 처리를 담당하는 내부 결제 서비스입니다. *(근거: artifact-P3.md §Repository Summary — README.md excerpt: "Internal payments service: card tokenization and settlement batching")*

## 지역별 다중 배포 (Multi-Region)

서비스는 여러 지역에 배포되어 있으며, 배포 구성은 `deploy/regions.yaml` 파일에 정의되어 있습니다. *(근거: artifact-P3.md §Repository Summary — "deployed multi-region (deploy/regions.yaml)"; §Frequently Changed Files — deploy/regions.yaml (1))*

## 메시지 큐 아키텍처

정산 작업(settlement jobs)은 Kafka를 통해 처리됩니다. *(근거: artifact-P3.md §Recent Changes — "74d4646 2026-07-04 kafka settlement + multi-region topology"; §Repository Summary — "Settlement jobs flow through Kafka (queue/)")*

이는 2026년 7월 4일 최근 변경사항으로 정착된 구조이며, 이전의 RabbitMQ 큐 기반 구조는 더 이상 사용되지 않습니다. *(근거: artifact-P3.md §Deprecated / Stale Notes — "superseded by code change: settlement moved to Kafka (queue/kafka.go, June 2026)")*

## 데이터 저장소

**민감 데이터 분리:**
- 카드 PAN(Primary Account Number)은 Vault 서비스에만 저장되며, API 계층에서는 이를 보관하지 않습니다. *(근거: artifact-P3.md §Known Decisions — "[verified — human:kim] Card PANs are stored only in the vault service; the API layer never persists them.")*

**데이터베이스:**
- 결제 및 환불 데이터는 로컬 데이터베이스에 저장됩니다. (store/db.go) *(근거: artifact-P3.md §Frequently Changed Files — store/db.go (2); §Related Modules — store/ (1 files))*

## 핵심 처리 흐름

**결제 처리 (charge):**
- handlers/charge.go에서 결제 요청을 처리합니다. *(근거: artifact-P3.md §Frequently Changed Files — handlers/charge.go (1))*
- 현재 멱등성 키(idempotency keys) 강제는 미구현 상태입니다. *(근거: artifact-P3.md §Open Tasks — "handlers/charge.go:2 — FIXME: idempotency keys not enforced")*

**환불 처리 (refund):**
- 환불은 제공자 웹훅을 통해 비동기로 처리됩니다. *(근거: artifact-P3.md §Known Decisions — "[verified — human:kim] Refunds are processed asynchronously via provider webhooks, not inline.")*
- handlers/refund.go에서 환불 로직을 구현합니다. *(근거: artifact-P3.md §Frequently Changed Files — handlers/refund.go (1))*

**정산 스케줄 (Settlement Cron):**
- 정산 크론 작업의 진행 상황은 main.go에서 관리됩니다. *(근거: artifact-P3.md §Open Tasks — "main.go:2 — TODO: wire settlement cron")*
- 현재 재시도(retry) 로직이 구현되지 않았습니다. *(근거: artifact-P3.md §Known Risks — "Settlement cron appears to lack retry logic — main.go carries a TODO (wire settlement cron retry)")*

## 인증 및 보안

- 인증 메커니즘: JWT 토큰, 15분 만료 주기
- 토큰 갱신: POST /auth/refresh 엔드포인트 *(근거: artifact-P3.md §Known Decisions — "[verified — human:kim] Auth is JWT with 15-minute expiry; refresh via POST /auth/refresh")*

## 미해결 기술 부채

재해복구 계획 수립 시 고려할 미해결 사항:
- Settlement cron 재시도 로직 부재
- 멱등성 키 검증 미구현
- 데이터베이스 마이그레이션 (sqlc 미적용) *(근거: artifact-P3.md §Open Tasks)*
