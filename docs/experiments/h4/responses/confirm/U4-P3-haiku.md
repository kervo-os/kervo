[haiku/no-repo/confirm]

# 환불 처리 방식 — 팀 이해 정리

## 확정된 진술 (Verified)

### 1. 비동기 처리 방식
**진술**: 환불은 비동기적으로 provider webhooks를 통해 처리되며, inline 처리가 아니다.

- **확신도**: 높음 (95%)
- **근거**: Known Decisions에서 [verified — human:kim] 표시
- **주의**: ⚠ conflict 플래그 존재 — 다른 곳에서 상충하는 정보 있을 가능성

---

## 구현 신호 (Low-Medium Confidence)

### 2. 환불 핸들러 존재
**진술**: `handlers/refund.go` 환불 핸들러가 구현되어 있다.

- **확신도**: 중간 (70%)
- **근거**: Recent Changes에 "bb2a9d7 2026-07-03 add refund handler" 기록
- **제약**: 파일 내용을 직접 확인하지 않았음 — 실제 구현 방식 불명

---

## 불명확한 영역

### 3. Webhook 상세 메커니즘
**확인되지 않은 사항**:
- 어느 provider의 webhook인가? (Stripe, PayPal 등)
- 재시도 로직이 있는가?
- 실패한 webhook에 대한 처리 방식은?
- 멱등성(idempotency) 보장 방식은?

**관련 열린 작업**: `handlers/charge.go:2 — FIXME: idempotency keys not enforced` — 충전과 관련된 멱등성 미흡이 있으므로, 환불도 유사한 문제 가능성

---

## 종합 진단

**확실한 것**: 비동기 webhook 기반 처리 (verified)  
**추정 구현**: 최근 핸들러 추가됨 (7월 3일)  
**확인 필요**: handlers/refund.go 파일 검토 및 conflict 원인 파악  
**위험 신호**: idempotency 미구현 상태 (전체 핸들러)
