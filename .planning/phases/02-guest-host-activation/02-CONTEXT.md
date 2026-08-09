# Phase 2: Guest Host Activation - Context

**Gathered:** 2026-08-08
**Status:** Ready for planning

<domain>
## Phase Boundary

A Commander host who has never made an account pastes a deck, reviews what was parsed,
and lands on their own live board — and an existing authenticated host travels the same
road without a guest identity being invented for them.

Three requirements: **REQ-ACT-004** (reusable deck import + commander review UI),
**REQ-ACT-005** (guest identity and account-claim backend), **REQ-ACT-006** (public
quick-start host flow at `/play`).

**This phase ends at `/games/:id`.** Criterion 1 stops at arrival on the board.
`BoardView.vue` behavior, board responsiveness, and the invite/join path (ACT-007/008,
Phase 3) are all outside this boundary. REQ-A5 advances at the **backend** level only —
`claimGuestAccount` ships, the claim *UI* does not.

</domain>

<decisions>
## Implementation Decisions

### Guest lifetime and identity durability

- **D-2.1:** Guest rows live **indefinitely**, distinguished by the `is_guest` flag rather
  than reaped on a timer. This resolves **OPEN-2** as *neither* of its stated options —
  the question asked "24 hours or seven days"; the answer is that the backing row never
  expires at all. The 24-hour token from REQ-ACT-005 is unchanged.
  — **Reversibility:** one-way — once real guest identities accumulate with attached game
  history, retrofitting an expiry window means deciding whose identity to destroy. There is
  no migration back from "we kept everyone" to "we keep people for a week."

- **D-2.2:** `expires_at` **stays on the schema and stays enforced in authorization**, but
  it governs *only* whether a session credential may mint a **new** token. It never triggers
  row deletion. The independently-retryable cleanup path REQ-ACT-005 requires is still built
  and still tested, but targets only rows explicitly marked for removal — under D-2.1 it has
  no scheduled work to do.
  — **Reversibility:** reversible — the column and the job both exist; only the predicate
  that feeds them would change.

- **D-2.3:** A guest whose 24-hour token expires gets a **silent re-issue** while their row
  is alive. Because D-2.1 makes the row always alive, the token becomes an implementation
  detail rather than a user-facing deadline. Nobody is shown an "expired session" wall.

- **D-2.4:** Silent re-issue makes the re-auth value a **bearer credential**, so it must be
  a **separate secret from the analytics session ID**. Phase 1's `edhgo/session-id` stays
  exactly what it is — a low-stakes attribution identifier that gets attached to product
  events. A *distinct* high-entropy guest credential is minted, stored under its own
  localStorage key, and sent only to the auth path.
  — **Reversibility:** costly — merging or splitting these later means every already-issued
  guest device holds the wrong shape of credential, and the funnel keyed on the analytics ID
  breaks if the two are conflated after the fact.

### Guest and display names

- **D-2.5:** Generated guest names are **MTG-flavored adjective-noun** pairs (e.g. "Brave
  Sliver") from a **curated** word list. Curation is part of the work: no pairing may be
  offensive or absurd. A fixed list bounds the namespace, so a numeric suffix is the
  expected overflow strategy.

- **D-2.6:** Generated names are **globally unique across all users** — the same namespace
  as real usernames, colliding with neither. Requires retry-on-conflict at generation. Note
  this interacts directly with D-2.1: because rows are immortal, names are **never recycled**
  and the namespace only ever grows.

- **D-2.7:** A visitor-typed display name is **free-form and NOT unique**. Two people may
  both be "Dylan"; the row ID is the identity and the name is a label. No "that name is
  taken" failure may appear on the activation path — the phase's premise is never being
  redirected through `/login` or `/signup`.

- **D-2.8:** D-2.6 and D-2.7 cannot both live in today's `username` column, which is `UNIQUE`.
  Therefore: add a **new nullable, non-unique `display_name` column**. `username` remains the
  unique system identifier and holds the generated MTG name for guests. **Every surface
  currently rendering `Username` must switch to `display_name ?? username`** — this is a wide
  sweep (`BoardView.vue`, `ScoreView.vue`, `GamesView.vue`, `AppNav.vue` at minimum) and is
  easy to under-do.
  — **Reversibility:** one-way — adding the column needs a migration on both prod and test
  schemas, and once the view layer reads `display_name ?? username` everywhere, collapsing
  back to a single column means choosing which name to destroy for every guest who typed one.

- **D-2.9:** At claim time the permanent username is **freely chosen, prefilled** from
  whatever the user has been called — typed display name if present, generated name
  otherwise. The prefill may collide, so the claim form carries the ordinary taken-name path.
  Only the `claimGuestAccount` backend contract is in scope this phase; the form itself is later.

### Viewport and responsive baseline

- **D-2.10:** The activation path supports **desktop and tablet, ≥768px** — one breakpoint,
  applied to the **new activation components only**. This resolves **OPEN-3**. Because the
  codebase contains **zero `@media` queries today**, this establishes the app's first
  responsive pattern; ACT-011 in Phase 4 will inherit whatever shape it takes, so it is worth
  getting right rather than fast.

- **D-2.11:** `/play` on a phone **works, with a heads-up**. The import flow itself functions
  — pasting a decklist is not layout-hungry — but a visible notice sets expectations about the
  board before the visitor invests a 100-card paste. Consistent with criterion 4's
  always-a-visible-way-forward ethos. No hard bounce.

### Failure recovery

- **D-2.12:** In-progress activation state — the paste, corrected entries, commander choices,
  and display name — lives in **`sessionStorage`**. It survives a refresh and a failed
  navigation (the failure modes criterion 4 actually names) and dies with the tab, so a long
  decklist is not left on a shared machine. **Must be explicitly cleared on successful game
  creation** or stale state resurfaces on the next visit.

- **D-2.13:** The four recoverable failure classes (provider down, preview error,
  guest-session error, create error) carry a **stable machine-readable error code from the
  server**, which the client maps to copy *and to the correct affordance* — retry for a
  provider timeout, "fix these cards" for unresolved entries, "paste instead" when the
  provider is down. This is the only approach that satisfies "never shows raw GraphQL, SQL,
  or provider output" while giving each failure a distinct exit. Model the code vocabulary on
  Phase 1's closed event vocabulary: a fixed, server-owned, allowlisted set.
  — **Reversibility:** costly — the codes become a client-consumed contract; widening is easy,
  but renaming or removing one breaks a deployed client mid-funnel.

### Claude's Discretion

Three in-scope gray areas were surfaced and deliberately not discussed. Planner's call, but
each carries a noted consequence:

- **Guest-creation kill switch default.** REQ-ACT-005 mandates a feature flag and kill switch.
  Phase 1's provider switch (`DECK_PROVIDER_ENABLED`) defaults **off**. If guest creation also
  defaults off, this phase ships dark and "done" means something different than it appears —
  the activation funnel it exists to open stays closed until someone flips a flag. Decide this
  explicitly and state it in the plan; do not let it default by copy-paste from Phase 1.

- **`/play`'s relationship to the existing flows.** REQ-ACT-004 says "replace duplicated deck
  entry in `FormCreateGame.vue` and `JoinGameView.vue`" while keeping "the authenticated create
  flow usable during rollout." How aggressive that replacement is — whether `LandingView.vue`'s
  call-to-action changes, whether `FormCreateGame` is replaced outright or runs in parallel —
  was not settled. Both readings satisfy the requirement text.

- **Commander review UX for partners and backgrounds.** `app/src/services/commanderPartner.ts`
  already exists and encodes partner logic. Whether the new commander-review component consumes
  it as-is, extends it, or supersedes it is open. Phase 1 decided the *product surface* of
  candidates (up to 3 suggestions, always show nearest, computed eagerly) but said nothing about
  multi-commander selection.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone specifications

- `.planning/ROADMAP.md` § "Phase 2: Guest Host Activation" — the five success criteria,
  the two-wave ticket order (ACT-004 ∥ ACT-005 → ACT-006), and the `UI hint: yes` marker
- `.planning/REQUIREMENTS.md` — REQ-ACT-004 / 005 / 006 full text with dependency order;
  REQ-A1, A2, A5, A7 for the product-level framing this phase advances
- `.planning/PROJECT.md` — core value, out-of-scope list, and the resolved OPEN-1 entry

### Prior phase decisions this phase inherits

- `.planning/phases/01-measured-deck-import-foundation/01-CONTEXT.md` — the unresolved-card
  model (warnings / blocking-errors / `CanContinue`), D-07 printing-metadata persistence,
  the source-type taxonomy, and the closed product-event vocabulary. **Do not re-litigate
  these.**
- `.planning/phases/01-measured-deck-import-foundation/01-VERIFICATION.md` — what Phase 1
  actually proved, including the accounting invariant this phase's UI renders
- `docs/analytics/product-event-vocabulary.md` — the closed 15-event vocabulary. Criterion 5
  names exactly which events are server-written (`game_created`, `guest_session_created`)
  and which are client-emitted (`quick_start_viewed`, `deck_import_started`,
  `game_create_started`).

### Adjacent work — read for boundaries, do not implement

- `docs/plans/2026-05-15-vedh-auth-storage-migration.md` — the HttpOnly-cookie JWT migration
  is explicitly out of scope for this milestone. D-2.4's new guest credential lands in
  localStorage alongside the existing token and inherits the same known limitation; do not
  attempt the cookie migration here, but do not make it harder either.

### Code the guest work must respect

- `app/src/stores/auth.ts` — the `{ID, Username, Token}` profile persisted under
  `edhgo/auth`, with no notion of expiry or guest-ness. D-2.4 and D-2.8 both land here.
- `persistence/migrations/20210307154621_init_db.up.sql` and
  `persistence/migrations/20211229130157_enforce_unique_usernames.up.sql` — the actual
  `users` shape and the `username_unique` constraint that forces D-2.8.
- `app/src/services/productEvents.ts` — Phase 1's session-ID persistence and event emitter,
  which D-2.4 deliberately leaves alone.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- **`app/src/services/productEvents.ts`** — Phase 1's client event service with persisted
  session ID and allowlisted campaign attribution. Criterion 5's three client-emitted events
  go through it. Its retry semantics matter: criterion 5 requires a retried call not to
  double-count a conversion.
- **`app/src/services/commanderPartner.ts`** — existing partner/background logic the new
  commander-review component should consume rather than reimplement (see Claude's Discretion).
- **`app/src/stores/games.ts`** — already wraps `createGame` / `joinGame`; the quick-start
  flow calls the *existing* `createGame` rather than a new mutation.
- **`server/deck_import.go`** — `previewDeck` and the `DeckPreview` type (`Warnings`,
  `CanContinue`, `BlockingErrors`) that `DeckImportPanel` renders.

### Established Patterns

- **Every route except `/login` and `/signup` is `requiresAuth: true`** (`app/src/router/index.ts`).
  `/play` will be the first genuinely public *functional* route, and the `beforeEach` guard
  needs a public-but-stateful case that does not exist yet.
- **Error handling is a hardcoded generic string per call site** — `stores/games.ts:87` sets
  `errorMessage.value = 'Unable to load games'` in a blanket catch. D-2.13 deliberately
  departs from this pattern; expect to introduce the code-mapping layer rather than extend
  the existing one.
- **`Warnings` / `BlockingErrors` are flat `[String!]!` with no codes** (`server/schema.graphql:80-82`).
  Parser-level messages are already user-safe strings; D-2.13's codes are additive and are
  about *transport and flow* failures, not about re-typing the parser output.
- **No responsive layer exists.** Zero `@media` queries across `app/src`. D-2.10 creates the
  first one.

### Integration Points

- **The duplicated deck textarea.** `app/src/components/games/FormCreateGame.vue:76` and
  `app/src/views/JoinGameView.vue:31` contain a **literally duplicated** block labelled
  "Decklist (CSV: quantity,name per line)" with an identical placeholder — and both still
  describe the CSV-only format Phase 1 replaced. These are the two call sites REQ-ACT-004
  replaces, and both are currently *lying to the user* about what syntaxes work.
- **`Handle` is the GAME's handle, not a user's.** `InputCreateGame.Handle`
  (`server/schema.graphql:275`) labels the game. REQ-ACT-006's "submit `Handle` and `FormatID`
  correctly" is about game labelling and is trivially misread as a user handle — which would
  be a real bug given this phase is otherwise full of user-naming work.
- **`users` has no display-name column.** Table is `username` / `password` / `uuid` /
  `timestamp` only. D-2.8's migration touches both prod and test schemas, as REQ-ACT-005
  already requires for `is_guest` and `expires_at`.

</code_context>

<specifics>
## Specific Ideas

- **OPEN-2 was answered outside its own option set.** The roadmap framed it as "24 hours or
  seven days." The actual decision — rows live forever, flagged as guests — is neither, and
  it changes REQ-ACT-005's cleanup requirement from an operational necessity into a dormant
  safety valve. The requirement text and the decision were deliberately reconciled (D-2.2)
  rather than letting one silently win. A planner should not read REQ-ACT-005's "cleanup path
  for unreferenced expired guests" as evidence that D-2.1 is wrong.

- **Criterion 3's "stop working when the backing guest expires" is now vacuous by design.**
  Under D-2.1 the backing guest never expires, so that clause can never fire. The
  authorization check still gets built and tested (D-2.2 keeps `expires_at` enforced) — but
  verification should not expect to observe the expired-guest rejection in a live scenario,
  because no scenario produces one. Test it directly against a hand-marked row.

- **The immortal-row decision makes name collision a permanent problem, not a transient one.**
  A 7-day window would have let the namespace recycle. It cannot now. Whatever curated word
  list D-2.5 produces needs a deliberate overflow strategy from day one, not when it first
  collides in production.

- **The two textareas are the clearest evidence the phase is needed.** Both promise
  "CSV: quantity,name per line" — the exact format Phase 1's six-syntax parser superseded.
  Anyone using the app today is being told to use a syntax that is no longer the canonical one.

</specifics>

<deferred>
## Deferred Ideas

- **Claim-account UI.** `claimGuestAccount` ships as backend this phase (REQ-A5 "backend"),
  but the form, the prompt timing, and the "you're about to lose this game" nudge are not in
  scope. D-2.9 records the intended prefill behavior so whoever builds the UI inherits the
  decision rather than re-deciding it.
- **Board responsiveness.** D-2.10 covers the activation path only. Making `BoardView.vue`
  (96K) work at ≥768px is its own body of work and belongs with ACT-011 in Phase 4.
- **Guest name rename.** D-2.9 lets a claimer choose a fresh username, but there is no rename
  feature for anyone afterward. Not needed for activation; worth noting it does not exist.

</deferred>

---

*Phase: 2-Guest Host Activation*
*Context gathered: 2026-08-08*
