# Security Best Practices Report

## Executive Summary
This review found **6 security issues** in the current Go + Vue stack: **2 High**, **3 Medium**, and **1 Low**. The highest-risk items are authorization gaps that allow any authenticated user to read/subscribe to other games, and missing HTTP server timeouts that increase DoS exposure. The frontend also stores bearer tokens in `localStorage`, increasing account-takeover risk if any XSS occurs.

## High Severity

### [SBP-001] Broken object-level authorization on game reads and subscriptions
- Rule ID: `APP-AUTHZ-001`
- Severity: High
- Location: `server/games.go:37`, `server/games.go:68`, `server/games.go:92`, `server/authz.go:41`, `server/schema.graphql:74`, `server/schema.graphql:135`
- Evidence:
  - `Games` and `GetGame` only require authenticated context, not game membership.
  - `GameUpdated` checks only that `userID` matches caller identity, but does not check whether the caller is a participant in the target `gameID`.
  - Returned `Game` contains player state, including board/hand fields in schema.
- Impact: Any logged-in user can access or subscribe to game state for games they are not part of, leaking private gameplay state.
- Fix: Enforce `isUserInGame(game, authUser)` on `Games`, `GetGame`, and `GameUpdated` (or create an explicit visibility policy like public-vs-private games and enforce it consistently).
- Mitigation: If an immediate change is risky, add feature-flagged authz checks plus audit logging for cross-user game access.
- False positive notes: If all games are intentionally public by product design, document that policy explicitly and remove sensitive fields from public responses.

### [SBP-002] HTTP server is started without explicit timeout and header-size limits
- Rule ID: `GO-HTTP-001`
- Severity: High
- Location: `server/graphql.go:190`
- Evidence:
  - Server starts via `http.ListenAndServe(fmt.Sprintf(":%d", port), h)` with default zero-value timeouts.
- Impact: Increased susceptibility to slowloris and connection/resource exhaustion attacks.
- Fix: Replace `ListenAndServe` call with an explicit `http.Server` that sets `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, and `MaxHeaderBytes`.
- Mitigation: Edge/load-balancer connection limits can reduce risk but do not replace app-level timeout controls.
- False positive notes: If an upstream proxy already has strict limits, risk is reduced but still present at app listener level.

## Medium Severity

### [SBP-003] CORS and WebSocket origin policy are overly permissive
- Rule ID: `GO-HTTP-007`
- Severity: Medium
- Location: `server/graphql.go:167`, `server/graphql.go:184`
- Evidence:
  - WebSocket `CheckOrigin` always returns `true`.
  - HTTP middleware uses `cors.AllowAll()`.
- Impact: Broad cross-origin API exposure increases abuse surface and weakens browser-origin trust boundaries.
- Fix: Use explicit origin allowlists from config for both HTTP CORS and WS origin checks; restrict allowed headers/methods.
- Mitigation: Keep bearer-only auth and short token TTL to reduce blast radius.
- False positive notes: This is less severe than cookie-auth CORS misconfigurations, but still a hardening gap.

### [SBP-004] Metrics endpoint is exposed on the main public listener
- Rule ID: `GO-DEPLOY-002`
- Severity: Medium
- Location: `server/graphql.go:188`
- Evidence:
  - `/prometheus` is registered on the same listener as app traffic with no route-level auth.
- Impact: Unauthenticated metrics can leak operational details useful for recon and attack tuning.
- Fix: Move metrics to an internal-only listener or gate with auth/network ACL.
- Mitigation: At minimum, block `/prometheus` at ingress except from trusted monitoring IPs.
- False positive notes: If ingress/firewall already blocks public access, verify and document that runtime control.

### [SBP-005] Auth bearer token is persisted in `localStorage`
- Rule ID: `VUE-AUTH-001`
- Severity: Medium
- Location: `app/src/stores/auth.ts:16`, `app/src/stores/auth.ts:37`, `app/src/services/apollo.ts:50`, `app/src/services/apollo.ts:95`
- Evidence:
  - Auth profile (including `Token`) is read/written from `localStorage` and attached to `Authorization` headers.
- Impact: Any successful XSS can exfiltrate long-lived tokens and enable account takeover.
- Fix: Prefer backend-issued `HttpOnly` session cookies; if bearer tokens are kept, use short-lived access tokens in memory only and rotate frequently.
- Mitigation: Deploy strict CSP and strong XSS defenses while migrating auth storage strategy.
- False positive notes: Risk depends on XSS exposure; this finding does not assert an existing XSS.

## Low Severity

### [SBP-006] Debug logs include full game payload and deserialized game object
- Rule ID: `GO-CONFIG-001`
- Severity: Low
- Location: `server/games.go:81`, `server/games.go:87`
- Evidence:
  - Debug logs print raw game payload and full deserialized game object.
- Impact: Sensitive game state can leak to logs and downstream log systems.
- Fix: Remove payload/object logs or redact sensitive fields before logging.
- Mitigation: Restrict debug logging in production and tighten log retention/access controls.
- False positive notes: Exposure is limited if debug logging is always disabled in production.

## Notes
- This review focused on application code in this repository. Infrastructure controls (WAF, ingress auth, private networking, edge header injection) were not visible here and should be verified separately.
