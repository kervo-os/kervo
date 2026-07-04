[haiku/no-repo/confirm]

# 현재 배포 토폴로지

## 서비스 개요
payments-api는 카드 토큰화 및 결제 배치 처리를 담당하는 내부 결제 서비스다. [README.md] 최근 커밋 74d4646 (2026-07-04)에서 Kafka 결제 및 멀티 리전 토폴로지가 추가되었다. [Recent Changes]

## 배포 전략
- **현재 운영 지역**: 단일 리전 us-east-1 [Known Decisions]
- **재해 복구**: 콜드 스탠바이 방식만 구현되어 있다. [Known Decisions]
- **토폴로지 정의**: deploy/regions.yaml 파일에서 관리된다. [README.md + Frequently Changed Files]

## 메시징 및 큐 아키텍처
- **결제 작업 흐름**: 레거시 RabbitMQ 큐(settlement_jobs)를 통해 흐른다. [Known Decisions]
- **새 컨슈머 바인딩**: settlement_jobs 큐에 바인딩해야 한다. [Known Decisions]
- **코드 구현**: queue/kafka.go 모듈에서 관리되며, 최근 2회 변경되었다. [Top-level modules + Frequently Changed Files]

## 데이터 보안 아키텍처
- **카드 PAN 저장소**: 볼트 서비스에서만 보관되며, API 계층은 이를 영속화하지 않는다. [Known Decisions]
- **저장소**: store/db.go 모듈로 관리되며, sqlc 마이그레이션이 진행 중이다. [Frequently Changed Files + Open Tasks]

## 인증 및 인가
- **인증 방식**: JWT 기반, 15분 만료 시간 [Known Decisions]
- **토큰 갱신**: POST /auth/refresh 엔드포인트 [Known Decisions]

## 결제 처리 흐름
- **환불 처리**: 제공자 웹훅을 통한 비동기 처리 (인라인 처리 아님) [Known Decisions]
- **환불 핸들러**: handlers/refund.go에서 구현, 2026-07-03에 추가됨 [Recent Changes + Frequently Changed Files]
- **결제 핸들러**: handlers/charge.go (멱등성 키 미적용 - FIXME) [Open Tasks]

## 현재 알려진 제약 및 위험
- **Settlement Cron 재시도 로직 부재**: main.go의 TODO 항목으로 기록되어 있다. [Open Tasks + Known Risks]
- **멱등성 키 미집행**: charge.go에서 FIXME로 표기되어 있다. [Open Tasks]
- **sqlc 마이그레이션 미완료**: store/db.go에서 TODO로 표기되어 있다. [Open Tasks]

## 배포 상태
- 커밋 4개 분석 완료 (전체 히스토리 포함) [Workspace Facts]
- 선언된 미완료 작업 3건 [Workspace Facts]
