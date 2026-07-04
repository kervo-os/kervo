[haiku/no-repo/confirm]

## 신규 정산 컨슈머 서비스 메시지 큐/토픽 바인딩 가이드

### 답변: Kafka 사용 (구체적 토픽명은 문서에 미기재)

#### 근거

1. **README 요약 (사실 섹션)**
   - "Settlement jobs flow through Kafka (queue/)"
   - 정산 작업이 Kafka를 통해 흐름을 명시

2. **최근 커밋 이력 (결정적 증거)**
   - `74d4646` (2026-07-04): "kafka settlement + multi-region topology"
   - 가장 최근 커밋에서 Kafka 정산 토폴로지 구현됨

3. **코드 모듈 구조**
   - queue/kafka.go 파일이 자주 변경됨 (빈도: 1회)
   - queue/ 모듈이 Kafka 설정 담당

4. **마이그레이션 상황**
   - Stale Notes (오래된 정보)에는 "Settlement jobs go through the legacy RabbitMQ queue (settlement_jobs)"로 표시
   - 레거시 지정으로, 현재 상태가 아님을 명시

#### 결론

- **사용 기술**: Kafka (RabbitMQ 레거시 아님)
- **참고 파일**: queue/kafka.go (Kafka 클라이언트/토폴로지 구성)
- **배포**: deploy/regions.yaml에 멀티 리전 배포 설정 포함

#### 확실하지 않은 전제

- **구체적 토픽명**: 문서에 명시되지 않음
  - "settlement"로 시작할 것으로 예상되나, 정확한 토픽명은 queue/kafka.go 또는 CLAUDE.md 검토 필요
  - 파티션/컨슈머 그룹명도 동일

### 다음 단계

코드 저장소 접근 가능 시:
1. queue/kafka.go에서 Kafka 설정 및 토픽명 확인
2. CLAUDE.md에서 정산 메시징 아키텍처 문서 검토
3. deploy/regions.yaml에서 토픽별 리전 토폴로지 확인
