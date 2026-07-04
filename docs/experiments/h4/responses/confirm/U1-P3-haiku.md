[haiku/no-repo/confirm]

# 신규 정산 컨슈머 서비스 메시지 큐 바인딩 가이드

## 결론
신규 정산 컨슈머 서비스는 **Kafka**의 settlement-related 토픽에 바인딩해야 합니다.

## 근거

### 1. 현재 아키텍처
artifact-P3.md의 README 요약에 명시:
> "Internal payments service: card tokenization and settlement batching. **Settlement jobs flow through Kafka (queue/)**; deployed multi-region (deploy/regions.yaml)."

### 2. 최신 변경 이력
- 2026-07-04 커밋: `kafka settlement + multi-region topology` 
- 구현 파일: `queue/kafka.go`

### 3. Stale Notes 확인
문서의 Deprecated 섹션에서 확인할 수 있는 정보:
```
[stale — superseded by code change: settlement moved to Kafka (queue/kafka.go, June 2026)]
Settlement jobs go through the legacy RabbitMQ queue (settlement_jobs); 
new consumers should bind there.
```

이는 **과거에는 RabbitMQ의 `settlement_jobs` 큐를 사용했으나, 2026년 6월에 Kafka로 이동**했음을 의미합니다.

## 확실하지 않은 부분
- **구체적인 Kafka 토픽 이름**: 문서에 토픽명이 명시되지 않았습니다. `queue/kafka.go` 파일을 참조하여 실제 토픽명(예: `settlement-jobs`, `settlements`, `settlement-events` 등)을 확인해야 합니다.
- **파티션 전략 및 consumer group 설정**: 마찬가지로 `queue/kafka.go` 구현을 확인 필요합니다.

## 추천 다음 단계
1. 레포지토리의 `queue/kafka.go` 파일 검토 → Kafka 토픽명 확인
2. `deploy/regions.yaml` 확인 → 지역별 토픽/클러스터 설정 확인
3. 기존 consumer 구현 패턴 참고 → handler 스타일 및 error handling 확인
