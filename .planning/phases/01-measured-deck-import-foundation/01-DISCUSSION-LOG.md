# Phase 1: Measured Deck Import Foundation - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-08-03
**Phase:** 1-measured-deck-import-foundation
**Areas discussed:** Unresolved-card policy, Deck export noise handling, Provider spike framing (OPEN-1), Event durability + session identity, Fuzzy-match threshold, Event + metadata allowlists, sourceFormat detection

---

## Unresolved-card policy

### Can the player still create/join with unmatched cards?

| Option | Description | Selected |
|--------|-------------|----------|
| Proceed with visible warning | Preview flags entries, CanContinue true, player starts anyway | ✓ |
| Block until corrected | CanContinue false until every entry fixed or deleted | |
| Threshold — proceed under a limit | Warn and allow when few, block past a cutoff | |

**User's choice:** Proceed with visible warning
**Notes:** Prioritizes the ≥60% activation target over the ≥98% resolution target where they conflict.

### What should a near-miss offer?

| Option | Description | Selected |
|--------|-------------|----------|
| Offer ranked suggestions to pick | 2-3 candidates the player explicitly accepts | ✓ |
| Flag unresolved, no suggestions | Mark unresolved, player retypes | |
| You decide | Claude's discretion | |

**User's choice:** Offer ranked suggestions to pick
**Notes:** Constrained by the standing rule that fuzzy matching may never silently rewrite a deck — hence explicit acceptance. Requires new similarity work; `s.Cards()` is exact-match SQL today.

### Do unresolved entries count toward the deck-size rule?

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — count them | An unresolved row still occupies a slot | |
| No — exclude until resolved | Only matched cards count toward the cap | ✓ |
| You decide | Claude's discretion | |

**User's choice:** No — exclude until resolved

### Follow-up: what lands in the library for unresolved entries?

| Option | Description | Selected |
|--------|-------------|----------|
| Omit from the library | 100-card paste with 3 bad names → 97-card library | ✓ |
| Include as visible placeholders | Deck stays 100 with marked unknown cards | |
| Include, and count them after all | Reverses the size-rule answer for one consistent number | |

**User's choice:** Omit from the library
**Notes:** Asked because proceed-with-warning plus exclude-from-count left the library contents ambiguous. Omitting keeps preview counts and board contents in agreement.

---

## Deck export noise handling

### Set codes, collector numbers, category tags

| Option | Description | Selected |
|--------|-------------|----------|
| Strip and ignore them | Parse name, discard the rest | |
| Strip, but warn once | Same parse plus a single notice | |
| Don't support — treat as unresolved | Only the six named syntaxes parse | |
| *(free text)* | Parse and store the metadata by source | ✓ |

**User's choice:** Free-text — "Let's parse and store the additional metadata by source (i.e. Moxfield, Archidekt, etc...)"
**Notes:** Follow-up established the intent is **persistence**, for display of exact printings and a planned future custom-proxy-print feature. Claude initially warned this would need a migration outside the SPEC's schema constraints and would touch the board-state model; that warning was **wrong and retracted** — `upsertGame` stores games as a JSONB payload, so new `Card` fields persist for free. Also surfaced: the `cards` table already carries `setcode`/`number`/`scryfallid`, and the current name-keyed lookup already picks a printing arbitrarily.

### Sideboard / Maybeboard sections

| Option | Description | Selected |
|--------|-------------|----------|
| Drop with a visible count | Excluded, player told how many | ✓ |
| Drop silently | Excluded with no mention | |
| Import into the library | Treat all cards as deck content | |

**User's choice:** Drop with a visible count

### Should a `Commander` header seed candidates?

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — preselect, player confirms | Header cards become preselected candidates | ✓ |
| Yes — but only as a suggestion | Highlighted but not preselected | |
| No — detect from card data only | Ignore the header entirely | |

**User's choice:** Yes — preselect, player confirms

### Printing absent from the MTGJSON snapshot

| Option | Description | Selected |
|--------|-------------|----------|
| Match name, warn about printing | Keep the card, warn art may differ | ✓ |
| Match name, fall back silently | Resolve by name, no mention | |
| Treat as unresolved | Requested printing unavailable → entry drops | |

**User's choice:** Match name, warn about printing
**Notes:** Rejecting silent fallback keeps consistency with the never-silently-rewrite rule; rejecting unresolved avoids a stale snapshot shrinking decks when new sets appear.

---

## Provider spike framing (OPEN-1)

### Seed the spike with a preference?

| Option | Description | Selected |
|--------|-------------|----------|
| Start neutral — let the spike decide | Both evaluated against the same criteria | ✓ |
| Try Archidekt first | Fall back to Moxfield | |
| Try Moxfield first | Fall back to Archidekt | |

**User's choice:** Start neutral — let the spike decide
**Notes:** OPEN-1 stays genuinely open; planning favours no candidate.

### How should the plan handle the two outcomes?

| Option | Description | Selected |
|--------|-------------|----------|
| Spike task + decision checkpoint | Halt at checkpoint:decision before adapter code | ✓ |
| Plan both branches up front | Adapter and no-go tasks both written, gated | |
| Assume success, replan if it fails | Adapter path only | |

**User's choice:** Spike task + decision checkpoint

### Hard gate conditions beyond the locked SSRF controls (multi-select)

| Option | Description | Selected |
|--------|-------------|----------|
| No API key or account needed | Works with no credentials of any kind | |
| Terms of service permit it | Documented ToS check | |
| Stable response shape | Versioned or stable enough that a parser won't silently rot | ✓ |
| Reasonable rate limits | Limits that survive real activation traffic | |

**User's choice:** Stable response shape only
**Notes:** Three conditions deliberately left as non-blocking. They may still be recorded in the decision record for the human approving at the checkpoint, but none independently fails the gate.

### App-level credentials?

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — app-level credentials allowed | vEDH holds its own key/account | ✓ |
| No — fully anonymous access only | No credentials at all | |
| Prefer anonymous, allow key if needed | Anonymous target, key as fallback | |

**User's choice:** Yes — app-level credentials allowed
**Notes:** Asked specifically because leaving "no API key" unselected could have been an oversight. It was not — the distinction from *user* credentials (still forbidden by PROJECT.md) was intentional. Implies deploy-time secret management.

---

## Event durability + session identity

### Should a failed event write fail the user's action?

| Option | Description | Selected |
|--------|-------------|----------|
| Never — event failure is swallowed | Logged and counted, action always succeeds | ✓ |
| Same transaction — both or neither | Event insert rides game creation's transaction | |
| Best-effort with durable retry | In-process queue with retry | |

**User's choice:** Never — event failure is swallowed
**Notes:** Accepted cost is under-counted funnel denominators during an incident. The Prometheus drop counter is the agreed mitigation and is therefore required, not optional.

### Where should the session ID live?

| Option | Description | Selected |
|--------|-------------|----------|
| localStorage — per browser | Survives tabs, reloads, restarts | ✓ |
| sessionStorage — per tab | New ID per tab | |
| Cookie | Server-visible on every request | |

**User's choice:** localStorage — per browser
**Notes:** Cookie was disfavoured partly because it would pre-empt the out-of-scope HttpOnly auth migration.

### Client event delivery

| Option | Description | Selected |
|--------|-------------|----------|
| Fire-and-forget per event | One mutation each, failures ignored | ✓ |
| Batched with flush | Queue, flush on interval and page-hide | |
| You decide | Claude's discretion | |

**User's choice:** Fire-and-forget per event
**Notes:** Batching would blur the very timestamps that define the time-to-board metrics.

---

## Fuzzy-match threshold

### How many suggestions per unresolved entry?

| Option | Description | Selected |
|--------|-------------|----------|
| Up to 3 | Covers realistic near-misses, scans in one glance | ✓ |
| Single best guess | Fastest, fails on ambiguous stems | |
| Up to 5 | Widest net, denser preview | |

**User's choice:** Up to 3

### Nothing clears the cutoff — what does the player see?

| Option | Description | Selected |
|--------|-------------|----------|
| Unresolved, no suggestions | Offer nothing rather than garbage | |
| Show nearest anyway, marked low confidence | Always show something, visually flagged | ✓ |
| You decide | Claude's discretion | |

**User's choice:** Show nearest anyway, marked low confidence

### When are suggestions computed?

| Option | Description | Selected |
|--------|-------------|----------|
| Eagerly, in the preview response | One round trip, instant correction | ✓ |
| Lazily, on player request | Fast common path, extra round trip per correction | |
| You decide | Claude's discretion | |

**User's choice:** Eagerly, in the preview response
**Notes:** Claude flagged (without re-asking) that eager plus always-show-nearest defines the worst-case cost path — a fully-unresolved 100-card paste runs similarity on every row inside the preview request. Recorded in CONTEXT.md as an implementation risk for the planner to bound.

---

## Event + metadata allowlists

### Is the 15-event vocabulary closed?

| Option | Description | Selected |
|--------|-------------|----------|
| Closed — exactly these 15 | Hardcoded constant, anything else rejected | ✓ |
| Closed, but config-extensible | Extensible via server config | |
| Open with prefix convention | Accept anything matching a pattern | |

**User's choice:** Closed — exactly these 15
**Notes:** The PRD already specifies all 15 events and their metadata, so this is transcription rather than invention.

### How should metadata keys be allowlisted?

| Option | Description | Selected |
|--------|-------------|----------|
| Per-event key lists | Each event declares exactly which keys it may carry | ✓ |
| Global key union | One list across all events | |

**User's choice:** Per-event key lists

### What happens to a rejected event?

| Option | Description | Selected |
|--------|-------------|----------|
| Drop, log, and count | Prometheus rejection counter | ✓ |
| Return a GraphQL error | Explicit failure the client could react to | |
| Drop silently | No record | |

**User's choice:** Drop, log, and count
**Notes:** Consistent with fire-and-forget delivery — no client is listening for an error, so a counter is the only surface that would be read.

---

## sourceFormat detection

### One mechanism or two, across Phase 1 paste and Phase 2 URL detection?

| Option | Description | Selected |
|--------|-------------|----------|
| One shared enum, two detectors | Single enum, separate paste-shape and URL-host detectors | ✓ |
| Fully unified detector | One function taking text or URL | |
| Keep them separate | Independent enums per phase | |

**User's choice:** One shared enum, two detectors
**Notes:** Prevents Phase 2 inventing a parallel vocabulary that would split funnel data and inflate Prometheus label cardinality.

### Enum granularity

| Option | Description | Selected |
|--------|-------------|----------|
| Provider-level | moxfield, archidekt, plain_text, unknown | ✓ |
| Format-level | csv_quoted, csv_bare, mtgo_style, etc. | |
| Provider-level, with format in the payload | Enum coarse, finer detail non-indexed | |

**User's choice:** Provider-level

---

## Claude's Discretion

- **`previewDeck` rate-limiting scope** — offered as a fourth gray area in the second round and not selected for discussion. The SPEC already mandates rate limiting on guest creation, deck import, and public invite lookup; only whether `previewDeck` gets its own limit was open.
- **Similarity algorithm, cutoff value, and scoring function** for the ranked suggestions — only the product surface was decided.
- **Paste-shape detection heuristics** distinguishing moxfield / archidekt / plain_text, and what an ambiguous paste resolves to.

## Deferred Ideas

- **Custom proxy prints for specific cards** — surfaced as the motivation for persisting printing metadata. A new user-facing capability belonging to a later phase; Phase 1 only captures the printing identity that would otherwise be lost at import time.
