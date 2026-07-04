[haiku/no-repo/confirm]

# 게이트웨이 인증 미들웨어: 토큰 정책 기술

## 명시된 정책

### Access Token (JWT)
- **만료 시간**: 15분 [**High confidence**]
  - 문서의 "Known Decisions" 섹션에 명시: "Auth is JWT with 15-minute expiry"
  - 게이트웨이 미들웨어는 이 만료 시간을 검증해야 함

- **갱신 방식**: POST /auth/refresh 엔드포인트 [**High confidence**]
  - 문서 명시: "refresh via POST /auth/refresh"
  - 클라이언트가 만료된 토큰 대신 이 엔드포인트로 새 토큰을 요청하는 방식으로 추정

### Refresh Token
- **별도 정책**: 명확한 정보 부재 [**High confidence**]
  - 문서에는 access token 정책만 기술되어 있고, refresh token의 만료 시간, 저장 방식, 회전 정책이 기술되지 않음
  - 미들웨어 구현 시 이 부분은 별도 설계 문서 또는 코드 검토 필요

## 미들웨어 구현 전제

### 확인된 요구사항
1. **Token 검증**: 모든 요청에서 JWT의 서명 및 만료 여부 검증 필요 [**High confidence**]
   - 결제 서비스(card tokenization, settlement batching)이므로 인증 요구사항이 높음
   - Card PAN이 vault service에서만 관리되는 점에서 보안 중시 아키텍처 확인

2. **만료 토큰 처리**: 만료된 JWT는 거부하고 클라이언트가 /auth/refresh로 갱신하도록 유도 [**High confidence**]
   - 15분 만료 정책 + refresh 엔드포인트 존재로부터 추론

### 불확실한 영역
- **Refresh token 갱신**: 갱신 시 새로운 refresh token 발급 여부 [**Low confidence**]
  - 문서에 기술 없음
  
- **Rate limiting**: 갱신 엔드포인트에 대한 rate limiting 정책 [**Low confidence**]
  - 문서에 기술 없음

- **Token 저장소**: Refresh token 저장 위치 및 방식 (Redis, DB, 메모리 등) [**Low confidence**]
  - 문서에 기술 없음

- **미들웨어 배치 위치**: 모든 엔드포인트에 적용 또는 특정 경로만 [**Low confidence**]
  - 문서에 기술 없음. 결제 관련 handlers/(charge.go, refund.go)는 보호 필요로 추정되지만, /auth/refresh 자체는 보호 대상 외로 추정

## 권장 사항

다음 항목들은 구현 전 팀과 확인 필요:
1. Refresh token의 TTL 및 회전 정책
2. Token 저장소 선택 (Redis vs DB)
3. /auth/refresh 엔드포인트의 rate limiting 정책
4. 미들웨어가 적용되지 않아야 할 public 엔드포인트 목록
5. Token 갱신 중 요청 동시성 처리 방식
