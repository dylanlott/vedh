# Phase 2: Guest Host Activation - Pattern Map

**Mapped:** 2026-08-08
**Files analyzed:** 21 (new + modified)
**Analogs found:** 18 / 21

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `app/src/components/decks/DeckImportPanel.vue` | component | request-response | `app/src/components/games/FormCreateGame.vue` | role-match (extracted subset) |
| `app/src/components/decks/CommanderReview.vue` | component | request-response | `app/src/components/games/FormCreateGame.vue` (typeahead block) | role-match |
| `app/src/views/QuickStartView.vue` | component (view) | request-response + event-driven | `app/src/views/JoinGameView.vue` | role-match |
| `app/src/components/games/FormCreateGame.vue` | component | request-response | itself (existing) | exact — modified in place |
| `app/src/views/JoinGameView.vue` | component (view) | request-response | itself (existing) | exact — modified in place |
| `app/src/stores/auth.ts` | store | CRUD (client-side profile) | itself (existing) | exact — extended in place |
| `app/src/router/index.ts` | route/config | request-response (nav guard) | itself (existing) | exact — extended in place |
| `app/src/graphql/mutations.ts` | config (GraphQL documents) | request-response | itself (existing, `LOGIN_MUTATION`/`SIGNUP_MUTATION`) | exact |
| `app/src/graphql/queries.ts` | config (GraphQL documents) | request-response | itself (existing) | exact |
| `app/src/styles/breakpoints.scss` | config | n/a | none — first file of its kind | no analog |
| `server/guest_users.go` | controller (GraphQL resolver) | request-response + CRUD | `server/users.go` (`Signup`/`Login`) | exact |
| `server/users.go` (`Login` modification) | controller | request-response | itself (existing) | exact — modified in place |
| `server/authz.go` (possible expiry helper) | middleware | request-response | `server/authz.go` (`requireAuth`) | role-match |
| `server/graphql.go` (`Conf` fields) | config | n/a | `DeckProviderEnabled` field block | exact |
| `pkg/ratelimit/limiter.go` (`SurfaceGuestSession`) | config | n/a | `SurfaceDeckImport`/`SurfaceProductEvent` consts | exact |
| `server/ratelimit.go` (call sites, unchanged mechanism) | middleware | request-response | `allowRequest`/`clientKeyFor` (existing) | exact — reused, not modified |
| `persistence/migrations/{TS}_guest_users.up/down.sql` | migration | batch | `persistence/migrations/20260804120000_product_events.up.sql` | exact |
| `persistence/migrations_test/{TS}_guest_users.up/down.sql` | migration | batch | matching pair pattern (`init_db`/`enforce_unique_usernames`) | exact |
| `server/guest_users_test.go` | test | request-response | `server/deck_import_test.go` (`TestMigrations_CardNameSearch`, resolver tests) | role-match |
| `app/__tests__/QuickStartView.spec.ts` | test | request-response | `app/__tests__/JoinGame.integration.spec.ts` | role-match |
| `app/__tests__/DeckImportPanel.spec.ts` | test | request-response | `app/__tests__/FormCreateGame.integration.spec.ts` | role-match |
| Display-name sweep (`BoardView.vue`, `ScoreView.vue`, `GamesView.vue`, `AppNav.vue`, `GameAnalysisView.vue`) | component (template edit) | transform (render) | each file's own existing `{{ }}` Username interpolation sites | exact — edited in place |

## Pattern Assignments

### `server/guest_users.go` (controller, request-response + CRUD)

**Analog:** `server/users.go` (`Signup`, `Login`, `hashPassword`, `checkPasswordHash`)

**Imports pattern** (`server/users.go:1-13`):
```go
package server

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
	"golang.org/x/crypto/bcrypt"
)
```
`guest_users.go` additionally needs `github.com/openmtg/edh-go/pkg/ratelimit` (for `SurfaceGuestSession`) and `crypto/rand` for the D-2.4 credential — never `math/rand` (see `server/backronym.go`'s use of `math/rand`, explicitly flagged in RESEARCH.md as *not* a security pattern to copy).

**Row insert + collision retry pattern** (`server/users.go:23-35`):
```go
id := uuid.New().String()
stmt := `
INSERT INTO "users" (uuid, username, password)
VALUES ($1, $2, $3)
RETURNING uuid, username;
`
result, err := s.db.Query(stmt, id, username, hashed)
if err != nil {
	if strings.Contains(err.Error(), "username_unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate key value") {
		return nil, errs.New("That username is already taken. Try another one.")
	}
	return nil, errs.Wrap(err)
}
```
For `GuestSession`, on this same error class, generate a new adjective-noun pair (or numeric-suffix variant per D-2.5) and retry the INSERT, bounded by a max-attempts constant — never surface a "taken" error to the visitor (D-2.7).

**Token issuance pattern** (`server/users.go:44,88` + `server/auth.go:110-129`):
```go
t, err := newAuthToken(user, time.Hour*24)
if err != nil {
	return nil, errs.Wrap(err)
}
user.Token = &t
```
Call unmodified for a guest: `newAuthToken(guestUser, time.Hour*24)`.

**Password hashing pattern** (`server/users.go:104-107`):
```go
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
```
Reuse verbatim for the guest's random, never-revealed password.

**`Login` is_guest rejection point** (`server/users.go:53-102`, specifically the query at line 62):
```go
q := `SELECT "uuid", "username", "password" FROM "users" WHERE username=$1;`
```
Add `is_guest` to the SELECT list and `Scan` target; after the existing `checkPasswordHash` check, add a rejection when `is_guest` is true, mirroring the existing error style: `return nil, errs.New("failed to authenticate")` (do not leak "this is a guest account" — same generic message keeps behavior indistinguishable from a wrong password, which is itself defense-in-depth).

**Error handling style:** every error path in `users.go` wraps with `errs.Wrap(err)` or returns a hand-written `errs.New(...)` — no custom error-code type exists yet in Go. D-2.13's error codes are a **new** concept for this phase; model them as small sentinel errors (e.g. `var errGuestCreationDisabled = errs.New("guest creation is currently disabled")`) that the client maps to a code, or attach a machine-readable code via GraphQL extensions — RESEARCH.md's Code Examples section shows the intended resolver shape:
```go
// server/guest_users.go (new, modeled on server/product_events.go:151-159's
// TrackProductEvent call shape and server/deck_providers.go:337-339's providerEnabled())
func (s *graphQLServer) GuestSession(ctx context.Context, displayName *string, sessionID string) (*User, error) {
	if !s.guestCreationEnabled() {
		return nil, errGuestCreationDisabled
	}
	clientKey := clientKeyFor(ctx, sessionID)
	if !s.allowRequest(ctx, ratelimit.SurfaceGuestSession, clientKey) {
		return nil, errRateLimited
	}
	// ... generate name, hash random password, insert row, issue token ...
	s.recordProductEvent(ctx, ProductEvent{
		Name:      "guest_session_created",
		SessionID: sessionID,
		UserID:    &guestUser.ID,
	}, false)
	return guestUser, nil
}
```

---

### `server/graphql.go` (config, kill switch + rate-limit fields)

**Analog:** `DeckProviderEnabled` field + `providerEnabled()` predicate

**Kill-switch field pattern** (`server/graphql.go:63-69`):
```go
// DeckProviderEnabled is the provider kill switch (D-16/T-01-28).
// Defaults to false: a deploy must be safe before this is
// deliberately turned on...
DeckProviderEnabled bool `envconfig:"DECK_PROVIDER_ENABLED" default:"false"`
```

**Consuming predicate pattern** (`server/deck_providers.go:337-339`):
```go
func (s *graphQLServer) providerEnabled() bool {
	return s != nil && s.cfg.DeckProviderEnabled && len(s.deckProviderAllowedHosts) > 0
}
```
Add `GuestCreationEnabled bool `envconfig:"GUEST_CREATION_ENABLED" default:"false"`` and a `guestCreationEnabled()` predicate on `*graphQLServer`. **Kill-switch default is Claude's Discretion** — CONTEXT.md flags this explicitly; do not silently inherit `false` by copy-paste without recording the decision in the plan.

Also mirror `DeckImportRatePerMinute`/`DeckImportRateBurst` (`server/graphql.go:53-54`) with `GuestSessionRatePerMinute`/`GuestSessionRateBurst` feeding the **same** `s.limiter` Registry (do not construct a second `Registry` — see Anti-Patterns below).

---

### `pkg/ratelimit/limiter.go` (config, new Surface constant)

**Analog:** existing `Surface` const block (`pkg/ratelimit/limiter.go:32-37`)
```go
const (
	// SurfaceDeckImport identifies the previewDeck mutation.
	SurfaceDeckImport Surface = "deck_import"
	// SurfaceProductEvent identifies the trackProductEvent mutation.
	SurfaceProductEvent Surface = "product_event"
)
```
Add `SurfaceGuestSession Surface = "guest_session"` to this same block — the package's own doc comment (lines 24-29) explicitly anticipates this addition.

---

### `server/ratelimit.go` (reused verbatim — no modification needed)

**Analog / itself:**
```go
// allowRequest — server/ratelimit.go:34-46
func (s *graphQLServer) allowRequest(ctx context.Context, surface ratelimit.Surface, clientKey string) bool {
	allowed := true
	if s != nil && s.limiter != nil {
		allowed = s.limiter.Allow(surface, clientKey)
	}
	outcome := rateLimitOutcomeAllowed
	if !allowed {
		outcome = rateLimitOutcomeLimited
	}
	collectors.ObserveRateLimit(surface, outcome)
	return allowed
}

// clientKeyFor — server/ratelimit.go:71-81
func clientKeyFor(ctx context.Context, sessionID string) string {
	if addr, ok := remoteAddrFromContext(ctx); ok {
		if trimmed := strings.TrimSpace(addr); trimmed != "" {
			return "addr:" + trimmed
		}
	}
	if trimmed := strings.TrimSpace(sessionID); trimmed != "" {
		return "session:" + trimmed
	}
	return "unknown"
}
```
Both functions are called with `ratelimit.SurfaceGuestSession` as the new surface parameter — **no changes to either function's body.**

---

### `persistence/migrations/{TS}_guest_users.up.sql` / `.down.sql` (migration, batch)

**Analog:** matched-pair pattern, e.g. `persistence/migrations/20260804120000_product_events.up.sql` and its `migrations_test/` sibling — byte-identical content, same filename, two directories.

**Existing `users` shape** (`persistence/migrations/20210307154621_init_db.up.sql:1-6`):
```sql
CREATE TABLE IF NOT EXISTS users (
  username VARCHAR(255),
  password VARCHAR(255),
  uuid VARCHAR(255) UNIQUE,
  timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```
Plus `persistence/migrations/20211229130157_enforce_unique_usernames.up.sql:1`:
```sql
ALTER TABLE users ADD CONSTRAINT username_unique UNIQUE (username);
```

**New migration shape (both directories, identical content):**
```sql
ALTER TABLE users ADD COLUMN is_guest BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN expires_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN display_name VARCHAR(255);
ALTER TABLE users ADD COLUMN guest_credential_hash VARCHAR(255);
```
`.down.sql` drops the four added columns by name; do not touch `username_unique`.

---

### `app/src/router/index.ts` (route/config, request-response nav guard)

**Analog:** itself — existing route array + `beforeEach` (`app/src/router/index.ts:4-88`)

**Existing public-route pattern** (lines 10-21):
```typescript
{
  path: '/login',
  name: 'login',
  component: () => import('../views/LoginView.vue'),
  meta: { public: true },
},
```

**Existing guard** (lines 79-88):
```typescript
router.beforeEach((to, from, next) => {
  if (to.meta.public) {
    return next();
  }
  const auth = useAuthStore();
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return next({ name: 'login', query: { redirect: to.fullPath } });
  }
  return next();
});
```
**New `/play` entry** (per RESEARCH.md's Code Examples, correct per Pitfall 3 — `meta.public: true` is right at the router level; the guest/authenticated branch happens *inside* `QuickStartView.vue`, not in the guard):
```typescript
{
  path: '/play',
  name: 'quick-start',
  component: () => import('../views/QuickStartView.vue'),
  meta: { public: true },
},
```
Do **not** invent a third meta flag (e.g. `publicStateful`) unless a concrete need surfaces — `meta.public` plus in-component `auth.isAuthenticated` branching already covers criterion 1.

---

### `app/src/stores/auth.ts` (store, CRUD — client profile)

**Analog:** itself — existing `AuthProfile` shape and `login`/`signup` actions

**Current profile shape and persistence** (lines 7-38):
```typescript
interface AuthProfile {
  ID: string;
  Username: string;
  Token: string;
}
const STORAGE_KEY = 'edhgo/auth';
function loadPersistedProfile(): AuthProfile | null {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as AuthProfile;
  } catch (error) {
    console.warn('[auth] failed to parse profile:', error);
    localStorage.removeItem(STORAGE_KEY);
    return null;
  }
}
// ...
watch(profile, (value) => {
  if (!value) { localStorage.removeItem(STORAGE_KEY); return; }
  localStorage.setItem(STORAGE_KEY, JSON.stringify(value));
}, { deep: true });
```
**Action-with-error-handling shape** (lines 42-69, `login`):
```typescript
async function login(credentials: { username: string; password: string; redirect?: string }) {
  loading.value = true;
  errorMessage.value = null;
  try {
    const { data } = await apolloClient.mutate<LoginMutation, LoginMutationVariables>({
      mutation: LOGIN_MUTATION,
      variables: { username: credentials.username, password: credentials.password },
    });
    if (!data?.login) throw new Error('Login returned empty response');
    profile.value = { ID: data.login.ID, Username: data.login.Username, Token: data.login.Token };
    return profile.value;
  } catch (error: unknown) {
    console.error('[auth] login failed', error);
    errorMessage.value = error instanceof Error ? error.message : 'Login failed';
    throw error;
  } finally {
    loading.value = false;
  }
}
```
**For this phase:** extend `AuthProfile` with `IsGuest?: boolean` (or omit — see RESEARCH.md Open Question 2, default to *not* trusting a stale JWT claim) and add a `createGuestSession(displayName, sessionID)` action following this exact try/catch/finally shape, calling the new `GUEST_SESSION_MUTATION`. D-2.4's new bearer credential is a **separate** value — store it under a new key (e.g. `edhgo/guest-credential`), **not** inside `AuthProfile`/`edhgo/auth`, and never attach it as `Authorization` header for ordinary calls (Pitfall 4).

**Session-ID storage precedent for the new credential key** (`app/src/services/productEvents.ts:59-71`, defensive storage accessor — model the new credential's storage helper on this, but using its own key and never conflating with `edhgo/session-id`):
```typescript
function getStorage(): Storage | null {
  try {
    if (typeof localStorage !== 'undefined' && localStorage) return localStorage;
  } catch { /* ignore */ }
  try {
    if (typeof window !== 'undefined' && window.localStorage) return window.localStorage;
  } catch { /* ignore */ }
  return null;
}
```

---

### `app/src/views/QuickStartView.vue` (component/view, request-response + event-driven)

**Analog:** `app/src/views/JoinGameView.vue` (existing view wrapping a deck textarea + submit)

**Current duplicated decklist block to replace** (`JoinGameView.vue:31-32`, identical in `FormCreateGame.vue:76-77`):
```html
<span>Decklist (CSV: quantity,name per line)</span>
<textarea v-model="decklist" rows="6" placeholder="1, Atraxa, Pr…\n99, Basic Island"></textarea>
```
This exact block — and its stale CSV-only placeholder copy — is what `DeckImportPanel.vue` replaces at all three call sites.

**Existing submission payload shape** (`JoinGameView.vue:248-270`):
```typescript
// Decklist
// ...
Decklist: decklist.value,
// ...
User: auth.profile.Username,
```
`QuickStartView.vue` follows the same "collect local component state, submit as a normalized payload" shape but calls the **existing, unmodified** `createGame` from `app/src/stores/games.ts` (not a new mutation), per REQ-ACT-006 and RESEARCH.md's Architectural Responsibility Map.

**sessionStorage draft persistence — no existing precedent** (`grep -rn "sessionStorage" app/src` returns zero matches). Model on `productEvents.ts`'s defensive-accessor style shown above, but for `sessionStorage`:
```typescript
const DRAFT_KEY = 'edhgo/quickstart-draft';
function saveDraft(draft: QuickStartDraft): void {
  try { sessionStorage.setItem(DRAFT_KEY, JSON.stringify(draft)); } catch { /* ignore */ }
}
function clearDraft(): void {
  try { sessionStorage.removeItem(DRAFT_KEY); } catch { /* ignore */ }
}
// clearDraft() MUST be called on successful createGame() — D-2.12.
```

---

### `app/src/components/decks/DeckImportPanel.vue` / `CommanderReview.vue` (component, request-response)

**Analog:** `app/src/components/games/FormCreateGame.vue` (typeahead block, lines 39-45, 366-382) — reuse the `.typeahead` CSS class and list-with-active-item interaction pattern for the manual card-search fallback and commander candidate list.

**Server contract these components render** (`server/deck_import.go:186, 299-433` — `PreviewDeck` / `DeckPreview`):
```go
func (s *graphQLServer) PreviewDeck(ctx context.Context, input InputDeckImport) (*DeckPreview, error)
// ...
return &DeckPreview{
	Warnings:            warnings,
	CanContinue:         anyResolved, // an unmatched card never blocks the player
	BlockingErrors:      blocking,
	// Entries, CommanderCandidates, Unresolved also present
}, nil
```
`Warnings`/`BlockingErrors` are flat `[String!]!` with no codes today (`server/schema.graphql:80-82`) — these are already user-safe parser strings; D-2.13's new codes are additive and cover *transport/flow* failures (provider down, guest-session error, create error), not re-typing this output.

**Commander partner logic to consume, not reimplement:** `app/src/services/commanderPartner.ts` exports `canAddSecondCommander`/`isValidPartnerPair`/`partnerConstraintMessage` — `CommanderReview.vue` should call these directly rather than re-encoding partner/background rules (Don't Hand-Roll).

---

### Display-name sweep (`BoardView.vue`, `ScoreView.vue`, `GamesView.vue`, `AppNav.vue`, `GameAnalysisView.vue`, `FormCreateGame.vue`, `JoinGameView.vue`)

**Pattern:** every current display site uses default `{{ }}` Vue interpolation (never `v-html` — confirmed zero matches for `v-html.*[Uu]sername|display` across `app/src/**/*.vue`). Change is template-level only: `{{ player.Username }}` → `{{ player.DisplayName ?? player.Username }}` (or store-level computed if repeated). **Identity-comparison sites** (e.g. `BoardView.vue`'s `selfPlayer`/`isMe` joins, `GamesView.vue:106,114`) must keep comparing on `Username`, never `DisplayName` — only *display* sites get the fallback.

**The GraphQL-document half of this sweep (equally required, easy to under-do):** `app/src/graphql/queries.ts` (`GAMES_QUERY`, `GET_GAME_QUERY`, `GAME_UPDATED_SUBSCRIPTION`) and `app/src/graphql/mutations.ts` (`CREATE_GAME_MUTATION`, `JOIN_GAME_MUTATION`) must each add `DisplayName` to their `Players { ... }` selection set — a template fix with no matching query-document fix always resolves to `username` silently (Pitfall 1). Follow the existing `gql` tag pattern already used for `LOGIN_MUTATION`/`SIGNUP_MUTATION`.

---

### `app/src/styles/breakpoints.scss` (config, new — no precedent)

Per `02-UI-SPEC.md`'s Responsive Architecture section (locked, checker-verified):
```scss
// app/src/styles/breakpoints.scss (NEW)
$breakpoint-tablet: 768px;
```
Consumed via `@use '@/styles/breakpoints' as bp;` in each new component's `<style scoped lang="scss">` block, mobile-first (`min-width` ascending), exactly one query per component, scoped only to `QuickStartView.vue`, `DeckImportPanel.vue`, `CommanderReview.vue`, and the mobile heads-up notice. Three pre-existing ad hoc `@media (max-width: …)` queries exist (`LandingView.vue:702`, `BoardView.vue:2646`, `GamesView.vue:319`) — explicitly **not** to be extended or reconciled with this new pattern.

## Shared Patterns

### Server-authoritative event write with dedup
**Source:** `server/product_events.go:151-188` (`TrackProductEvent`, client-facing sibling) and the dedup index at `persistence/migrations/20260804120000_product_events.up.sql:30-38`
**Apply to:** `GuestSession` (writes `guest_session_created`) and any `createGame` call path touched by the guest flow (writes `game_created`) — both already members of the four-event dedup set; no new migration needed for dedup itself.
```go
s.recordProductEvent(ctx, ProductEvent{
	Name:      "guest_session_created",
	SessionID: input.SessionID,
	UserID:    &guestUser.ID,
	Outcome:   &outcome,
}, false) // false = server-authoritative
```

### Kill-switch + rate-limit gating
**Source:** `server/deck_providers.go:337-339` (`providerEnabled`), `server/ratelimit.go:34-46` (`allowRequest`)
**Apply to:** `GuestSession` resolver — check kill switch first, then rate limit, before any DB write.

### Client event emission
**Source:** `app/src/services/productEvents.ts:212-235` (`track()`)
**Apply to:** `quick_start_viewed` (on `/play` mount), `deck_import_started` (first paste/URL submit), `game_create_started` (just before `createGame` call) — all client-emitted per criterion 5, all go through the existing `track()` function unmodified.

### bcrypt/JWT reuse (never hand-roll)
**Source:** `server/auth.go:110-129` (`newAuthToken`), `server/users.go:104-107` (`hashPassword`)
**Apply to:** Guest row creation and any silent-reissue mechanism (see Open Questions in RESEARCH.md — no reissue mutation exists yet; plan must decide whether to build one this phase).

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `app/src/styles/breakpoints.scss` | config | n/a | First `@media`-driven, deliberately-architected responsive file in the codebase; the three existing ad hoc `@media` sites are explicitly not a pattern to extend. |
| sessionStorage draft persistence (`QuickStartView.vue`) | utility | event-driven | `grep -rn "sessionStorage" app/src` returns zero matches — no existing accessor to copy beyond the general defensive-storage shape in `productEvents.ts` (which uses `localStorage`, not `sessionStorage`). |
| D-2.4 guest re-auth credential storage/verification | service | request-response | No existing "second bearer credential" concept anywhere in `server/auth.go` or `app/src/stores/auth.ts`; RESEARCH.md's Open Question 1 flags that even the *mutation* this would call is undecided. |
| Curated MTG adjective-noun word list (D-2.5) | data/config | transform | No existing word-list or name-generation file in the repo; `server/backronym.go` is explicitly flagged as unrelated and not a pattern to follow. |

## Metadata

**Analog search scope:** `app/src/{router,stores,views,components,graphql,services,styles}`, `server/*.go`, `pkg/ratelimit/*.go`, `persistence/migrations*/`
**Files scanned:** ~25 (direct reads) plus grep sweeps across `app/src` and `server`
**Pattern extraction date:** 2026-08-08
