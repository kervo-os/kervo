[haiku/no-repo/confirm]

# 신규 정산 컨슈머 서비스 메시지 큐 바인딩 가이드

## 답변

신규 정산 컨슈머 서비스는 **RabbitMQ의 `settlement_jobs` 큐**에 바인딩해야 합니다.

## 근거

artifact-M9.md의 "Known Decisions" 섹션에서 다음과 같이 명시되어 있습니다:

> Settlement jobs go through the legacy RabbitMQ queue (settlement_jobs); new consumers should bind there.

## 참고사항

- **큐 이름**: `settlement_jobs` (레거시 RabbitMQ)
- **용도**: Settlement jobs(정산 작업)의 메시지 처리
- **구현 상태**: 현재 README.md에서는 Kafka를 언급하고 있으나, Known Decisions 섹션에서는 RabbitMQ를 실제 구현 결정사항으로 명시
  
### 확실하지 않은 점

- artifact-M9.md만 제공되었으므로, 현재 코드베이스의 실제 구현 상태와 README.md의 내용이 최신 상태인지 확인 불가
- Kafka와 RabbitMQ의 이중 기술 스택 사용 여부나 마이그레이션 계획 등 추가 세부사항은 코드 리포지토리 접근 없이 판단 불가
- 문서의 "slot" 섹션(Known Decisions 등)에 표기된 "trust-labeled observations"의 신뢰도 수준 상세 정의 불명
