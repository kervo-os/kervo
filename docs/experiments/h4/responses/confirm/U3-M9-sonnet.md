[sonnet/no-repo/confirm]

# 게이트웨이 인증 미들웨어 전제: 토큰 정책(만료·갱신) 검토

**근거 문서**: `artifact-M9.md` (h4-repo, kervo 자동 생성 컨텍스트 아티팩트) 단 하나. 실제 저장소 코드(예: 미들웨어 구현 파일)는 열람하지 못했다. 이 문서 자체가 "Fact sections are deterministic; slot sections carry trust-labeled observations"라고 명시하므로, 아래에서는 어떤 섹션 출처인지에 따라 확신도를 구분했다.

## 0. 먼저 짚어야 할 메타 전제 (확신도: 낮음 — 구조적 문제)

- 이 문서에는 **게이트웨이(gateway) 컴포넌트 자체가 등장하지 않는다.** Related Modules는 `deploy/`, `handlers/`, `queue/`, `store/` 4개뿐이고 `auth/`, `middleware/`, `gateway/` 류 모듈이 없다. README 요약도 "card tokenization and settlement" 서비스라고만 하지, 게이트웨이라고 부르지 않는다. 따라서 "게이트웨이에 인증 미들웨어를 붙인다"는 과제 전제 자체가 이 저장소(h4-repo) 안의 작업인지, 별도 상위 게이트웨이 레포에서의 작업인지 이 문서만으로는 판단 불가. **확신도: 낮음(불확실) — 이 문서는 게이트웨이의 존재를 확인해주지 않는다.**
- 토큰 정책 관련 유일한 근거는 "Known Decisions" 슬롯의 한 줄뿐이며, 이를 뒷받침하는 파일 증거(예: `auth.go`, `jwt.go`, `middleware.go`)가 Frequently Changed Files나 Related Modules 어디에도 없다. 즉 이 결정 항목은 코드 변경 이력으로 교차 검증되지 않는다. **확신도: 낮음 — 단일 미검증 소스.**

## 1. 문서에 명시된 토큰 정책 (출처: Known Decisions 슬롯)

원문: "Auth is JWT with 15-minute expiry; refresh via POST /auth/refresh."

이를 분해하면:

| 항목 | 진술 | 확신도 |
|---|---|---|
| 토큰 형식 | JWT (bearer로 추정, 명시 안 됨) | 중 — 문서에 명시되어 있으나 slot(observation) 출처이고 코드 증거 없음 |
| 액세스 토큰 만료(TTL) | 15분 | 중 — 동일 사유. 다만 값 자체는 access token TTL로 일반적인 범위(짧은 만료)여서 그럴듯함(plausible)은 높음 |
| 갱신 방식 | `POST /auth/refresh` 엔드포인트 존재 | 중 — 엔드포인트 존재는 명시됐으나 handlers/ 하위에 이 라우트를 구현한 파일이 Frequently Changed Files에 잡히지 않음(즉 최근 변경 이력에 안 보임 → 오래된/변경 없는 코드이거나, 애초에 이 문서 스캔 범위 밖일 수 있음) |

**이 표 밖에서 문서가 말해주지 않는(=전제해서는 안 되는) 것들 — 확신도: 불명(정보 없음), 임의로 가정 금지**

- 리프레시 토큰 자체의 TTL, 회전(rotation) 정책, 1회성 여부
- 서명 알고리즘(HS256/RS256 등), 키 관리/JWKS 회전 방식
- 발급자(iss)/대상(aud) 검증 여부
- 토큰 전달 방식(Authorization 헤더 vs 쿠키), 쿠키라면 Secure/HttpOnly/SameSite 여부
- 만료 시 클럭 스큐(clock skew) 허용치
- 리프레시 토큰 폐기(revocation)/블랙리스트 메커니즘, 로그아웃 시 무효화 여부
- 리프레시 요청 자체의 인증 방식(리프레시 토큰만으로 되는지, mTLS/디바이스 바인딩이 있는지)

## 2. 문서 내부 모순으로 인해 신뢰도를 낮춰야 하는 정황 (확신도: 중 — 패턴은 명확하나 원인은 추정)

Known Decisions 슬롯의 다른 두 항목이 문서의 다른 섹션과 직접 모순된다:

- "Settlement jobs go through the legacy RabbitMQ queue" ↔ Recent Changes의 `74d4646 kafka settlement + multi-region topology` 커밋, Related Modules의 `queue/` 모듈, Frequently Changed Files의 `queue/kafka.go`. 즉 RabbitMQ가 아니라 Kafka로 이미 교체된 것으로 보이는데 Known Decisions는 갱신되지 않았다.
- "Payments runs single-region in us-east-1; DR is cold-standby only" ↔ README 요약의 "deployed multi-region (deploy/regions.yaml)", 그리고 같은 커밋의 "multi-region topology". 단일 리전이라는 결정과 멀티 리전 배포 사실이 정면으로 충돌.
- "Deprecated / Stale Notes" 섹션은 "None recorded"인데, 위 두 모순은 사실상 stale 항목임에도 별도로 표시되지 않았다.

**결론: 이 문서의 Known Decisions 슬롯은 최소 2건이 이미 사실과 어긋나는 stale 정보를 무표시로 담고 있다.** 같은 슬롯에 있는 JWT 15분 만료/리프레시 항목도 동일한 신뢰 등급(관찰/미검증)으로 취급해야 하며, 최신 코드 상태와 일치한다는 보장이 없다. 게이트웨이 미들웨어 설계에 이 값을 그대로 하드코딩 전제하는 것은 위험 — **코드 확인 전까지 "잠정치"로만 취급 권고.**

## 3. 멀티 리전 배포와 토큰 검증의 상호작용 (확신도: 낮음 — 추론)

README와 최근 커밋 근거상 실제로는 멀티 리전 배포 상태로 보이는데(Known Decisions의 "단일 리전" 항목이 stale이라는 전제하에), 만약 그렇다면 게이트웨이가 리전별로 존재할 경우 JWT 서명 키(또는 JWKS)를 리전 간 공유해야 검증 일관성이 유지된다. 이 문서는 키 공유/동기화 방식에 대해 아무 정보도 주지 않으므로, 미들웨어 설계 시 "리전 간 서명 키 동기화 여부"를 별도로 확인해야 할 미확인 전제로 남겨야 한다.

## 4. Refund 웹훅과 인증 스코프 (확신도: 낮음 — 추론, 문서에 직접 언급 없음)

"Refunds are processed asynchronously via provider webhooks, not inline" (Known Decisions)이 사실이라면, 결제 제공자(provider)가 호출하는 웹훅 엔드포인트는 사용자 JWT 인증 대상이 아닐 가능성이 높다(제공자는 사용자 토큰을 가질 수 없음 — 보통 서명 검증/HMAC 방식). 게이트웨이에 인증 미들웨어를 일괄 적용할 경우 웹훅 경로를 JWT 검사에서 제외하거나 별도 인증 방식을 적용해야 한다는 추정이 가능하지만, 이 역시 문서가 직접 확인해주지 않으므로 **전제가 아니라 확인이 필요한 가설**로 취급.

## 5. 관련은 있으나 토큰 정책과 직접 연결되지 않는 항목

- `handlers/charge.go`의 FIXME("idempotency keys not enforced")는 토큰 갱신/재시도 흐름과 개념적으로 인접할 수 있으나(예: 리프레시 후 재시도 시 중복 청구 방지), 문서는 이 둘을 연결 짓지 않는다. **확신도: 낮음 — 과도한 추론 경계.**
- Card PAN 저장 정책(vault 서비스에만 저장)은 토큰 정책과 무관.

## 요약: 게이트웨이 미들웨어 설계 전 반드시 확인(문서만으로는 확정 불가)해야 할 목록

1. 이 저장소가 실제로 "게이트웨이"인지, 아니면 게이트웨이는 별도 컴포넌트인지 (확신도: 낮음)
2. JWT 15분 만료 / `POST /auth/refresh` 정책이 현재 코드와 일치하는지 — Known Decisions 슬롯에 이미 최소 2건의 stale 항목이 확인된 만큼 검증 없이 신뢰 불가 (확신도: 중, 단 "의심해야 한다"는 결론 자체의 확신도는 높음)
3. 리프레시 토큰의 TTL/회전/폐기 정책 — 문서에 정보 전무 (확신도: 없음/불명)
4. 서명 알고리즘, 키 관리, 멀티/단일 리전 여부에 따른 키 동기화 — 문서에 정보 전무, 또한 리전 관련 진술 자체가 모순 상태 (확신도: 낮음)
5. 웹훅 경로 등 JWT 인증에서 제외되어야 할 엔드포인트 존재 여부 — 정황상 개연성은 있으나 미확인 (확신도: 낮음)

전체적으로, 이 문서가 제공하는 유일한 구체적 수치(15분 만료, `/auth/refresh` 경로)는 존재는 하지만 (a) 코드로 뒷받침되지 않고 (b) 같은 신뢰 등급의 다른 항목들이 이미 stale로 확인된 전례가 있어, "확정된 사실"이 아니라 "검증이 필요한 잠정 관찰값"으로만 전제해야 한다.
