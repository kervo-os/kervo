[sonnet/no-repo/confirm]

# 게이트웨이 인증 미들웨어 도입 전 전제해야 할 토큰 정책(만료·갱신)

> 근거 문서: `artifact-K7.md` (자동 생성 컨텍스트 아티팩트, repo=h4-repo). 저장소 코드는 열람하지 않았으며, 아래 모든 진술은 이 문서 한 건에서만 도출했다. 문서 자체가 "Fact sections are deterministic; slot sections carry trust-labeled observations"라고 명시하고 있어, 섹션 종류에 따라 신뢰 수준이 다르다는 점을 전제로 확신도를 매겼다.

## 0. 먼저 짚어야 할 메타 신뢰도 문제

- 이 문서의 인증 관련 진술은 전부 **"Known Decisions" 슬롯**(`slot:decisions`)에서 나온다. 문서 서두는 Fact 섹션(Repository Summary, Commands, Recent Changes 등)과 Slot 섹션(Known Decisions 포함)을 명확히 구분하며, Slot은 "trust-labeled observations"라고 정의한다. 즉 저장소를 실제로 읽고 확정한 사실이 아니라 자동 추출된 관찰일 가능성이 있다.
  **확신도: 높음** (이는 문서 스키마 자체의 명시적 진술이므로).
- "Possible Current Goal" 슬롯은 "A confirmed goal becomes the first Verified observation"라고 되어 있어, 이 문서 스키마에는 관찰 신뢰도에 등급(예: 미확정 vs Verified)이 존재함을 암시한다. 그런데 Known Decisions 섹션의 각 항목에는 개별 신뢰 등급 표시가 없다. 따라서 "JWT 15분 만료" 항목이 "Verified" 수준인지 단순 추출 수준인지는 **이 문서만으로는 판단 불가**.
  **확신도: 판단불가** (문서에 등급 라벨이 붙어있지 않음).
- 같은 문서의 "Deprecated / Stale Notes" 섹션에는 "Settlement jobs go through the legacy RabbitMQ queue"라는 문구가 있는데, README 요약은 "Settlement jobs flow through Kafka"라고 말한다. 즉 이 문서 안에서도 슬롯 간 정보 불일치(스테일 정보 혼입)가 이미 한 건 확인된다. 이는 인증 슬롯에 대해서도 유사한 스테일 가능성을 배제할 수 없다는 정황 증거다.
  **확신도: 중간** (직접적 반증은 아니고 유사 사례로부터의 유추).

## 1. 문서에서 직접 진술된 항목 (Known Decisions 슬롯 근거)

| 항목 | 진술 | 확신도 | 비고 |
|---|---|---|---|
| 인증 방식 | JWT 기반 인증 | 중간 | Slot 관찰이며 개별 등급 미표시 |
| Access token 만료 | 15분 | 중간 | 숫자까지 구체적으로 명시되어 다소 신뢰도를 높게 볼 수 있으나, 이 역시 Slot 관찰 |
| 갱신 경로 | `POST /auth/refresh` | 중간 | 엔드포인트 경로가 명시적이라 구현 시 그대로 참조 가능하나, 실재 여부는 코드 확인 전까지 가정 |

게이트웨이 미들웨어를 붙일 때 이 세 가지는 "전제"로 삼을 수 있는 최소 정보다: (a) Authorization 헤더 등에서 JWT를 파싱·검증하는 로직이 필요하고, (b) 액세스 토큰의 유효기간은 15분이라는 짧은 주기이므로 미들웨어는 만료 검증을 엄격히(clock 기준) 수행해야 하며, (c) 만료 시 클라이언트가 별도 `/auth/refresh` 엔드포인트로 갱신한다는 흐름을 게이트웨이가 막지 않아야 한다(즉 `/auth/refresh` 자체는 만료된 access token으로도 호출 가능해야 하는지, 아니면 별도의 refresh token으로 호출하는지는 아래 2절 참조).

## 2. 문서에 없어 반드시 "미확정 가정"으로 남겨야 하는 항목

아래는 게이트웨이 인증 미들웨어 설계에 실무적으로 필수적이지만, 문서에 어떤 형태로도 언급되지 않은 항목이다. 실제 구현 전 저장소 코드(예상 위치: `handlers/`, 혹은 별도 `auth/`·`middleware/` 모듈이 있다면 그곳) 확인이 필요하다.

- **Refresh token 자체의 만료/회전 정책**: 문서는 access token의 15분 만료만 언급하고, refresh token의 수명, 1회성 여부(rotation), 재사용 탐지(reuse detection) 여부는 전혀 언급하지 않는다.
  **확신도: 판단불가** (문서에 정보 없음 — 추측 금지).
- **서명 알고리즘 및 키 관리**: HS256(대칭키) vs RS256/ES256(비대칭키) 여부, 키 로테이션 정책 불명.
  **확신도: 판단불가**.
- **멀티리전 환경에서의 키/시크릿 공유 여부**: README 발췌에 "deployed multi-region (deploy/regions.yaml)"이 명시되어 있다. 여러 리전에 게이트웨이가 분산 배포된다면, 한 리전에서 발급한 토큰을 다른 리전 게이트웨이가 검증할 수 있어야 하는데, 서명 키를 리전 간 공유하는지, 리전별 로컬 검증 키가 있는지는 문서에 전혀 없다. 이는 멀티리전 배포가 명시된 상태에서 인증 미들웨어를 설계할 때 실제로 막힐 수 있는 지점이라 별도로 강조한다.
  **확신도: 낮음(리스크 추정)** — 멀티리전 배포 사실 자체는 문서 근거가 있으나(README 발췌), 그로부터 "토큰 검증에 리전 간 이슈가 있을 것"이라는 결론은 이 문서에 직접 서술되어 있지 않은 추론이다.
- **로그아웃/토큰 폐기(revocation, 블랙리스트) 메커니즘 존재 여부**: 언급 없음.
  **확신도: 판단불가**.
- **`/auth/refresh` 자체의 보호 방식**: 이 엔드포인트가 게이트웨이 인증 미들웨어의 적용 대상(즉 미들웨어를 통과해야 하는지, 우회해야 하는지)인지 불명. 일반적으로 refresh 엔드포인트는 access-token 검증 미들웨어에서 제외(bypass)해야 하는 경우가 많지만, 이 문서는 그런 예외 규칙을 언급하지 않는다.
  **확신도: 판단불가** — 실무 관행에 근거한 추정일 뿐, 문서 근거 없음.
- **Rate limiting / brute-force 방지**가 `/auth/refresh`에 적용되는지: 언급 없음.
  **확신도: 판단불가**.
- **issuer/audience 클레임 검증, clock skew 허용치**: 언급 없음.
  **확신도: 판단불가**.
- **인증 미들웨어 자체가 이미 코드베이스에 존재하는지 여부**: "Related Modules" 목록은 `deploy/`, `handlers/`, `queue/`, `store/` 네 개뿐이고 `auth/`나 `middleware/`류 모듈은 나열되지 않는다. "Frequently Changed Files" 목록에도 인증 관련 파일이 없다. 따라서 "게이트웨이에 인증 미들웨어를 붙인다"는 과제 전제 자체가 신규 구현(from scratch)인지, 기존 미들웨어 확장인지 이 문서만으로는 알 수 없다.
  **확신도: 낮음** (모듈 목록에 없다는 것이 곧 "존재하지 않는다"의 증명은 아님 — 아티팩트가 파일 변경 빈도 기반 요약이라 정적/미변경 파일은 누락될 수 있음).

## 3. 인접 리스크로부터의 낮은 확신도 추론 (참고용, 설계 확정 근거로 사용 금지)

- Open Tasks에 `handlers/charge.go:2 — FIXME: idempotency keys not enforced`가 있다. 결제 요청에 멱등성 키가 강제되지 않는다는 것은, 만약 인증 갱신(refresh) 중 클라이언트가 요청을 재시도하는 구간에서 중복 charge가 발생할 잠재 경로가 있을 수 있음을 시사한다. 다만 이는 인증 정책과 직접 연결된 서술이 문서에 없으므로 추론일 뿐이다.
  **확신도: 낮음** (인증-멱등성 연결은 이 문서에 명시되어 있지 않은 조합 추론).
- `main.go:2 — TODO: wire settlement cron`, `store/db.go:2 — TODO: migrate to sqlc`는 인증 미들웨어와 직접 관련이 없어 보이며 본 과제 범위에서 제외.
  **확신도: 높음(무관하다는 판단에 대해)**.

## 4. 요약: 게이트웨이 인증 미들웨어 설계 시 최소 전제

1. Access token은 JWT, 만료 15분 — 확신도 중간(Slot 관찰, 등급 미표시).
2. 만료 시 갱신 경로는 `POST /auth/refresh` — 확신도 중간.
3. Refresh token 수명, 회전, 폐기, 서명키 종류/로테이션, 멀티리전 키 공유, refresh 엔드포인트의 미들웨어 예외 여부, rate limiting, issuer/audience 검증 — **전부 문서 미기재, 판단불가**. 실제 구현 전 반드시 저장소 코드(추정 위치: `handlers/`, 존재한다면 `auth/`·`middleware/`)를 직접 확인해야 한다.
4. 이 문서의 Slot 섹션은 자동 추출된 관찰이며, 문서 내 다른 슬롯(Deprecated/Stale Notes)에서 이미 한 건의 정보 불일치(Kafka vs RabbitMQ)가 확인되었으므로, 인증 슬롯 내용도 구현 전 코드 대조 없이 그대로 신뢰하지 않을 것을 권장한다 — 확신도 중간의 권고사항.
