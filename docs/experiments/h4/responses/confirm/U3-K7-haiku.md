[haiku/no-repo/confirm]

# Gateway Authentication Middleware: Token Policy Assumptions

## Confirmed Token Policy (from Known Decisions)

### 1. Token Type: JWT
- **Assertion**: Authentication uses JSON Web Tokens (JWT).
- **Confidence**: [확정] Explicitly stated in artifact context.
- **Implications for middleware**: Must perform JWT signature verification; should validate standard claims (exp, iat, etc.).

### 2. Token Expiry: 15 Minutes
- **Assertion**: Access tokens expire after 15 minutes.
- **Confidence**: [확정] Explicitly stated in artifact context.
- **Implications for middleware**: Token validation must check `exp` claim; requests with expired tokens should be rejected with 401 status.

### 3. Refresh Mechanism: POST /auth/refresh
- **Assertion**: Clients refresh tokens via HTTP POST to `/auth/refresh` endpoint.
- **Confidence**: [확정] Explicitly stated in artifact context.
- **Implications for middleware**: This endpoint must bypass normal JWT validation (or accept refresh tokens); the endpoint itself must issue new access tokens.

---

## Inferred Policies (Not Explicitly Documented)

### 4. Refresh Token Existence
- **Assertion**: A refresh token mechanism exists to support the `/auth/refresh` endpoint.
- **Confidence**: [높음] POST /auth/refresh implies a persistent credential; typical JWT patterns use separate refresh tokens.
- **Implications for middleware**: Middleware likely does not validate refresh tokens on standard endpoints; only the refresh endpoint handles refresh token logic.

### 5. Token Payload / Claims
- **Assertion**: JWT contains standard claims (iss, sub, iat, exp) and possibly service-specific claims (user_id, roles, permissions).
- **Confidence**: [중간] Common JWT practice; no explicit claim schema in artifact.
- **Implications for middleware**: Middleware should extract user identity from a standard claim (e.g., `sub`) for authorization context.

### 6. Token Storage Location
- **Assertion**: Clients store and transmit tokens via HTTP Authorization header (Bearer scheme).
- **Confidence**: [중간] Standard REST API convention; not documented in artifact.
- **Implications for middleware**: Middleware should extract token from `Authorization: Bearer <token>` header; reject requests missing this header with 401.

### 7. Token Signing Algorithm
- **Assertion**: JWT signatures use an asymmetric or symmetric cryptographic algorithm (e.g., RS256, HS256).
- **Confidence**: [낮음] No algorithm specified in artifact.
- **Implications for middleware**: Must verify algorithm before trusting signature; prevent algorithm downgrade attacks.

### 8. No Token Revocation Mentioned
- **Assertion**: Explicit token revocation (blacklist/blocklist) logic is not documented.
- **Confidence**: [낮음] Absence from artifact; refund webhooks and async processing do not imply revocation.
- **Implications for middleware**: Assume tokens are valid until expiry unless evidence of revocation logic appears; consider if logout requires revocation list.

---

## Missing Information (Requires Clarification)

### 9. Refresh Token Expiry
- **Status**: [미정] Not documented.
- **Critical for**: Determining when refresh chains must stop and re-authentication is required.

### 10. Concurrent Session Policy
- **Status**: [미정] Not documented.
- **Critical for**: Determining if multiple active access tokens per user are allowed.

### 11. CORS / Cross-Origin Token Handling
- **Status**: [미정] Not documented.
- **Critical for**: Browser-based clients and cross-region deployments (multi-region topology mentioned).

### 12. Token Scope / Permission Model
- **Status**: [미정] Not documented.
- **Critical for**: Authorization decisions within handlers (who can call /charge, /refund).

---

## Recommended Middleware Checklist

For gateway auth middleware implementation, validate in order:

1. ✓ [확정] Extract JWT from `Authorization: Bearer` header; reject if missing (401).
2. ✓ [확정] Verify JWT signature using configured key.
3. ✓ [확정] Check `exp` claim; reject if expired (401).
4. ✓ [중간] Validate `iss` (issuer) and `aud` (audience) claims if present.
5. ? [미정] Check any revocation list (if implemented).
6. ? [미정] Extract user identity (from `sub` or custom claim) for downstream context.
7. ? [미정] Enforce scope/role checks if permission model exists.

---

## Summary

**Minimal Known Policy**: JWT, 15-min access token TTL, refresh via POST /auth/refresh.

**Implementation Risk**: Refresh token lifecycle, permission model, and revocation strategy are not documented in the artifact. Clarify these with the team before finalizing middleware.
