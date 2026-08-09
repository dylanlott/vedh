# Phase 2: Guest Host Activation - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-08-08
**Phase:** 2-Guest Host Activation
**Areas discussed:** Guest lifetime & cleanup, Guest display names, Viewport target, Recovery state durability

---

## Guest lifetime & cleanup

### Q1 — How long should the guest row itself live before becoming eligible for cleanup? (OPEN-2)

| Option | Description | Selected |
|--------|-------------|----------|
| 7 days (Recommended) | Token expires at 24h, row survives a week; Saturday's host can claim on Wednesday | |
| 24 hours | Row and token expire together; one concept, minimal orphan data, same-day claiming only | |
| 7 days, extended on activity | Sliding window from last-seen; active guests never expire | |
| **Other (free text)** | **"Guest rows should live indefinitely but should be marked as guests"** | ✓ |

**User's choice:** Free-text — rows live indefinitely, flagged `is_guest`.
**Notes:** Answered outside the offered option set. Resolves OPEN-2 as *neither* of its two
stated alternatives. Orchestrator flagged two downstream consequences before proceeding:
(a) REQ-ACT-005 mandates `expires_at` plus a cleanup path for "unreferenced expired guests",
which has nothing to delete under this model, and criterion 3's "stop working when the backing
guest expires" can never fire; (b) combined with silent re-issue, the persisted session ID
becomes a permanent bearer credential. Both were turned into explicit follow-up questions (Q3, Q4)
rather than left for the planner to discover.

### Q2 — A guest's 24h token expires while their game is still live. What should happen?

| Option | Description | Selected |
|--------|-------------|----------|
| Re-issue silently while row is alive (Recommended) | Returning visitor gets a fresh token invisibly; token becomes an implementation detail | ✓ |
| Show expired state, offer claim | Token death is visible and becomes the conversion moment | |
| Read-only board, no actions | Board visible but inert until claimed; adds a third auth state | |

**User's choice:** Re-issue silently while row is alive.
**Notes:** Because Q1 made rows immortal, "while the row is alive" is always true — so the
24-hour token is never user-visible.

### Q3 — Given rows never expire, what should `expires_at` and the required cleanup path become?

| Option | Description | Selected |
|--------|-------------|----------|
| Keep column, govern re-issue only (Recommended) | `expires_at` stays enforced in authz but gates token minting, never deletion; cleanup dormant but tested | ✓ |
| Keep column, cleanup anonymizes | Cleanup strips names from dormant rows instead of deleting them | |
| Drop expiry from rows entirely | No `expires_at` at all; contradicts REQ-ACT-005 text | |

**User's choice:** Keep column, govern re-issue only.
**Notes:** Reconciles the decision with REQ-ACT-005's literal text rather than letting one
silently override the other.

### Q4 — If the session ID can mint tokens indefinitely, how should it be treated?

| Option | Description | Selected |
|--------|-------------|----------|
| Separate secret from analytics ID (Recommended) | New high-entropy credential under its own key, auth path only; `edhgo/session-id` stays analytics-only | ✓ |
| Reuse session ID, rotate on each issue | One identifier, rotated per re-issue; breaks funnel continuity | |
| Reuse session ID as-is | Simplest; accepts permanent localStorage access to a guest's board | |

**User's choice:** Separate secret from analytics ID.
**Notes:** Keeps a value that gets attached to product events out of the business of granting
sessions. The repo's existing HttpOnly-cookie JWT migration remains out of milestone scope.

---

## Guest display names

### Q1 — What shape should a generated guest name take?

| Option | Description | Selected |
|--------|-------------|----------|
| MTG-flavored adjective-noun (Recommended) | "Brave Sliver" from a curated Magic-flavored list | ✓ |
| Adjective-noun, generic word list | "Brave Otter"; neutral vocabulary, easier to vet | |
| "Guest 4821" | Unmistakable and unbounded, but reads as a system artifact | |

**User's choice:** MTG-flavored adjective-noun.
**Notes:** Curation is part of the work — no offensive or absurd pairings. Fixed list bounds
the namespace, so overflow strategy is required.

### Q2 — What must a generated guest name be unique against?

| Option | Description | Selected |
|--------|-------------|----------|
| Globally unique across all users (Recommended) | Same namespace as usernames; never collides with either | ✓ |
| Unique per game only | Smaller burden, but claiming resolves conflicts at the worst moment | |
| Separate namespace from usernames | Clean separation, but relocates the conflict to claim time | |

**User's choice:** Globally unique across all users.
**Notes:** Interacts with the immortal-row decision — names are never recycled, so the
namespace only grows.

### Q3 — When a visitor types their OWN display name, must it be unique too?

| Option | Description | Selected |
|--------|-------------|----------|
| Free-form, not unique (Recommended) | Two people can both be "Dylan"; no taken-name friction on the activation path | ✓ |
| Unique, reject on conflict | Username-like semantics; injects a failure state into the flow | |
| Free-form, auto-disambiguate | Silent "(2)" suffix on collision within a game | |

**User's choice:** Free-form, not unique.

### Q4 — Where does the free-form typed display name live, given `users` has only a UNIQUE `username`?

*Asked after the scout found that `users` is `username`/`password`/`uuid`/`timestamp` with a
`username_unique` constraint, and that `InputCreateGame.Handle` is the game's handle, not a
user's — so Q2 and Q3 could not both be satisfied by the existing schema.*

| Option | Description | Selected |
|--------|-------------|----------|
| New non-unique `display_name` column (Recommended) | `username` stays unique and holds the generated name; `display_name` holds what was typed | ✓ |
| Typed name overwrites username | One column, but forces uniqueness and contradicts Q3 | |
| `display_name` for everyone, incl. real accounts | Wider blast radius now, less special-casing later | |

**User's choice:** New non-unique `display_name` column.
**Notes:** Requires every `Username` render site to become `display_name ?? username`.

### Q5 — At claim time, what should the new permanent username default to?

| Option | Description | Selected |
|--------|-------------|----------|
| Freely chosen, prefilled with display name (Recommended) | Prefill from typed name, else generated name; editable | ✓ |
| Keep the generated guest name | Never conflicts, but few people want "Brave Sliver" permanently | |
| Freely chosen, empty field | Normal signup semantics; discards a name they identified with | |

**User's choice:** Freely chosen, prefilled with display name.
**Notes:** Only the `claimGuestAccount` backend contract is in scope this phase.

---

## Viewport target

*Scoping note raised before asking: this phase's UI surface is the activation path only —
criterion 1 ends at `/games/:id`, so `BoardView.vue` (96K) is out of scope.*

### Q1 — What viewport range must the activation path support for the quiet beta? (OPEN-3)

| Option | Description | Selected |
|--------|-------------|----------|
| Desktop + tablet (≥768px) (Recommended) | One breakpoint on new activation components; covers the iPad-at-the-table case | ✓ |
| Desktop only (≥1024px) | Matches the board's reality; defers the question to Phase 4 | |
| Fluid, no fixed breakpoint | Intrinsically fluid layout; harder to verify | |

**User's choice:** Desktop + tablet (≥768px).
**Notes:** Resolves OPEN-3. The codebase has zero `@media` queries, so this establishes the
app's first responsive pattern and ACT-011 in Phase 4 inherits it.

### Q2 — Someone opens `/play` on a phone. What should happen?

| Option | Description | Selected |
|--------|-------------|----------|
| Works, with a heads-up about the board (Recommended) | Import functions; visible notice sets expectations before a long paste | ✓ |
| Explicit unsupported message | Honest but a hard bounce on a real visitor | |
| No special handling | Zero work; undefined behavior on the measured surface | |

**User's choice:** Works, with a heads-up about the board.

---

## Recovery state durability

*Context raised before asking: `Warnings`/`BlockingErrors` are flat `[String!]!` with no codes,
and `stores/games.ts:87` shows the existing pattern is a hardcoded generic string per call site.*

### Q1 — Where should the in-progress activation state live?

| Option | Description | Selected |
|--------|-------------|----------|
| sessionStorage (Recommended) | Survives refresh and failed navigation; dies with the tab | ✓ |
| Component state only | Literally satisfies criterion 4, but a refresh loses a long paste | |
| localStorage | Survives browser restart; decklist persists indefinitely on device | |

**User's choice:** sessionStorage.
**Notes:** Must be explicitly cleared on successful game creation.

### Q2 — How should the four failure classes become messages with the right way forward?

| Option | Description | Selected |
|--------|-------------|----------|
| Server error codes, client maps to copy (Recommended) | Stable machine-readable code lets the client pick the affordance per failure | ✓ |
| Client-side mapping per call site | Extends the existing pattern; brittle as causes multiply | |
| Single generic recoverable-error state | Cheapest; gives the wrong advice when the provider is down | |

**User's choice:** Server error codes, client maps to copy.
**Notes:** Vocabulary should be modelled on Phase 1's closed, server-owned event vocabulary.

---

## Claude's Discretion

Surfaced as candidate areas at the closing gate; user chose to proceed to context rather than
discuss them. Recorded in CONTEXT.md with their consequences:

- **Guest-creation kill switch default** — REQ-ACT-005 requires a flag. Phase 1's provider
  switch defaults off; if this one does too, the phase ships dark.
- **`/play`'s relationship to existing flows** — whether `LandingView.vue`'s CTA changes and
  whether `FormCreateGame.vue` is replaced outright or runs in parallel during rollout.
- **Commander review UX for partners/backgrounds** — whether the new component consumes,
  extends, or supersedes the existing `commanderPartner.ts`.

## Deferred Ideas

- **Claim-account UI** — backend ships this phase; the form, prompt timing, and conversion
  nudge do not.
- **Board responsiveness** — activation path only this phase; `BoardView.vue` belongs with
  ACT-011 in Phase 4.
- **Guest name rename** — no rename feature exists for anyone after claiming. Not needed for
  activation.
