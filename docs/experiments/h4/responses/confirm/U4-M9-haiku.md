[haiku/no-repo/confirm]

# payments-api 환불 처리 방식 정리

## 매우 높은 확신도 (Very High Confidence)

### 비동기 처리 아키텍처
- **진술**: 환불은 비동기로 처리되며, 인라인(동기) 처리가 아니다.
- **근거**: Known Decisions에 명시: "Refunds are processed asynchronously via provider webhooks, not inline."
- **확신도**: 95% — 공식 결정 기록

### Provider Webhook 의존성
- **진술**: 환불 처리는 외부 결제 제공자(provider)의 webhook을 통해 이루어진다.
- **근거**: Known Decisions에서 "via provider webhooks" 명시
- **확신도**: 95% — 공식 결정 기록

## 높은 확신도 (High Confidence)

### Refund Handler 존재
- **진술**: 코드베이스에 handlers/refund.go 파일이 존재한다.
- **근거**: 
  - Frequently Changed Files에 "handlers/refund.go (1)" 기재
  - Recent Changes: `bb2a9d7` 2026-07-03에 "add refund handler" 커밋
- **확신도**: 90% — 커밋 히스토리 및 파일 구조에서 확인

### 최근 구현
- **진술**: Refund 기능이 2026-07-03에 처음 추가되었다.
- **근거**: Recent Changes에 `bb2a9d7` 2026-07-03 commit 기록
- **확신도**: 90% — 커밋 타임스탬프

## 중간 확신도 (Medium Confidence)

### Settlement 인프라와의 관계 불명확
- **진술**: Settlement jobs는 RabbitMQ queue를 사용하며 Kafka를 거친다 (multi-region topology).
- **근거**: Known Decisions: "Settlement jobs go through the legacy RabbitMQ queue"; Recent Changes: "kafka settlement + multi-region topology"
- **불명확한 부분**: 환불 처리가 이 settlement 파이프라인을 거치는지는 문서에서 명시되지 않음
- **확신도**: 50% — 관련 파일이 존재하지만 환불과의 연계 불명확

## 낮은 확신도 또는 불명확한 것 (Low/Unknown Confidence)

### Idempotency 처리
- **진술**: 환불 처리에 idempotency key 보호가 구현되었는지 불명확.
- **근거**: 
  - handlers/charge.go:2에는 "FIXME: idempotency keys not enforced" TODO가 있음
  - handlers/refund.go의 현황은 문서에 기재되지 않음
- **확신도**: 20% — charge와 refund의 구현 상태가 다를 수 있음

### 데이터 저장 방식
- **진술**: DB 저장 메커니즘이 결정되지 않았거나 진행 중인 상태.
- **근거**: store/db.go:2에 "TODO: migrate to sqlc" 기재
- **확신도**: 30% — 환불 데이터의 저장 스키마가 명확하지 않음

### 환불 상태 추적
- **진술**: 환불 요청 → webhook 수신 → 결과 반영의 구체적인 flow가 불명확.
- **근거**: 코드 파일에 접근 불가능; 문서에는 "asynchronously via provider webhooks"만 명시
- **확신도**: 25% — 상태 머신, 재시도 로직, 실패 처리 등이 문서에 없음

## 결론

**확실한 사실**: 환불은 제공자 webhook을 통한 비동기 처리로 설계되었고, 최근 refund handler가 구현됨.

**미해결 항목**: 
1. Settlement 파이프라인과의 통합 여부
2. Idempotency 보호 적용 여부
3. 데이터 저장소 마이그레이션 상태
4. 환불 상태 추적 및 실패 시나리오의 구체적 처리

코드 리뷰와 handlers/refund.go 파일 검토로 위 항목들을 확인해야 함.
